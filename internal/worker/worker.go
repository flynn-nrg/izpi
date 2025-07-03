package worker

import (
	"fmt"
	"net"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"

	"github.com/godbus/dbus/v5"
	"github.com/grandcat/zeroconf"
	"github.com/holoplot/go-avahi"
	"github.com/pbnjay/memory"
	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

type workerServer struct {
	pb_discovery.UnimplementedWorkerDiscoveryServiceServer
	pb_control.UnimplementedRenderControlServiceServer

	workerID         string
	availableCores   uint32
	totalMemoryBytes uint64
	freeMemoryBytes  uint64
	currentStatus    pb_discovery.WorkerStatus

	scene            *scene.Scene
	sampler          sampler.Sampler
	samplesPerPixel  int
	numRays          uint64
	maxDepth         int
	imageResolutionX int
	imageResolutionY int
	background       *vec3.Vec3Impl
	ink              *vec3.Vec3Impl

	randPool sync.Pool

	wg sync.WaitGroup
}

// NewWorkerServer creates and returns a new workerServer instance.
func newWorkerServer(numCores uint32) *workerServer {
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
		randPool:         sync.Pool{New: func() interface{} { return fastrandom.NewWithDefaults() }},
	}
}

// StartWorker initializes and runs the Izpi worker services.
func StartWorker(numCores uint32) {
	log.Infof("Starting Izpi Worker")
	log.Infof("Configured cores: %d", numCores)

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("Failed to get hostname for worker ID: %v", err)
	}
	workerID := hostname
	log.Infof("Worker ID (Hostname): %s", workerID)

	var assignedPort int

	// The listener on the worker side can still be on a random port.
	// The leader will get this port via mDNS.
	lis, err := net.Listen("tcp", ":0")
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}

	assignedPort = lis.Addr().(*net.TCPAddr).Port
	log.Infof("gRPC server listening on all interfaces: %s (port %d)", lis.Addr().String(), assignedPort)

	grpcServer := grpc.NewServer()
	workerSrv := newWorkerServer(numCores)

	pb_discovery.RegisterWorkerDiscoveryServiceServer(grpcServer, workerSrv)
	pb_control.RegisterRenderControlServiceServer(grpcServer, workerSrv)
	reflection.Register(grpcServer)

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("Failed to serve gRPC: %v", err)
		}
	}()

	// --- Zeroconf (mDNS/DNS-SD) Advertising ---
	serviceType := "_izpi-worker._tcp"
	serviceName := fmt.Sprintf("Izpi-Worker-%s", workerID)

	txtRecords := []string{
		fmt.Sprintf("worker_id=%s", workerID),
		fmt.Sprintf("available_cores=%d", numCores),
		fmt.Sprintf("grpc_port=%d", assignedPort),
	}

	// Create a channel to signal when mDNS advertising should stop
	// This is important for both zeroconf.Register and avahi.EntryGroup
	var mDNSServerCloser func() // Function to call to stop mDNS advertisement

	currentOS := runtime.GOOS
	log.Infof("Operating environment: %s", currentOS)

	switch currentOS {
	case "linux", "freebsd":
		// Use go-avahi for Linux and FreeBSD
		conn, err := dbus.SystemBus()
		if err != nil {
			log.Fatalf("Failed to connect to D-Bus system bus for Avahi: %v", err)
		}

		server, err := avahi.ServerNew(conn)
		if err != nil {
			conn.Close() // Close D-Bus connection on Avahi server creation failure
			log.Fatalf("Failed to create Avahi server client: %v", err)
		}

		entryGroup, err := server.EntryGroupNew()
		if err != nil {
			server.Close() // Close Avahi server client on entry group creation failure
			conn.Close()   // Close D-Bus connection
			log.Fatalf("Failed to create new Avahi entry group: %v", err)
		}

		fqdnHostname := hostname + ".local"

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
			"local",      // Domain
			fqdnHostname, // Hostname
			uint16(assignedPort),
			avahiTxtRecords,
		)
		if err != nil {
			entryGroup.Reset() // Try to clean up
			server.Close()
			conn.Close()
			log.Fatalf("Failed to add service to Avahi entry group: %v", err)
		}

		err = entryGroup.Commit()
		if err != nil {
			entryGroup.Reset() // Try to clean up
			server.Close()
			conn.Close()
			log.Fatalf("Failed to commit Avahi entry group: %v", err)
		}
		log.Infof("Avahi service '%s.%s' registered successfully on port %d with TXT: %v", serviceName, serviceType, assignedPort, txtRecords)

		// Define the closer for Avahi
		mDNSServerCloser = func() {
			log.Info("Unpublishing Avahi service...")
			if err := entryGroup.Reset(); err != nil {
				log.Errorf("Error resetting Avahi entry group: %v", err)
			}
			server.Close()

			if err := conn.Close(); err != nil {
				log.Errorf("Error closing D-Bus connection: %v", err)
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
			log.Fatalf("Failed to register Zeroconf service: %v", err)
		}
		log.Infof("Zeroconf service '%s' advertising on port %d with TXT: %v", serviceName, assignedPort, txtRecords)

		// Define the closer for grandcat/zeroconf
		mDNSServerCloser = func() {
			log.Info("Shutting down grandcat/zeroconf service...")
			server.Shutdown()
		}

	default:
		log.Warnf("Unsupported operating system for mDNS advertising: %s. mDNS will not be advertised.", currentOS)
		// Provide a no-op closer if mDNS isn't supported
		mDNSServerCloser = func() {
			log.Info("No mDNS service to shut down on this OS.")
		}
	}

	// Ensure the mDNS service is shut down gracefully on exit
	defer mDNSServerCloser()

	// --- Graceful Shutdown ---
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan // Blocks until a signal is received

	log.Info("Shutting down Izpi Worker...")
	grpcServer.GracefulStop()
	log.Info("gRPC server stopped.")
	log.Info("Izpi Worker shut down gracefully.")
}
