package main

import (
	"fmt"
	"image"
	"os"
	"runtime"
	"runtime/pprof"
	"strings"
	"sync"

	"github.com/flynn-nrg/izpi/pkg/colours"
	"github.com/flynn-nrg/izpi/pkg/display"
	"github.com/flynn-nrg/izpi/pkg/floatimage"
	"github.com/flynn-nrg/izpi/pkg/output"
	"github.com/flynn-nrg/izpi/pkg/postprocess"
	"github.com/flynn-nrg/izpi/pkg/render"
	"github.com/flynn-nrg/izpi/pkg/sampler"
	"github.com/flynn-nrg/izpi/pkg/scenes"

	"github.com/alecthomas/kong"

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
	OutputMode  string `name:"output-mode" help:"Output mode: png, hdr or pfm" default:"png"`
	OutputFile  string `type:"file" name:"output-file" help:"Output file." default:"${defaultOutputFile}"`
	Verbose     bool   `name:"v" help:"Print rendering progress bar" default:"true"`
	Preview     bool   `name:"p" help:"Display rendering progress in a window" default:"true"`
	DisplayMode string `name:"display-mode" help:"Display mode: fyne or sdl" default:"fyne"`
	CpuProfile  string `name:"cpu-profile" help:"Enable cpu profiling"`
}

func main() {
	var disp display.Display
	var err error
	var canvas image.Image

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

	/*
		sceneFile, err := os.Open(flags.Scene)
		if err != nil {
			log.Fatal(err)
		}

		scene, err := scene.FromYAML(sceneFile, filepath.Dir(flags.Scene), 0)
		if err != nil {
			log.Fatal(err)
		}
	*/

	scene, err := scenes.Challenger(float64(flags.XSize) / float64(flags.YSize))
	if err != nil {
		log.Fatal(err)
	}

	if flags.CpuProfile != "" {
		f, err := os.Create(flags.CpuProfile)
		if err != nil {
			log.Fatal(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

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
	case "hdr":
		var hdrCanvas image.Image
		var err error

		if floatNRGBACanvas, ok := canvas.(*floatimage.FloatNRGBA); ok {
			hdrCanvas, err = floatNRGBACanvas.ToHDR()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("image format is not FloatNRGBA")
		}

		outFileName := strings.Replace(flags.OutputFile, "png", "hdr", 1)
		out, err := output.NewHDR(outFileName)
		if err != nil {
			log.Fatal(err)
		}

		err = out.Write(hdrCanvas)
		if err != nil {
			log.Fatal(err)
		}
	case "pfm":
		var hdrCanvas image.Image
		var err error

		if floatNRGBACanvas, ok := canvas.(*floatimage.FloatNRGBA); ok {
			hdrCanvas, err = floatNRGBACanvas.ToHDR()
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal("image format is not FloatNRGBA")
		}

		outFileName := strings.Replace(flags.OutputFile, "png", "pfm", 1)
		out, err := output.NewPFM(outFileName)
		if err != nil {
			log.Fatal(err)
		}

		err = out.Write(hdrCanvas)
		if err != nil {
			log.Fatal(err)
		}
	}

	if flags.Preview {
		disp.Wait()
	}
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
