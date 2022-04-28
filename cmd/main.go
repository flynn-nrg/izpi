package main

import (
	"flag"
	"log"
	"math/rand"
	"runtime"
	"time"

	"gitlab.com/flynn-nrg/izpi/pkg/output"
	"gitlab.com/flynn-nrg/izpi/pkg/render"
	"gitlab.com/flynn-nrg/izpi/pkg/scenes"
)

func main() {

	numWorkers := flag.Int("num-workers", runtime.NumCPU(), "the number of worker threads")
	nx := flag.Int("x", 500, "output image x size")
	ny := flag.Int("y", 500, "output image y size")
	ns := flag.Int("samples", 1000, "number of samples per ray")
	outputFile := flag.String("output", "output.png", "output file")
	verbose := flag.Bool("v", false, "verbose")
	flag.Parse()

	rand.Seed(time.Now().UnixNano())

	world, cam := scenes.CornellBox(float64(*nx) / float64(*ny))
	r := render.New(cam, world, *nx, *ny, *ns, *numWorkers, 10, *verbose)
	canvas := r.Render()

	out, err := output.NewPNG(*outputFile)
	if err != nil {
		log.Fatal(err)
	}

	err = out.Write(canvas)
	if err != nil {
		log.Fatal(err)
	}
}
