package worker

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	// Added for simulating delays
	"github.com/godbus/dbus/v5"
	"github.com/grandcat/zeroconf"
	"github.com/holoplot/go-avahi"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes" // Added for gRPC status codes
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status" // Added for gRPC status package

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport" // Import for transport service
)

// workerServer implements pb_discovery.WorkerDiscoveryServiceServer and pb_control.RenderControlServiceServer.
// It DOES NOT implement TransportServiceServer, as the worker will be a client for scene/texture/triangle streaming.
type workerServer struct {
	pb_discovery.UnimplementedWorkerDiscoveryServiceServer
	pb_control.UnimplementedRenderControlServiceServer

	workerID         string
	availableCores   uint32
	totalMemoryBytes uint64
	freeMemoryBytes  uint64
	currentStatus    pb_discovery.WorkerStatus

	// Stored assets for mocking purposes (in a real app, these would be loaded into memory or GPU)
	loadedScene     *pb_transport.Scene
	loadedTextures  map[string][]byte
	loadedTriangles []*pb_transport.Triangle
}

// NewWorkerServer creates and returns a new workerServer instance.
func NewWorkerServer(numCores uint32) *workerServer {
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("Failed to get hostname for worker ID: %v", err)
	}
	workerID := hostname

	totalMem := memory.TotalMemory()
	freeMem := memory.FreeMemory()

	return &workerServer{
		workerID:         workerID,
		availableCores:   numCores,
		totalMemoryBytes: totalMem,
		freeMemoryBytes:  freeMem,
		currentStatus:    pb_discovery.WorkerStatus_FREE,
		loadedTextures:   make(map[string][]byte), // Initialize the map
	}
}

// --- WorkerDiscoveryServiceServer Implementations ---

func (s *workerServer) QueryWorkerStatus(ctx context.Context, req *pb_discovery.QueryWorkerStatusRequest) (*pb_discovery.QueryWorkerStatusResponse, error) {
	logrus.Printf("WorkerDiscoveryService: QueryWorkerStatus called on worker %s", s.workerID)
	return &pb_discovery.QueryWorkerStatusResponse{
		NodeName:         s.workerID,
		AvailableCores:   s.availableCores,
		TotalMemoryBytes: s.totalMemoryBytes,
		FreeMemoryBytes:  s.freeMemoryBytes,
		Status:           s.currentStatus,
	}, nil
}

// --- RenderControlServiceServer Implementations ---

// sendStatus sends a RenderSetupResponse with the given status and an optional error message.
func (s *workerServer) sendStatus(stream pb_control.RenderControlService_RenderSetupServer, cfgStatus pb_control.RenderSetupStatus, errMsg string) error {
	resp := &pb_control.RenderSetupResponse{
		Status:       cfgStatus,
		ErrorMessage: errMsg,
	}
	if err := stream.Send(resp); err != nil {
		logrus.Errorf("Failed to send RenderSetupResponse status %s (error: %s): %v", cfgStatus.String(), errMsg, err)
		return err
	}
	logrus.Infof("RenderSetup: Sent status: %s %s", cfgStatus.String(), errMsg)
	return nil
}

// getScene fetches the Scene proto message from the asset provider.
func (s *workerServer) getScene(ctx context.Context, transportClient pb_transport.SceneTransportServiceClient, sceneName string) (*pb_transport.Scene, error) {
	req := &pb_transport.GetSceneRequest{SceneName: sceneName}
	scene, err := transportClient.GetScene(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene '%s': %w", sceneName, err)
	}
	return scene, nil
}

// streamTextureFile streams a texture file from the asset provider.
// `expectedTotalSize` is used for pre-allocation, obtained from ImageTexture.size.
func (s *workerServer) streamTextureFile(ctx context.Context, transportClient pb_transport.SceneTransportServiceClient, filename string, expectedTotalSize uint64) ([]byte, error) {
	req := &pb_transport.StreamTextureFileRequest{
		Filename:  filename,
		Offset:    0,         // Start from beginning
		ChunkSize: 64 * 1024, // Consistent chunk size
	}
	stream, err := transportClient.StreamTextureFile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to open texture stream for %s: %w", filename, err)
	}

	textureData := make([]byte, 0, expectedTotalSize) // Pre-allocate based on ImageTexture.size
	receivedBytes := uint64(0)

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == context.Canceled {
				return nil, fmt.Errorf("texture stream cancelled for %s", filename)
			}
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.NotFound {
					return nil, fmt.Errorf("texture '%s' not found on provider", filename)
				}
				if s.Code() == codes.Unavailable {
					logrus.Warnf("Texture stream for %s closed by server gracefully (Unavailable). Received %d of %d bytes.", filename, receivedBytes, expectedTotalSize)
					break // Server closed stream, assume end.
				}
			}
			if err.Error() == "EOF" { // gRPC stream end
				logrus.Infof("Finished streaming texture '%s'. Received %d of %d bytes.", filename, receivedBytes, expectedTotalSize)
				break
			}
			return nil, fmt.Errorf("failed to receive texture chunk for %s: %w", filename, err)
		}

		// resp.GetSize() here is the size of the *current chunk*, not the total size.
		// Use len(resp.GetData()) instead if chunk size is meant for individual chunk size
		textureData = append(textureData, resp.GetChunk()...)
		receivedBytes += uint64(len(resp.GetChunk()))
	}

	if expectedTotalSize > 0 && receivedBytes != expectedTotalSize {
		return nil, fmt.Errorf("texture '%s' stream ended prematurely. Expected %d bytes, got %d", filename, expectedTotalSize, receivedBytes)
	}

	return textureData, nil
}

// streamTriangles streams triangles from the asset provider.
func (s *workerServer) streamTriangles(ctx context.Context, transportClient pb_transport.SceneTransportServiceClient, sceneName string, totalTriangles uint64, batchSize uint32) ([]*pb_transport.Triangle, error) {
	allTriangles := make([]*pb_transport.Triangle, 0, totalTriangles) // Pre-allocate for all triangles
	var fetchedCount uint64 = 0

	// Loop to fetch all batches
	for fetchedCount < totalTriangles {
		req := &pb_transport.StreamTrianglesRequest{
			SceneName: sceneName,
			BatchSize: batchSize,
			Offset:    fetchedCount, // Request from the current offset
		}
		stream, err := transportClient.StreamTriangles(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to open triangles stream for scene '%s': %w", sceneName, err)
		}

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == context.Canceled {
					return nil, fmt.Errorf("triangles stream cancelled for scene '%s'", sceneName)
				}
				if s, ok := status.FromError(err); ok {
					if s.Code() == codes.NotFound {
						return nil, fmt.Errorf("triangles for scene '%s' not found on provider", sceneName)
					}
					if s.Code() == codes.Unavailable {
						logrus.Warnf("Triangles stream for scene '%s' closed by server gracefully (Unavailable). Received %d of %d triangles.", sceneName, fetchedCount, totalTriangles)
						// This might happen if the server decides to close the stream early
						break // Assume end of stream for this batch
					}
				}
				if err.Error() == "EOF" { // gRPC stream end
					logrus.Infof("Finished streaming triangles for scene '%s'. Fetched %d of %d total triangles.", sceneName, fetchedCount, totalTriangles)
					break
				}
				return nil, fmt.Errorf("failed to receive triangle batch for scene '%s': %w", sceneName, err)
			}

			allTriangles = append(allTriangles, resp.GetTriangles()...)
			fetchedCount += uint64(len(resp.GetTriangles()))

			// If the batch returned fewer than requested (and we haven't hit total), it means we're at the end
			if uint64(len(resp.GetTriangles())) < uint64(batchSize) {
				logrus.Infof("Received partial triangle batch. Assuming end of stream for scene '%s'. Fetched %d of %d total triangles.", sceneName, fetchedCount, totalTriangles)
				break
			}
		}

		// After receiving all responses for one batch request, if fetchedCount is still less than totalTriangles,
		// the outer loop will continue for the next batch.
		if fetchedCount >= totalTriangles {
			break // All triangles fetched
		}
	}

	if fetchedCount != totalTriangles {
		return nil, fmt.Errorf("triangles stream for scene '%s' ended prematurely. Expected %d triangles, got %d", sceneName, totalTriangles, fetchedCount)
	}

	return allTriangles, nil
}

// RenderSetup is a streaming RPC to configure a worker node.
// It fetches scene data, textures, and triangles from the asset provider.
func (s *workerServer) RenderSetup(req *pb_control.RenderSetupRequest, stream pb_control.RenderControlService_RenderSetupServer) error {
	s.currentStatus = pb_discovery.WorkerStatus_ALLOCATED

	logrus.Printf("RenderControlService: RenderSetup called by %s", s.workerID)
	logrus.Printf("RenderSetup Configuration: Scene='%s', Sampler='%s', AssetProvider='%s', JobID='%s'",
		req.GetSceneName(), req.GetSampler().String(), req.GetAssetProvider(), req.GetJobId())

	assetProviderAddr := req.GetAssetProvider()
	if assetProviderAddr == "" {
		errMsg := "Asset provider address is empty"
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.InvalidArgument, errMsg)
	}

	ctx := stream.Context() // Use the stream's context for asset fetching client

	// Establish connection to asset provider (SceneTransportService)
	// Using WithBlock to ensure connection is established before proceeding
	assetConn, err := grpc.NewClient(assetProviderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errMsg := fmt.Sprintf("Failed to dial asset provider %s: %v", assetProviderAddr, err)
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.Unavailable, errMsg)
	}
	defer assetConn.Close()
	transportClient := pb_transport.NewSceneTransportServiceClient(assetConn)

	s.currentStatus = pb_discovery.WorkerStatus_BUSY_RENDER_SETUP

	// Step 1: Send LOADING_SCENE status and fetch scene file
	if err := s.sendStatus(stream, pb_control.RenderSetupStatus_LOADING_SCENE, ""); err != nil {
		return status.Errorf(codes.Internal, "failed to send LOADING_SCENE status: %v", err)
	}
	logrus.Infof("RenderSetup: Attempting to fetch scene '%s' from '%s'...", req.GetSceneName(), assetProviderAddr)

	scene, err := s.getScene(ctx, transportClient, req.GetSceneName())
	if err != nil {
		errMsg := fmt.Sprintf("Failed to load scene '%s': %v", req.GetSceneName(), err)
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.NotFound, errMsg) // Use NotFound for scene not found
	}
	s.loadedScene = scene // Store the loaded scene
	logrus.Infof("RenderSetup: Successfully loaded scene '%s' (version: %s). Contains %d materials, %d spheres.",
		scene.GetName(), scene.GetVersion(), len(scene.GetMaterials()), len(scene.GetObjects().GetSpheres()))

	if scene.GetStreamTriangles() {
		logrus.Infof("RenderSetup: Scene indicates triangles need to be streamed. Total triangles: %d", scene.GetTotalTriangles())
		triangles, err := s.streamTriangles(ctx, transportClient, scene.GetName(), scene.GetTotalTriangles(), 1000) // Fetch in batches of 1000
		if err != nil {
			errMsg := fmt.Sprintf("Failed to stream triangles for scene '%s': %v", scene.GetName(), err)
			s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
			return status.Error(codes.Internal, errMsg)
		}
		s.loadedTriangles = triangles // Store the loaded triangles
		logrus.Infof("RenderSetup: Successfully streamed %d triangles for scene '%s'.", len(triangles), scene.GetName())
	} else {
		logrus.Infof("RenderSetup: Scene indicates triangles are embedded or not streamed.")
		// If triangles are embedded, they would be in scene.GetObjects().GetTriangles()
		// You might want to copy them or process them here.
		s.loadedTriangles = scene.GetObjects().GetTriangles()
		logrus.Infof("RenderSetup: Using %d embedded triangles from scene.", len(s.loadedTriangles))
	}

	// Step 2: Send STREAMING_TEXTURES status and fetch textures
	if err := s.sendStatus(stream, pb_control.RenderSetupStatus_STREAMING_TEXTURES, ""); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to send STREAMING_TEXTURES status: %v", err))
	}
	logrus.Infof("RenderSetup: Streaming textures from '%s'...", assetProviderAddr)

	// Collect all unique ImageTexture filenames from materials
	texturesToFetch := make(map[string]*pb_transport.ImageTexture)
	for _, mat := range scene.GetMaterials() {
		if mat.GetLambert() != nil && mat.GetLambert().GetAlbedo() != nil {
			if imgTex := mat.GetLambert().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		// Add similar logic for other material types and their textures (e.g., DiffuseLight, Metal, PBR)
		if mat.GetDiffuselight() != nil && mat.GetDiffuselight().GetEmit() != nil {
			if imgTex := mat.GetDiffuselight().GetEmit().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		if mat.GetIsotropic() != nil && mat.GetIsotropic().GetAlbedo() != nil {
			if imgTex := mat.GetIsotropic().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		if mat.GetMetal() != nil && mat.GetMetal().GetAlbedo() != nil {
			if imgTex := mat.GetMetal().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		if mat.GetPbr() != nil {
			if imgTex := mat.GetPbr().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetRoughness().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetMetalness().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetNormalMap().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetSss().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
	}

	for filename, imgTex := range texturesToFetch {
		logrus.Infof("RenderSetup: Fetching texture '%s' (expected size: %d bytes)...", filename, imgTex.GetSize())
		texData, err := s.streamTextureFile(ctx, transportClient, filename, imgTex.GetSize())
		if err != nil {
			errMsg := fmt.Sprintf("Failed to load texture '%s': %v", filename, err)
			s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
			return status.Error(codes.Internal, errMsg)
		}
		s.loadedTextures[filename] = texData // Store the fetched texture data
		logrus.Infof("RenderSetup: Successfully loaded texture '%s'. Actual size: %d bytes.", filename, len(texData))
	}
	logrus.Infof("RenderSetup: Finished streaming %d unique textures.", len(s.loadedTextures))

	// Step 3: Send READY status
	if err := s.sendStatus(stream, pb_control.RenderSetupStatus_READY, ""); err != nil {
		return status.Errorf(codes.Internal, "failed to send READY status: %v", err)
	}
	logrus.Infof("RenderSetup: Worker is READY for rendering with scene '%s', %d triangles, and %d textures.",
		req.GetSceneName(), len(s.loadedTriangles), len(s.loadedTextures))

	s.currentStatus = pb_discovery.WorkerStatus_BUSY_RENDERING

	return nil // Successfully configured
}

func (s *workerServer) RenderTile(req *pb_control.RenderTileRequest, stream pb_control.RenderControlService_RenderTileServer) error {
	logrus.Printf("RenderControlService: RenderTile called by %s - Tile: [%d,%d] to [%d,%d)",
		s.workerID, req.GetX0(), req.GetY0(), req.GetX1(), req.GetY1())

	chunkWidth := uint32(16)
	chunkHeight := uint32(16)
	totalPixelsInChunk := int(chunkWidth * chunkHeight * 3)

	for y := req.GetY0(); y < req.GetY1(); y += chunkHeight {
		for x := req.GetX0(); x < req.X1; x += chunkWidth {
			select {
			case <-stream.Context().Done():
				logrus.Printf("RenderTile stream cancelled for tile [%d,%d]: %v", req.GetX0(), req.GetY0(), stream.Context().Err())
				return stream.Context().Err()
			default:
			}

			pixels := make([]float32, totalPixelsInChunk)

			resp := &pb_control.RenderTileResponse{
				Width:  chunkWidth,
				Height: chunkHeight,
				PosX:   x,
				PosY:   y,
				Pixels: pixels,
			}
			if err := stream.Send(resp); err != nil {
				logrus.Printf("Failed to send RenderTileResponse chunk for tile [%d,%d]: %v", req.GetX0(), req.GetY0(), err)
				return fmt.Errorf("failed to send stream chunk: %w", err)
			}
		}
	}

	logrus.Printf("RenderControlService: Finished streaming tile [%d,%d] to [%d,%d) by %s",
		req.GetX0(), req.GetY0(), req.GetX1(), req.GetY1(), s.workerID)
	return nil
}

func (s *workerServer) RenderEnd(ctx context.Context, req *pb_control.RenderEndRequest) (*pb_control.RenderEndResponse, error) {
	s.currentStatus = pb_discovery.WorkerStatus_FREE

	logrus.Printf("RenderControlService: RenderEnd called by %s", s.workerID)
	stats := &pb_control.RenderEndResponse{
		TotalRenderTimeMs: 12345,
		TotalRaysTraced:   987654321,
	}
	return stats, nil
}

// StartWorker initializes and runs the Izpi worker services.
func StartWorker(numCores uint32) {
	logrus.Infof("Starting Izpi Worker")
	logrus.Infof("Configured cores: %d", numCores)

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("Failed to get hostname for worker ID: %v", err)
	}
	workerID := hostname
	logrus.Infof("Worker ID (Hostname): %s", workerID)

	var assignedPort int

	// The listener on the worker side can still be on a random port.
	// The leader will get this port via mDNS.
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}

	assignedPort = lis.Addr().(*net.TCPAddr).Port
	logrus.Infof("gRPC server listening on all interfaces: %s (port %d)", lis.Addr().String(), assignedPort)

	grpcServer := grpc.NewServer()
	workerSrv := NewWorkerServer(numCores) // Assuming NewWorkerServer is defined elsewhere

	pb_discovery.RegisterWorkerDiscoveryServiceServer(grpcServer, workerSrv)
	pb_control.RegisterRenderControlServiceServer(grpcServer, workerSrv)
	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// --- Zeroconf (mDNS/DNS-SD) Advertising ---
	serviceType := "_izpi-worker._tcp"
	serviceName := "Izpi-Worker"

	txtRecords := []string{
		fmt.Sprintf("worker_id=%s", workerID),
		fmt.Sprintf("available_cores=%d", numCores),
		fmt.Sprintf("grpc_port=%d", assignedPort),
	}

	// Create a channel to signal when mDNS advertising should stop
	// This is important for both zeroconf.Register and avahi.EntryGroup
	var mDNSServerCloser func() // Function to call to stop mDNS advertisement

	currentOS := runtime.GOOS
	logrus.Infof("Detected operating system: %s", currentOS)

	switch currentOS {
	case "linux", "freebsd":
		// Use go-avahi for Linux and FreeBSD
		conn, err := dbus.SystemBus()
		if err != nil {
			logrus.Fatalf("Failed to connect to D-Bus system bus for Avahi: %v", err)
		}
		// conn.Close() will be deferred within mDNSServerCloser

		server, err := avahi.ServerNew(conn)
		if err != nil {
			conn.Close() // Close D-Bus connection on Avahi server creation failure
			logrus.Fatalf("Failed to create Avahi server client: %v", err)
		}
		// server.Close() will be deferred within mDNSServerCloser

		entryGroup, err := server.EntryGroupNew()
		if err != nil {
			server.Close() // Close Avahi server client on entry group creation failure
			conn.Close()   // Close D-Bus connection
			logrus.Fatalf("Failed to create new Avahi entry group: %v", err)
		}
		// entryGroup.Reset() will be deferred within mDNSServerCloser

		// Convert TXT records to []byte slice of slices for Avahi
		avahiTxtRecords := make([][]byte, len(txtRecords))
		for i, t := range txtRecords {
			avahiTxtRecords[i] = []byte(t)
		}

		err = entryGroup.AddService(
			avahi.InterfaceUnspec,
			avahi.ProtoUnspec,
			0, // Flags
			serviceName,
			serviceType,
			"local",  // Domain
			hostname, // Hostname
			uint16(assignedPort),
			avahiTxtRecords,
		)
		if err != nil {
			entryGroup.Reset() // Try to clean up
			server.Close()
			conn.Close()
			logrus.Fatalf("Failed to add service to Avahi entry group: %v", err)
		}

		err = entryGroup.Commit()
		if err != nil {
			entryGroup.Reset() // Try to clean up
			server.Close()
			conn.Close()
			logrus.Fatalf("Failed to commit Avahi entry group: %v", err)
		}
		logrus.Infof("Avahi service '%s.%s' registered successfully on port %d with TXT: %v", serviceName, serviceType, assignedPort, txtRecords)

		// Define the closer for Avahi
		mDNSServerCloser = func() {
			logrus.Info("Unpublishing Avahi service...")
			if err := entryGroup.Reset(); err != nil {
				logrus.Errorf("Error resetting Avahi entry group: %v", err)
			}
			server.Close()

			if err := conn.Close(); err != nil {
				logrus.Errorf("Error closing D-Bus connection: %v", err)
			}
		}

	case "darwin":
		// Use grandcat/zeroconf for macOS
		server, err := zeroconf.Register(
			serviceName,
			serviceType,
			"local.",
			assignedPort,
			txtRecords,
			nil, // interfaces: nil means all suitable interfaces
		)
		if err != nil {
			logrus.Fatalf("Failed to register Zeroconf service: %v", err)
		}
		logrus.Infof("Zeroconf service '%s' advertising on port %d with TXT: %v", serviceName, assignedPort, txtRecords)

		// Define the closer for grandcat/zeroconf
		mDNSServerCloser = func() {
			logrus.Info("Shutting down grandcat/zeroconf service...")
			server.Shutdown()
		}

	default:
		logrus.Warnf("Unsupported operating system for mDNS advertising: %s. mDNS will not be advertised.", currentOS)
		// Provide a no-op closer if mDNS isn't supported
		mDNSServerCloser = func() {
			logrus.Info("No mDNS service to shut down on this OS.")
		}
	}

	// Ensure the mDNS service is shut down gracefully on exit
	defer mDNSServerCloser()

	// --- Graceful Shutdown ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Blocks until a signal is received

	logrus.Info("Shutting down Izpi Worker...")
	grpcServer.GracefulStop()
	logrus.Info("gRPC server stopped.")
	logrus.Info("Izpi Worker shut down gracefully.")
}
