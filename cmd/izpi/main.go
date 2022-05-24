package main

import (
	"fmt"
	"image"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"github.com/flynn-nrg/izpi/pkg/display"
	"github.com/flynn-nrg/izpi/pkg/output"
	"github.com/flynn-nrg/izpi/pkg/postprocess"
	"github.com/flynn-nrg/izpi/pkg/render"
	"github.com/flynn-nrg/izpi/pkg/scenes"

	"github.com/alecthomas/kong"

	log "github.com/sirupsen/logrus"
)

const (
	programName       = "izpi"
	defaultXSize      = "500"
	defaultYSize      = "500"
	defaultSamples    = "1000"
	defaultOutputFile = "output.png"
)

var flags struct {
	LogLevel   string `name:"log-level" help:"The log level: error, warn, info, debug, trace." default:"info"`
	NumWorkers int64  `name:"num-workers" help:"Number of worker threads" default:"${defaultNumWorkers}"`
	XSize      int64  `name:"x" help:"Output image x size" default:"${defaultXSize}"`
	YSize      int64  `name:"y" help:"Output image y size" default:"${defaultYSize}"`
	Samples    int64  `name:"samples" help:"Number of samples per ray" default:"${defaultSamples}"`
	OutputFile string `type:"file" name:"output-file" help:"Output file." default:"${defaultOutputFile}"`
	Verbose    bool   `name:"v" help:"Print rendering progress bar"`
	Preview    bool   `name:"p" help:"Display rendering progress in a window"`
	Normal     bool   `name:"n" help:"Render the normals at the ray intersection point"`
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
			"defaultOutputFile": defaultOutputFile,
		})

	rand.Seed(time.Now().UnixNano())

	setupLogging(flags.LogLevel)

	// Render
	scene, err := scenes.DisplacementTest(float64(flags.XSize) / float64(flags.YSize))
	if err != nil {
		log.Fatal(err)
	}

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

	r := render.New(scene, int(flags.XSize), int(flags.YSize), int(flags.Samples), int(flags.NumWorkers), flags.Verbose, previewChan, flags.Preview, flags.Normal)

	wg := sync.WaitGroup{}
	wg.Add(1)
	// Detach the renderer as SDL needs to use the main thread for everything.
	go func() {
		canvas = r.Render()
		wg.Done()
	}()

	if flags.Preview {
		disp = display.NewSDLDisplay("Izpi Render Output", int(flags.XSize), int(flags.YSize), previewChan)
		disp.Start()
	}

	wg.Wait()

	// Post-process pipeline.
	//file, err := os.Open("test.cube")
	//if err != nil {
	//	log.Fatal(err)
	//}
	//cg, err := postprocess.NewColourGradingFromCube(file)
	//if err != nil {
	//	log.Fatal(err)
	//}
	pp := postprocess.NewPipeline([]postprocess.Filter{
		postprocess.NewClamp(1.0),
		//	cg,
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
