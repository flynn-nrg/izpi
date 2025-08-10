package main

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"runtime/pprof"
	"runtime/trace"

	"github.com/flynn-nrg/izpi/internal/config"
	"github.com/flynn-nrg/izpi/internal/leader"
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
	defaultOutputFile       = "output.exr"
	defaultSceneFile        = "examples/cornell_box_transparent_pyramid_spectral.pbtxt"
	defaultDiscoveryTimeout = "3"
)

var flags struct {
	LogLevel         string `name:"log-level" help:"The log level: error, warn, info, debug, trace." default:"info"`
	Scene            string `type:"existingfile" name:"scene" help:"Scene file to render" default:"${defaultSceneFile}"`
	NumWorkers       int64  `name:"num-workers" help:"Number of worker threads" default:"${defaultNumWorkers}"`
	XSize            int64  `name:"x" help:"Output image x size" default:"${defaultXSize}"`
	YSize            int64  `name:"y" help:"Output image y size" default:"${defaultYSize}"`
	Samples          int64  `name:"samples" help:"Number of samples per ray" default:"${defaultSamples}"`
	Sampler          string `name:"sampler-type" help:"Sampler function to use: spectral, colour, albedo, normal, wireframe" default:"colour"`
	Depth            int64  `name:"max-depth" help:"Maximum depth" default:"${defaultMaxDepth}"`
	OutputMode       string `name:"output-mode" help:"Output mode: png, exr, hdr or pfm" default:"exr"`
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
	ctx := context.Background()

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

	log.Infof("Running as %q", flags.Role)

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

	cfg := &config.Config{
		Scene:            flags.Scene,
		NumWorkers:       flags.NumWorkers,
		XSize:            flags.XSize,
		YSize:            flags.YSize,
		Samples:          flags.Samples,
		Sampler:          flags.Sampler,
		Depth:            flags.Depth,
		OutputMode:       flags.OutputMode,
		OutputFile:       flags.OutputFile,
		Verbose:          flags.Verbose,
		Preview:          flags.Preview,
		DisplayMode:      flags.DisplayMode,
		DiscoveryTimeout: flags.DiscoveryTimeout,
	}

	switch flags.Role {
	case "leader":
		runAsLeader(ctx, cfg, false)
	case "standalone":
		runAsLeader(ctx, cfg, true)
	case "worker":
		runAsWorker()
	default:
		log.Fatalf("unknown role %q", flags.Role)
	}
}

func runAsLeader(ctx context.Context, cfg *config.Config, standalone bool) {
	leader.RunAsLeader(ctx, cfg, standalone)
}

func runAsWorker() {
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
