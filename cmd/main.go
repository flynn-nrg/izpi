package main

import (
	"flag"
	"fmt"
	"image"
	"math/rand"
	"time"

	"gitlab.com/flynn-nrg/izpi/pkg/camera"
	"gitlab.com/flynn-nrg/izpi/pkg/render"
	"gitlab.com/flynn-nrg/izpi/pkg/scenes"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

func main() {
	numWorkers := flag.Int("num-workers", 1, "the number of worker threads")
	nx := flag.Int("x", 400, "output image x size")
	ny := flag.Int("y", 200, "output image y size")
	ns := flag.Int("samples", 100, "number of samples per ray")

	flag.Parse()

	canvas := image.NewNRGBA(image.Rectangle{Min: image.Point{X: 0, Y: 0}, Max: image.Point{X: *nx, Y: *ny}})
	rand.Seed(time.Now().UnixNano())

	fmt.Printf("P3\n%v %v\n255\n", *nx, *ny)

	world := scenes.Final()
	lookFrom := &vec3.Vec3Impl{X: 478.0, Y: 278.0, Z: -600.0}
	lookAt := &vec3.Vec3Impl{X: 278, Y: 278, Z: 0}
	vup := &vec3.Vec3Impl{Y: 1}
	distToFocus := 10.0
	aperture := 0.0
	aspect := float64(*nx) / float64(*ny)
	vfov := float64(40.0)
	time0 := 0.0
	time1 := 1.0
	cam := camera.New(lookFrom, lookAt, vup, vfov, aspect, aperture, distToFocus, time0, time1)

	render.Render(cam, world, canvas, *ns, *numWorkers)

	for j := *ny - 1; j >= 0; j-- {
		for i := 0; i < *nx; i++ {
			pixel := canvas.At(i, j)
			r, g, b, _ := pixel.RGBA()
			fmt.Printf("%v %v %v\n", r>>8, g>>8, b>>8)
		}
	}
}
