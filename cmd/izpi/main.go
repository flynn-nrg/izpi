package main

import (
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
	"github.com/flynn-nrg/izpi/internal/discovery"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/output"
	"github.com/flynn-nrg/izpi/internal/postprocess"
	"github.com/flynn-nrg/izpi/internal/render"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/worker"

	"github.com/alecthomas/kong"

	log "github.com/sirupsen/logrus"
)

const (
	programName             = "izpi"
	defaultXSize            = "500"
	defaultYSize            = "500"
	defaultSamples          = "1000"
	defaultMaxDepth         = "50"
	defaultOutputFile       = "output.png"
	defaultSceneFile        = "examples/cornell_box.yaml"
	defaultDiscoveryTimeout = "3"

	displayWindowTitle = "Izpi Render Output"
)

var flags struct {
	LogLevel         string `name:"log-level" help:"The log level: error, warn, info, debug, trace." default:"info"`
	Scene            string `type:"existingfile" name:"scene" help:"Scene file to render" default:"${defaultSceneFile}"`
	NumWorkers       int64  `name:"num-workers" help:"Number of worker threads" default:"${defaultNumWorkers}"`
	XSize            int64  `name:"x" help:"Output image x size" default:"${defaultXSize}"`
	YSize            int64  `name:"y" help:"Output image y size" default:"${defaultYSize}"`
	Samples          int64  `name:"samples" help:"Number of samples per ray" default:"${defaultSamples}"`
	Sampler          string `name:"sampler-type" help:"Sampler function to use: colour, albedo, normal, wireframe" default:"colour"`
	Depth            int64  `name:"max-depth" help:"Maximum depth" default:"${defaultMaxDepth}"`
	OutputMode       string `name:"output-mode" help:"Output mode: png, exr, hdr or pfm" default:"png"`
	OutputFile       string `type:"file" name:"output-file" help:"Output file." default:"${defaultOutputFile}"`
	Verbose          bool   `name:"v" help:"Print rendering progress bar" default:"true"`
	Preview          bool   `name:"p" help:"Display rendering progress in a window" default:"true"`
	DisplayMode      string `name:"display-mode" help:"Display mode: fyne or sdl" default:"fyne"`
	CpuProfile       string `name:"cpu-profile" help:"Enable cpu profiling"`
	Instrument       bool   `name:"instrument" help:"Enable instrumentation" default:"false"`
	Role             string `name:"role" help:"Role: worker, leader or standalone" default:"standalone"`
	DiscoveryTimeout int64  `name:"discovery-timeout" help:"Discovery timeout in seconds" default:"${defaultDiscoveryTimeout}"`
}

func main() {

	kong.Parse(&flags,
		kong.Name(programName),
		kong.Description("A path tracer implemented in Go"),
		kong.Vars{
			"defaultNumWorkers":       fmt.Sprintf("%v", runtime.NumCPU()),
			"defaultXSize":            defaultXSize,
			"defaultYSize":            defaultYSize,
			"defaultSamples":          defaultSamples,
			"defaultMaxDepth":         defaultMaxDepth,
			"defaultOutputFile":       defaultOutputFile,
			"defaultSceneFile":        defaultSceneFile,
			"defaultDiscoveryTimeout": defaultDiscoveryTimeout,
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

func run_as_leader(scene *scene.Scene, standalone bool) {
	var disp display.Display
	var err error
	var canvas image.Image

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

	if !standalone {
		discovery, err := discovery.New(time.Second * time.Duration(flags.DiscoveryTimeout))
		if err != nil {
			log.Fatalln("Failed to initialize discovery:", err.Error())
		}

		workerHosts, err := discovery.FindWorkers()
		if err != nil {
			log.Fatalln("Failed to find workers:", err.Error())
		}

		log.Infof("Found %d worker(s)", len(workerHosts))
		for _, workerHost := range workerHosts {
			log.Infof("Worker: %s", workerHost.GetNodeName())
		}
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
