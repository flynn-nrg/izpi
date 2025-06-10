package worker

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"

	"github.com/grandcat/zeroconf"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
	pb_empty "google.golang.org/protobuf/types/known/emptypb"
)

// workerServer implements pb_discovery.WorkerDiscoveryServiceServer and pb_control.RenderControlServiceServer.
// It DOES NOT implement TransportServiceServer, as the worker will be a client for texture streaming.
type workerServer struct {
	pb_discovery.UnimplementedWorkerDiscoveryServiceServer
	pb_control.UnimplementedRenderControlServiceServer

	workerID         string
	availableCores   uint32
	totalMemoryBytes uint64
	freeMemoryBytes  uint64
	currentStatus    pb_discovery.WorkerStatus
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

func (s *workerServer) RenderConfiguration(ctx context.Context, req *pb_control.RenderConfigurationRequest) (*pb_empty.Empty, error) {
	logrus.Printf("RenderControlService: RenderConfiguration called by %s - Scene: %s, Sampler: %s",
		s.workerID, req.GetSceneName(), req.GetSampler().String())
	return &pb_empty.Empty{}, nil
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

func (s *workerServer) RenderEnd(ctx context.Context, req *pb_empty.Empty) (*pb_control.RenderEndResponse, error) {
	logrus.Printf("RenderControlService: RenderEnd called by %s", s.workerID)
	stats := &pb_control.RenderEndResponse{
		TotalRenderTimeMs: 12345,
		TotalRaysTraced:   987654321,
	}
	return stats, nil
}

// getMulticastInterfaces returns a slice of network interfaces suitable for mDNS.
func getMulticastInterfaces() ([]net.Interface, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return nil, fmt.Errorf("failed to get network interfaces: %w", err)
	}

	var validIfaces []net.Interface
	for _, iface := range ifaces {
		if iface.Flags&net.FlagLoopback != 0 ||
			iface.Flags&net.FlagUp == 0 ||
			iface.Flags&net.FlagPointToPoint != 0 {
			logrus.Debugf("Skipping interface %s: Flags: %v", iface.Name, iface.Flags)
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			logrus.Warnf("Failed to get addresses for interface %s: %v", iface.Name, err)
			continue
		}
		hasIP := false
		for _, addr := range addrs {
			if ipNet, ok := addr.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ipNet.IP.To4() != nil || ipNet.IP.To16() != nil {
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

// StartWorker initializes and runs the Izpi worker services.
func StartWorker(port int, numCores uint32) {
	logrus.Infof("Starting Izpi Worker")
	logrus.Infof("Configured cores: %d", numCores)

	hostname, err := os.Hostname()
	if err != nil {
		logrus.Fatalf("Failed to get hostname for worker ID: %v", err)
	}
	workerID := hostname
	logrus.Infof("Worker ID (Hostname): %s", workerID)

	// --- gRPC Server Setup: Explicitly listen on both IPv4 and IPv6 ---
	var lis net.Listener
	var assignedPort int

	// Listen on IPv4
	lis4, err4 := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err4 != nil {
		logrus.Warnf("Failed to listen on IPv4 (will try IPv6): %v", err4)
	} else {
		assignedPort = lis4.Addr().(*net.TCPAddr).Port
		logrus.Infof("gRPC server listening on IPv4: %s (port %d)", lis4.Addr().String(), assignedPort)
		lis = lis4 // Set the primary listener to IPv4
	}

	// Listen on IPv6 (even if IPv4 succeeded, as it's typically dual-stack aware on macOS)
	// We'll create a separate listener if IPv4 was explicitly bound, or fall back if IPv4 failed.
	var lis6 net.Listener
	lis6, err6 := net.Listen("tcp6", fmt.Sprintf(":%d", port))
	if err6 != nil {
		logrus.Warnf("Failed to listen on IPv6: %v", err6)
	} else {
		// If IPv4 didn't succeed, use this as the primary listener and set the port
		if lis == nil {
			assignedPort = lis6.Addr().(*net.TCPAddr).Port
			logrus.Infof("gRPC server listening on IPv6: %s (port %d)", lis6.Addr().String(), assignedPort)
			lis = lis6
		} else {
			// If both IPv4 and IPv6 listeners are created, ensure they are on the same port (which Listen will handle)
			logrus.Infof("gRPC server also listening on IPv6: %s (port %d)", lis6.Addr().String(), lis6.Addr().(*net.TCPAddr).Port)
		}
	}

	if lis == nil {
		logrus.Fatalf("Failed to start gRPC server: Could not listen on either IPv4 or IPv6.")
	}

	grpcServer := grpc.NewServer()
	workerSrv := NewWorkerServer(numCores)

	pb_discovery.RegisterWorkerDiscoveryServiceServer(grpcServer, workerSrv)
	pb_control.RegisterRenderControlServiceServer(grpcServer, workerSrv)
	reflection.Register(grpcServer)

	go func() {
		// Start serving on the primary listener
		if err := grpcServer.Serve(lis); err != nil {
			logrus.Fatalf("Failed to serve gRPC on primary listener: %v", err)
		}
	}()

	// If a separate IPv6 listener was created, start serving it too
	if lis6 != nil && lis6 != lis {
		go func() {
			if err := grpcServer.Serve(lis6); err != nil {
				logrus.Errorf("Failed to serve gRPC on secondary IPv6 listener: %v", err)
			}
		}()
	}

	// --- Zeroconf (mDNS/DNS-SD) Advertising ---
	serviceType := "_izpi-worker._tcp"
	serviceName := "Izpi-Worker"

	txtRecords := []string{
		fmt.Sprintf("worker_id=%s", workerID),
		fmt.Sprintf("available_cores=%d", numCores),
		fmt.Sprintf("grpc_port=%d", assignedPort),
	}

	server, err := zeroconf.Register(
		serviceName,
		serviceType,
		"local.",
		assignedPort,
		txtRecords,
		nil,
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
