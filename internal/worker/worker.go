package worker

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	// Still imported for potential future use or if a fallback to UUID is needed
	"github.com/grandcat/zeroconf"
	"github.com/pbnjay/memory" // For system memory information
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"     // Specific import for control service protos
	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery" // Specific import for discovery service protos
	pb_empty "google.golang.org/protobuf/types/known/emptypb"         // Import for google.protobuf.Empty
)

// workerServer implements pb_discovery.WorkerDiscoveryServiceServer and pb_control.RenderControlServiceServer.
// It DOES NOT implement TransportServiceServer, as the worker will be a client for texture streaming.
type workerServer struct {
	// Embed the unimplemented server structs for each service to satisfy their interfaces.
	// This ensures our server won't break compilation if new methods are added to the .proto files,
	// and provides default (unimplemented) stubs.
	pb_discovery.UnimplementedWorkerDiscoveryServiceServer
	pb_control.UnimplementedRenderControlServiceServer

	workerID string // Now stores the hostname
	// System information fields
	availableCores   uint32 // Number of CPU cores available to the worker (tunable via flag)
	totalMemoryBytes uint64 // Total physical memory in bytes
	freeMemoryBytes  uint64 // Available physical memory in bytes
	status           pb_discovery.WorkerStatus
}

// NewWorkerServer creates and returns a new workerServer instance.
// It now determines its ID from the hostname and takes only the number of cores.
func NewWorkerServer(numCores uint32) *workerServer {
	// Get hostname to use as the worker ID
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("Failed to get hostname for worker ID: %v", err)
	}
	workerID := hostname // Use hostname as the worker ID

	// Get total and free system memory using the pbnjay/memory package
	totalMem := memory.TotalMemory()
	freeMem := memory.FreeMemory()

	// The number of CPUs to use is passed as a parameter, with validation handled by the main function.
	// We'll use this directly as the availableCores for the worker.

	return &workerServer{
		workerID:         workerID, // Assign hostname as workerID
		availableCores:   numCores,
		totalMemoryBytes: totalMem,
		freeMemoryBytes:  freeMem,
		status:           pb_discovery.WorkerStatus_FREE,
	}
}

// --- WorkerDiscoveryServiceServer Implementations (Stubs for now) ---

// QueryWorkerStatus is a stub implementation for the WorkerDiscoveryServiceServer.
// It returns actual system information stored in the workerServer instance.
func (s *workerServer) QueryWorkerStatus(ctx context.Context, req *pb_discovery.QueryWorkerStatusRequest) (*pb_discovery.QueryWorkerStatusResponse, error) {
	logrus.Printf("WorkerDiscoveryService: QueryWorkerStatus called on worker %s", s.workerID)
	// Return actual system metrics from the struct
	return &pb_discovery.QueryWorkerStatusResponse{
		NodeName:         s.workerID, // Using workerID (hostname) as node_name
		AvailableCores:   s.availableCores,
		TotalMemoryBytes: s.totalMemoryBytes,
		FreeMemoryBytes:  s.freeMemoryBytes,
		Status:           s.status,
	}, nil
}

// --- RenderControlServiceServer Implementations (Stubs for now) ---

// RenderConfiguration is a stub implementation for the RenderControlServiceServer.
// It currently does nothing and returns no error.
func (s *workerServer) RenderConfiguration(ctx context.Context, req *pb_control.RenderConfigurationRequest) (*pb_empty.Empty, error) {
	logrus.Printf("RenderControlService: RenderConfiguration called by %s - Scene: %s, Sampler: %s",
		s.workerID, req.GetSceneName(), req.GetSampler().String())
	// Implement actual render configuration logic here
	return &pb_empty.Empty{}, nil
}

// RenderTile is a stub implementation for the RenderControlServiceServer.
// It simulates streaming back empty pixel data for demonstration.
func (s *workerServer) RenderTile(req *pb_control.RenderTileRequest, stream pb_control.RenderControlService_RenderTileServer) error {
	logrus.Printf("RenderControlService: RenderTile called by %s - Tile: [%d,%d] to [%d,%d)",
		s.workerID, req.GetX0(), req.GetY0(), req.GetX1(), req.GetY1())

	// Simulate streaming back some empty data chunks for demonstration purposes.
	// In a real implementation, you would perform rendering and stream actual pixel data.
	chunkWidth := uint32(16)
	chunkHeight := uint32(16)
	totalPixelsInChunk := int(chunkWidth * chunkHeight * 4) // RGBA floats

	for y := req.GetY0(); y < req.GetY1(); y += chunkHeight {
		for x := req.GetX0(); x < req.X1; x += chunkWidth {
			select {
			case <-stream.Context().Done():
				logrus.Printf("RenderTile stream cancelled for tile [%d,%d]: %v", req.GetX0(), req.GetY0(), stream.Context().Err())
				return stream.Context().Err()
			default:
				// Continue
			}

			// Create a dummy pixel chunk (all zeros for now)
			pixels := make([]float32, totalPixelsInChunk)
			// You would fill 'pixels' with actual rendered data here

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

// RenderEnd is a stub implementation for the RenderControlServiceServer.
// It currently accepts an empty proto and returns a message with statistics.
func (s *workerServer) RenderEnd(ctx context.Context, req *pb_control.RenderEndRequest) (*pb_control.RenderEndResponse, error) {
	logrus.Printf("RenderControlService: RenderEnd called by %s", s.workerID)
	// Implement actual render cleanup logic and collect statistics here.
	// For now, return dummy statistics.
	stats := &pb_control.RenderEndResponse{
		TotalRenderTimeMs: 12345,     // Example: 12.345 seconds
		TotalRaysTraced:   987654321, // Example: ~1 billion rays
	}
	return stats, nil
}

// StartWorker initializes and runs the Izpi worker services.
func StartWorker(numCores uint32) {
	// Logrus setup is expected to be handled by the main application.
	// We will use logrus.Infof/Fatalf directly.

	logrus.Infof("Starting Izpi Worker")
	logrus.Infof("Configured cores: %d", numCores)

	// Get hostname to use as the worker ID
	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("Failed to get hostname for worker ID: %v", err)
	}
	workerID := hostname // Use hostname as the worker ID
	logrus.Infof("Worker ID (Hostname): %s", workerID)

	// --- gRPC Server Setup ---
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}

	assignedPort := lis.Addr().(*net.TCPAddr).Port

	grpcServer := grpc.NewServer()

	// Create an instance of our worker server implementation
	workerSrv := NewWorkerServer(numCores) // Removed ID parameter

	// Register our gRPC services with the server
	pb_discovery.RegisterWorkerDiscoveryServiceServer(grpcServer, workerSrv)
	pb_control.RegisterRenderControlServiceServer(grpcServer, workerSrv)

	// Register reflection service on gRPC server. Useful for gRPCurl/grpcc.
	reflection.Register(grpcServer)

	// Start gRPC server in a goroutine
	go func() {
		logrus.Infof("gRPC server listening on port %d...", assignedPort)
		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// --- Zeroconf (mDNS/DNS-SD) Advertising ---
	serviceType := "_izpi-worker._tcp"
	serviceName := "Izpi-Worker"

	txtRecords := []string{
		fmt.Sprintf("worker_id=%s", workerID), // worker_id is now the hostname
		fmt.Sprintf("available_cores=%d", numCores),
		fmt.Sprintf("grpc_port=%d", assignedPort),
	}

	// Get interfaces suitable for mDNS advertising
	ifaces, err := getMulticastInterfaces()
	if err != nil {
		logrus.Fatalf("Failed to get multicast interfaces: %v", err)
	}
	if len(ifaces) == 0 {
		logrus.Warn("No suitable network interfaces found for mDNS advertising.")
	}

	server, err := zeroconf.Register(
		serviceName,
		serviceType,
		"local.",
		assignedPort, // Use the assigned port here
		txtRecords,
		nil, //ifaces, // Pass the filtered interfaces here
	)
	if err != nil {
		logrus.Fatalf("Failed to register Zeroconf service: %v", err)
	}
	defer server.Shutdown()
	logrus.Infof("Zeroconf service '%s' advertising on port %d with TXT: %v", serviceName, assignedPort, txtRecords)

	// --- Graceful Shutdown ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	<-sigChan

	logrus.Info("Shutting down Izpi Worker...")
	grpcServer.GracefulStop()
	logrus.Info("gRPC server stopped.")
	logrus.Info("Izpi Worker shut down gracefully.")
}

// getMulticastInterfaces returns a slice of network interfaces suitable for mDNS.
// It filters out loopback, down, or point-to-point interfaces.
func getMulticastInterfaces() ([]net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var validIfaces []net.Interface
	for _, iface := range ifaces {
		// Skip loopback, down, and point-to-point interfaces
		if iface.Flags&net.FlagLoopback != 0 ||
			iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagPointToPoint != 0 {
			logrus.Debugf("Skipping interface %s: Flags: %v", iface.Name, iface.Flags)
			continue
		}

		// Check if the interface has a valid IP address (IPv4 or IPv6)
		addrs, err := iface.Addrs()
		if err != nil {
			logrus.Warnf("Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		hasIP := false
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil { // Check for valid IPv4 or IPv6
					hasIP = true
					break
				}
			}
		}

		if !hasIP {
			logrus.Debugf("Skipping interface %s: No valid IP address found", iface.Name)
			continue
		}

		logrus.Infof("Including interface %s for mDNS advertising (Flags: %v, Addrs: %v)", iface.Name, iface.Flags, addrs)
		validIfaces = append(validIfaces, iface)
	}
	return validIfaces, nil
}
