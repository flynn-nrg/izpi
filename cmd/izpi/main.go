package main

import (
	"flag"
	"image"
	"log"
	"math/rand"
	"runtime"
	"sync"
	"time"

	"gitlab.com/flynn-nrg/izpi/pkg/display"
	"gitlab.com/flynn-nrg/izpi/pkg/output"
	"gitlab.com/flynn-nrg/izpi/pkg/postprocess"
	"gitlab.com/flynn-nrg/izpi/pkg/render"
	"gitlab.com/flynn-nrg/izpi/pkg/scenes"
)

func main() {
	var disp display.Display
	var err error
	var canvas image.Image

	numWorkers := flag.Int("num-workers", runtime.NumCPU(), "the number of worker threads")
	nx := flag.Int("x", 500, "output image x size")
	ny := flag.Int("y", 500, "output image y size")
	ns := flag.Int("samples", 1000, "number of samples per ray")
	outputFile := flag.String("output", "output.png", "output file")
	verbose := flag.Bool("v", false, "verbose")
	preview := flag.Bool("p", false, "display rendering progress in a window")

	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	// Render
	scene, err := scenes.PBRTest(float64(*nx) / float64(*ny))
	if err != nil {
		log.Fatal(err)
	}

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

	r := render.New(scene, *nx, *ny, *ns, *numWorkers, *verbose, previewChan, *preview)

	wg := sync.WaitGroup{}
	wg.Add(1)
	// Detach the renderer as SDL needs to use the main thread for everything.
	go func() {
		canvas = r.Render()
		wg.Done()
	}()

	if *preview {
		disp = display.NewSDLDisplay("Izpi Render Output", *nx, *ny, previewChan)
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
	out, err := output.NewPNG(*outputFile)
	if err != nil {
		log.Fatal(err)
	}

	err = out.Write(canvas)
	if err != nil {
		log.Fatal(err)
	}

	if *preview {
		disp.Wait()
	}
}
