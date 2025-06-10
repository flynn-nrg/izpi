package main

import (
	"context"
	"fmt"
	"image"
	"net"
	"os"
	"path/filepath"
	"runtime"
	"runtime/pprof"
	"runtime/trace"
	"strings"
	"sync"
	"time"

	"github.com/flynn-nrg/izpi/internal/colours"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/output"
	"github.com/flynn-nrg/izpi/internal/postprocess"
	"github.com/flynn-nrg/izpi/internal/render"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/worker"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/alecthomas/kong"
	"github.com/grandcat/zeroconf"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"

	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
)

const (
	programName       = "izpi"
	defaultXSize      = "500"
	defaultYSize      = "500"
	defaultSamples    = "1000"
	defaultMaxDepth   = "50"
	defaultOutputFile = "output.png"
	defaultSceneFile  = "examples/cornell_box.yaml"

	displayWindowTitle = "Izpi Render Output"
)

var flags struct {
	LogLevel    string `name:"log-level" help:"The log level: error, warn, info, debug, trace." default:"info"`
	Scene       string `type:"existingfile" name:"scene" help:"Scene file to render" default:"${defaultSceneFile}"`
	NumWorkers  int64  `name:"num-workers" help:"Number of worker threads" default:"${defaultNumWorkers}"`
	XSize       int64  `name:"x" help:"Output image x size" default:"${defaultXSize}"`
	YSize       int64  `name:"y" help:"Output image y size" default:"${defaultYSize}"`
	Samples     int64  `name:"samples" help:"Number of samples per ray" default:"${defaultSamples}"`
	Sampler     string `name:"sampler-type" help:"Sampler function to use: colour, albedo, normal, wireframe" default:"colour"`
	Depth       int64  `name:"max-depth" help:"Maximum depth" default:"${defaultMaxDepth}"`
	OutputMode  string `name:"output-mode" help:"Output mode: png, exr, hdr or pfm" default:"png"`
	OutputFile  string `type:"file" name:"output-file" help:"Output file." default:"${defaultOutputFile}"`
	Verbose     bool   `name:"v" help:"Print rendering progress bar" default:"true"`
	Preview     bool   `name:"p" help:"Display rendering progress in a window" default:"true"`
	DisplayMode string `name:"display-mode" help:"Display mode: fyne or sdl" default:"fyne"`
	CpuProfile  string `name:"cpu-profile" help:"Enable cpu profiling"`
	Instrument  bool   `name:"instrument" help:"Enable instrumentation" default:"false"`
	Role        string `name:"role" help:"Role: worker, leader or standalone" default:"standalone"`
}

func main() {

	kong.Parse(&flags,
		kong.Name(programName),
		kong.Description("A path tracer implemented in Go"),
		kong.Vars{
			"defaultNumWorkers": fmt.Sprintf("%v", runtime.NumCPU()),
			"defaultXSize":      defaultXSize,
			"defaultYSize":      defaultYSize,
			"defaultSamples":    defaultSamples,
			"defaultMaxDepth":   defaultMaxDepth,
			"defaultOutputFile": defaultOutputFile,
			"defaultSceneFile":  defaultSceneFile,
		})

	setupLogging(flags.LogLevel)

	sceneFile, err := os.Open(flags.Scene)
	if err != nil {
		log.Fatal(err)
	}

	log.Infof("Running as %q", flags.Role)

	scene, err := scene.FromYAML(sceneFile, filepath.Dir(flags.Scene), 0)
	if err != nil {
		log.Fatalf("Error loading scene: %v", err)
	}

	if flags.CpuProfile != "" {
		f, err := os.Create(flags.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	if flags.Instrument {
		f, err := os.Create("trace.out")
		if err != nil {
			log.Fatalf("failed to create trace file: %v", err)
		}
		defer func() {
			if err := f.Close(); err != nil {
				log.Fatalf("failed to close trace file: %v", err)
			}
		}()

		if err := trace.Start(f); err != nil {
			log.Fatalf("failed to start trace: %v", err)
		}
		defer trace.Stop()
	}

	switch flags.Role {
	case "leader":
		run_as_leader(scene, false)
	case "standalone":
		run_as_leader(scene, true)
	case "worker":
		run_as_worker()
	default:
		log.Fatalf("unknown role %q", flags.Role)
	}
}

type workerHost struct {
	hostname string
	port     uint16
}

func run_as_leader(scene *scene.Scene, standalone bool) {
	var disp display.Display
	var err error
	var canvas image.Image

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

	//var workerHosts []workerHost

	if !standalone {
		// Example leader-side browsing logic (conceptual)
		resolver, err := zeroconf.NewResolver(nil)
		if err != nil {
			log.Fatalln("Failed to initialize resolver:", err.Error())
		}

		entries := make(chan *zeroconf.ServiceEntry)
		go func(results <-chan *zeroconf.ServiceEntry) {
			for entry := range results {
				log.Println(entry)
			}
			log.Println("No more entries.")
		}(entries)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(10))
		defer cancel()
		err = resolver.Browse(ctx, "_izpi-worker._tcp", "local.", entries)
		if err != nil {
			log.Fatalln("Failed to browse:", err.Error())
		}

		for entry := range entries {
			log.Printf("Discovered worker: %s, Hostname: %s, Addrs: %v, Port: %d, Text: %v",
				entry.Instance, entry.HostName, entry.AddrIPv4, entry.Port, entry.Text)
			// ... proceed to dial ...
		}
	}

	// Programmatically get the leader's local IPv4 address for binding
	leaderLocalIP, err := getLocalIPv4Addr("en0") // Assuming 'en0' is your primary interface
	if err != nil {
		logrus.Fatalf("Failed to get local IPv4 address for en0: %v", err)
	}
	logrus.Infof("Leader's local IP address for binding: %s", leaderLocalIP.String())

	// This is your hardcoded list of workers for testing
	workerHosts := []workerHost{
		{hostname: "vesper.local", port: 58595},
		// Add more workers here if needed for testing
	}

	for _, wh := range workerHosts {
		logrus.Infof("Processing worker: %s:%d", wh.hostname, wh.port)

		// --- Step 1: Resolve hostname to get ALL IP addresses ---
		ips, err := net.LookupIP(wh.hostname)
		if err != nil {
			logrus.Errorf("Failed to lookup IP for worker %s: %v", wh.hostname, err)
			continue
		}

		logrus.Infof("Resolved IPs for %s: %v", wh.hostname, ips)

		var ipv4Addr string
		for _, ip := range ips {
			if ip.To4() != nil { // Check if it's an IPv4 address
				ipv4Addr = ip.String()
				break // Use the first IPv4 address found
			}
		}

		if ipv4Addr == "" {
			logrus.Errorf("No IPv4 address found for worker %s. Cannot proceed with IPv4 connection.", wh.hostname)
			continue
		}

		// Construct target using the explicit IPv4 address
		target := fmt.Sprintf("%s:%d", ipv4Addr, wh.port)
		logrus.Infof("Chosen IPv4 target for %s: %s", wh.hostname, target)

		// --- Step 2: Attempt a raw TCP connection using net.Dialer to isolate the issue ---
		dialer := &net.Dialer{
			Timeout:   3 * time.Second,
			KeepAlive: 30 * time.Second,
		}

		logrus.Infof("Attempting raw TCP dial to %s...", target)
		logrus.Infof("Local IP being used: %s", leaderLocalIP.String())
		logrus.Infof("Network interface being used: en0")

		// First attempt: Try with explicit local address
		dialer.LocalAddr = &net.TCPAddr{IP: leaderLocalIP, Port: 0}
		rawConn, rawErr := dialer.DialContext(context.Background(), "tcp", target)
		if rawErr != nil {
			logrus.Warnf("First connection attempt failed: %v", rawErr)

			// Second attempt: Try without specifying local address
			logrus.Info("Retrying without specifying local address...")
			dialer.LocalAddr = nil
			rawConn, rawErr = dialer.DialContext(context.Background(), "tcp", target)
			if rawErr != nil {
				logrus.Errorf("Second connection attempt also failed: %v", rawErr)
				logrus.Errorf("Detailed error info: %+v", rawErr)
				continue
			}
		}

		logrus.Infof("Raw TCP dial to %s SUCCEEDED. Closing raw connection.", target)
		rawConn.Close() // Close the raw connection, it was just for testing the path

		// --- Step 3: Proceed with gRPC DialContext if raw dial succeeded ---
		logrus.Infof("Attempting gRPC DialContext to worker %s (resolved to %s)...", wh.hostname, target)

		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Use the same dialer configuration that succeeded for the raw connection
		conn, err := grpc.DialContext(ctx, target,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
			grpc.WithDialer(func(addr string, timeout time.Duration) (net.Conn, error) {
				return dialer.DialContext(context.Background(), "tcp", addr)
			}),
		)
		if err != nil {
			logrus.Errorf("Failed to connect to worker %s: %v", target, err)
			continue
		}
		defer conn.Close()

		// --- Step 4: Create gRPC client and make RPC call ---
		discoveryClient := pb_discovery.NewWorkerDiscoveryServiceClient(conn)

		logrus.Infof("Calling QueryWorkerStatus on %s...", target)
		statusResp, err := discoveryClient.QueryWorkerStatus(ctx, &pb_discovery.QueryWorkerStatusRequest{})
		if err != nil {
			logrus.Errorf("Failed to query status from worker %s: %v", target, err)
			continue
		}

		// Print the response from the worker
		logrus.Infof("--- Status from Worker %s (at %s) ---", statusResp.GetNodeName(), target)
		logrus.Infof("  Node Name: %s", statusResp.GetNodeName())
		logrus.Infof("  Available Cores: %d", statusResp.GetAvailableCores())
		logrus.Infof("  Total Memory: %d bytes", statusResp.GetTotalMemoryBytes())
		logrus.Infof("  Free Memory: %d bytes", statusResp.GetFreeMemoryBytes())
		logrus.Infof("  Status: %s", statusResp.GetStatus().String())
		logrus.Info("--------------------------------------")
	}

	log.Info("Finished querying all specified workers.")
	// In a real leader, you'd likely keep track of these workers and their status

	r := render.New(scene, int(flags.XSize), int(flags.YSize), int(flags.Samples), int(flags.Depth),
		colours.Black, colours.White, int(flags.NumWorkers), flags.Verbose, previewChan, flags.Preview, sampler.StringToType(flags.Sampler))

	wg := sync.WaitGroup{}
	wg.Add(1)
	// Detach the renderer as SDL needs to use the main thread for everything.
	go func() {
		canvas = r.Render()
		wg.Done()
	}()

	if flags.Preview {
		switch flags.DisplayMode {
		case "fyne":
			disp = display.NewFyneDisplay(displayWindowTitle, int(flags.XSize), int(flags.YSize), previewChan)
			disp.Start()
		case "sdl":
			disp = display.NewSDLDisplay(displayWindowTitle, int(flags.XSize), int(flags.YSize), previewChan)
			disp.Start()
		default:
			log.Fatalf("unknown display mode %q", flags.DisplayMode)
		}
	}

	wg.Wait()

	switch flags.OutputMode {
	case "png":
		pp := postprocess.NewPipeline([]postprocess.Filter{
			postprocess.NewGamma(),
			postprocess.NewClamp(1.0),
		})
		err = pp.Apply(canvas, scene)
		if err != nil {
			log.Fatal(err)
		}

		// Output
		out, err := output.NewPNG(flags.OutputFile)
		if err != nil {
			log.Fatal(err)
		}

		err = out.Write(canvas)
		if err != nil {
			log.Fatal(err)
		}

	case "exr":
		outFileName := strings.Replace(flags.OutputFile, "png", "exr", 1)
		out, err := output.NewOIIO(outFileName)
		if err != nil {
			log.Fatal(err)
		}

		err = out.Write(canvas)
		if err != nil {
			log.Fatal(err)
		}
	}

	if flags.Preview {
		disp.Wait()
	}
}

func run_as_worker() {
	worker.StartWorker(uint32(flags.NumWorkers))

}

func setupLogging(level string) {
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	switch level {
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "debug":
		log.SetLevel(log.DebugLevel)
	case "trace":
		log.SetLevel(log.TraceLevel)
	}
}

// getLocalIPv4Addr finds the IPv4 address of the specified network interface.
// It returns the first non-loopback IPv4 address found for that interface.
func getLocalIPv4Addr(ifaceName string) (net.IP, error) {
	iface, err := net.InterfaceByName(ifaceName)
	if err != nil {
		return nil, fmt.Errorf("could not find interface %s: %w", ifaceName, err)
	}

	addrs, err := iface.Addrs()
	if err != nil {
		return nil, fmt.Errorf("could not get addresses for interface %s: %w", ifaceName, err)
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if !ok {
			continue
		}
		if ipNet.IP.To4() != nil && !ipNet.IP.IsLoopback() {
			return ipNet.IP, nil
		}
	}
	return nil, fmt.Errorf("no IPv4 address found for interface %s", ifaceName)
}
