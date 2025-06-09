package main

import (
	"context"
	"fmt"
	"image"
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

	"github.com/alecthomas/kong"
	"github.com/grandcat/zeroconf"

	log "github.com/sirupsen/logrus"
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

func run_as_leader(scene *scene.Scene, standalone bool) {
	var disp display.Display
	var err error
	var canvas image.Image

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

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
