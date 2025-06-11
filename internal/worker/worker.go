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

func (s *workerServer) RenderEnd(ctx context.Context, req *pb_control.RenderEndRequest) (*pb_control.RenderEndResponse, error) {
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

	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		logrus.Fatalf("Failed to listen: %v", err)
	}

	assignedPort = lis.Addr().(*net.TCPAddr).Port
	logrus.Infof("gRPC server listening on all interfaces: %s (port %d)", lis.Addr().String(), assignedPort)

	grpcServer := grpc.NewServer()
	workerSrv := NewWorkerServer(numCores)

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
