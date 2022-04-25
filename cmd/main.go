package main

import (
	"flag"
	"fmt"
	"image"
	"math/rand"
	"time"

	"gitlab.com/flynn-nrg/izpi/pkg/render"
	"gitlab.com/flynn-nrg/izpi/pkg/scenes"
)

func main() {

	numWorkers := flag.Int("num-workers", 1, "the number of worker threads")
	nx := flag.Int("x", 500, "output image x size")
	ny := flag.Int("y", 500, "output image y size")
	ns := flag.Int("samples", 1000, "number of samples per ray")

	flag.Parse()

	canvas := image.NewNRGBA(image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: *nx, Y: *ny}})
	rand.Seed(time.Now().UnixNano())

	world, cam := scenes.CornellBox(float64(*nx) / float64(*ny))
	render.Render(cam, world, canvas, *ns, *numWorkers)

	fmt.Printf("P3\n%v %v\n255\n", *nx, *ny)
	for j := *ny - 1; j >= 0; j-- {
		for i := 0; i < *nx; i++ {
			pixel := canvas.At(i, j)
			r, g, b, _ := pixel.RGBA()
			fmt.Printf("%v %v %v\n", r>>8, g>>8, b>>8)
		}
	}
}
