package assetprovider

import (
	"context"
	"fmt"
	"net"
	"os" // Added to get hostname
	"sync"
	"unsafe"

	// For simulating delays
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"github.com/flynn-nrg/floatimage/floatimage"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport" // Your generated transport proto
	"github.com/flynn-nrg/izpi/internal/texture"
)

const (
	// DefaultChunkSize for streaming bytes.
	DefaultTextureChunkSize = 64 * 1024 // 64 KiB
	// DefaultTriangleBatchSize for streaming triangles.
	DefaultTriangleBatchSize = 1000 // 1000 triangles per batch
)

// assetProviderServer implements pb_transport.SceneTransportServiceServer
// and holds the assets to be served.
type assetProviderServer struct {
	pb_transport.UnimplementedSceneTransportServiceServer

	scene       *pb_transport.Scene          // The main scene graph
	textures    map[string]*texture.ImageTxt // Map of filename to texture data
	triangles   []*pb_transport.Triangle     // Slice of all triangles for streaming
	trianglesMu sync.RWMutex                 // Mutex for triangles access if concurrency is needed (not strictly for this simple server)
}

// AssetProvider manages the gRPC server for serving assets.
type AssetProvider struct {
	grpcServer *grpc.Server
	listener   net.Listener
	address    string
	serverImpl *assetProviderServer
	wg         sync.WaitGroup
}

// New creates a new AssetProvider, initializes its gRPC server,
// and starts serving assets in a new goroutine.
// It returns the listener address (target) and an error if initialization fails.
func New(scene *pb_transport.Scene, textures map[string]*texture.ImageTxt, triangles []*pb_transport.Triangle) (*AssetProvider, string, error) {
	// If no port is specified, gRPC will pick a random available port.
	// We listen on "0.0.0.0:0" to bind to all available interfaces on a random port.
	lis, err := net.Listen("tcp", "0.0.0.0:0")
	if err != nil {
		return nil, "", fmt.Errorf("failed to listen for asset provider: %w", err)
	}

	grpcServer := grpc.NewServer()
	serverImpl := &assetProviderServer{
		scene:     scene,
		textures:  textures,
		triangles: triangles,
	}

	pb_transport.RegisterSceneTransportServiceServer(grpcServer, serverImpl)
	reflection.Register(grpcServer) // Enable reflection for gRPCurl/debugging

	// Get the hostname and the dynamically assigned port
	hostname, err := os.Hostname()
	if err != nil {
		lis.Close() // Close the listener if we can't get the hostname
		return nil, "", fmt.Errorf("failed to get hostname for asset provider address: %w", err)
	}
	port := lis.Addr().(*net.TCPAddr).Port
	targetAddress := fmt.Sprintf("%s:%d", hostname, port) // Format as hostname:port

	ap := &AssetProvider{
		grpcServer: grpcServer,
		listener:   lis,
		address:    targetAddress, // Use the new hostname:port address
		serverImpl: serverImpl,
	}

	ap.wg.Add(1)
	go func() {
		defer ap.wg.Done()
		log.Infof("AssetProvider gRPC server listening on %s (reported as %s)", lis.Addr().String(), ap.address)
		if err := grpcServer.Serve(lis); err != nil {
			log.Errorf("AssetProvider gRPC server failed to serve: %v", err)
		}
	}()

	return ap, ap.address, nil
}

// Stop gracefully shuts down the AssetProvider's gRPC server.
func (ap *AssetProvider) Stop() {
	log.Infof("Shutting down AssetProvider gRPC server on %s...", ap.address)
	ap.grpcServer.GracefulStop()
	ap.wg.Wait()
	log.Infof("AssetProvider gRPC server on %s stopped.", ap.address)
}

// GetScene implements pb_transport.SceneTransportServiceServer.
// It returns the stored scene graph.
func (s *assetProviderServer) GetScene(ctx context.Context, req *pb_transport.GetSceneRequest) (*pb_transport.Scene, error) {
	log.Infof("AssetProvider: GetScene called for '%s'", req.GetSceneName())

	if s.scene == nil || s.scene.GetName() != req.GetSceneName() {
		return nil, status.Errorf(codes.NotFound, "scene '%s' not found", req.GetSceneName())
	}
	return s.scene, nil
}

// StreamTextureFile implements pb_transport.SceneTransportServiceServer.
// It streams the texture data in chunks.
func (s *assetProviderServer) StreamTextureFile(req *pb_transport.StreamTextureFileRequest, stream pb_transport.SceneTransportService_StreamTextureFileServer) error {
	log.Infof("AssetProvider: StreamTextureFile called for '%s' (offset: %d, chunk_size: %d)",
		req.GetFilename(), req.GetOffset(), req.GetChunkSize())

	imageText, ok := s.textures[req.GetFilename()]
	if !ok {
		return status.Errorf(codes.NotFound, "texture file '%s' not found", req.GetFilename())
	}

	var totalSize uint64
	var rawData []float64

	if textureData, ok := imageText.GetData().(*floatimage.Float64NRGBA); ok {
		rawData = textureData.Pix
		totalSize = uint64(imageText.GetData().Bounds().Dx() * imageText.GetData().Bounds().Dy() * 4 * int(unsafe.Sizeof(float64(0))))
	} else {
		return status.Errorf(codes.Internal, "texture data is not a floatimage.FloatNRGBA")
	}

	// rawData is []float64
	// Convert []float64 to []byte using unsafe.Slice
	byteData := unsafe.Slice((*byte)(unsafe.Pointer(&rawData[0])), len(rawData)*int(unsafe.Sizeof(float64(0))))

	// Ensure offset is within bounds
	if req.GetOffset() >= totalSize && totalSize > 0 {
		return status.Errorf(codes.OutOfRange, "offset %d is out of range for texture '%s' (total size %d)", req.GetOffset(), req.GetFilename(), totalSize)
	}

	// Calculate effective chunk size (use request's chunk_size or default)
	chunkSize := req.GetChunkSize()
	if chunkSize == 0 {
		chunkSize = DefaultTextureChunkSize
	}

	// Start streaming from the requested offset
	currentOffset := req.GetOffset()
	for currentOffset < totalSize {
		select {
		case <-stream.Context().Done():
			log.Warnf("AssetProvider: StreamTextureFile for '%s' cancelled by client.", req.GetFilename())
			return stream.Context().Err()
		default:
			// Continue
		}

		endOffset := currentOffset + uint64(chunkSize)
		if endOffset > totalSize {
			endOffset = totalSize
		}

		chunk := byteData[currentOffset:endOffset]

		resp := &pb_transport.StreamTextureFileResponse{
			Chunk: chunk,
			Size:  totalSize,
		}

		if err := stream.Send(resp); err != nil {
			log.Errorf("AssetProvider: Failed to send texture chunk for '%s': %v", req.GetFilename(), err)
			return fmt.Errorf("failed to send texture chunk: %w", err)
		}
		currentOffset = endOffset
	}

	log.Infof("AssetProvider: Finished streaming texture '%s'", req.GetFilename())

	return nil
}

// StreamTriangles implements pb_transport.SceneTransportServiceServer.
// It streams triangle data in batches.
func (s *assetProviderServer) StreamTriangles(req *pb_transport.StreamTrianglesRequest, stream pb_transport.SceneTransportService_StreamTrianglesServer) error {
	log.Infof("AssetProvider: StreamTriangles called for scene '%s' (batch_size: %d, offset: %d)",
		req.GetSceneName(), req.GetBatchSize(), req.GetOffset())

	if s.scene == nil || s.scene.GetName() != req.GetSceneName() {
		return status.Errorf(codes.NotFound, "scene '%s' not found for triangle streaming", req.GetSceneName())
	}

	s.trianglesMu.RLock()
	totalTriangles := uint64(len(s.triangles))
	s.trianglesMu.RUnlock()

	if totalTriangles == 0 {
		log.Infof("AssetProvider: No triangles to stream for scene '%s'.", req.GetSceneName())
		return nil // No triangles to stream
	}

	// Ensure offset is within bounds
	if req.GetOffset() >= totalTriangles && totalTriangles > 0 {
		return status.Errorf(codes.OutOfRange, "offset %d is out of range for triangles in scene '%s' (total %d)", req.GetOffset(), req.GetSceneName(), totalTriangles)
	}

	// Calculate effective batch size
	batchSize := req.GetBatchSize()
	if batchSize == 0 {
		batchSize = DefaultTriangleBatchSize
	}

	currentOffset := req.GetOffset()
	for currentOffset < totalTriangles {
		select {
		case <-stream.Context().Done():
			log.Warnf("AssetProvider: StreamTriangles for '%s' cancelled by client.", req.GetSceneName())
			return stream.Context().Err()
		default:
			// Continue
		}

		endIndex := currentOffset + uint64(batchSize)
		if endIndex > totalTriangles {
			endIndex = totalTriangles
		}

		s.trianglesMu.RLock()
		batch := s.triangles[currentOffset:endIndex]
		s.trianglesMu.RUnlock()

		resp := &pb_transport.StreamTrianglesResponse{
			Triangles: batch,
		}

		if err := stream.Send(resp); err != nil {
			log.Errorf("AssetProvider: Failed to send triangle batch for '%s': %v", req.GetSceneName(), err)
			return fmt.Errorf("failed to send triangle batch: %w", err)
		}
		currentOffset = endIndex
	}

	log.Infof("AssetProvider: Finished streaming triangles for scene '%s' (total %d).", req.GetSceneName(), totalTriangles)

	return nil
}
