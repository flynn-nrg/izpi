// Package render implements the main rendering loop.
package render

import (
	"image"
	"image/color"
	"math"
	"math/rand"
	"sync"

	"gitlab.com/flynn-nrg/izpi/pkg/camera"
	"gitlab.com/flynn-nrg/izpi/pkg/hitable"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

type workUnit struct {
	cam        *camera.Camera
	world      *hitable.HitableSlice
	canvas     *image.NRGBA
	numSamples int
	x0         int
	x1         int
	y0         int
	y1         int
}

func colour(r ray.Ray, world *hitable.HitableSlice, depth int) *vec3.Vec3Impl {
	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		scattered, attenuation, ok := mat.Scatter(r, rec)
		emitted := mat.Emitted(rec.U(), rec.V(), rec.P())
		if depth < 50 && ok {
			// emitted + (attenuation * color)
			return vec3.Add(emitted, vec3.Mul(attenuation, colour(scattered, world, depth+1)))
		} else {
			return emitted
		}
	}
	return &vec3.Vec3Impl{}
}

func clamp(f float64) uint8 {
	i := int(255.99 * f)
	if i < 256 {
		return uint8(i)
	}

	return 255
}
func renderRect(w workUnit) {
	nx := w.canvas.Bounds().Max.X
	ny := w.canvas.Bounds().Max.Y
	for y := w.y0; y <= w.y1; y++ {
		for x := w.x0; x <= w.x1; x++ {
			col := &vec3.Vec3Impl{}
			for s := 0; s < w.numSamples; s++ {
				u := (float64(x) + rand.Float64()) / float64(nx)
				v := (float64(y) + rand.Float64()) / float64(ny)
				r := w.cam.GetRay(u, v)
				col = vec3.Add(col, colour(r, w.world, 0))
			}

			col = vec3.ScalarDiv(col, float64(w.numSamples))
			// gamma 2
			col = &vec3.Vec3Impl{X: math.Sqrt(col.X), Y: math.Sqrt(col.Y), Z: math.Sqrt(col.Z)}
			ir := clamp(col.X)
			ig := clamp(col.Y)
			ib := clamp(col.Z)
			w.canvas.SetNRGBA(x, y, color.NRGBA{R: ir, G: ig, B: ib, A: 255})
		}
	}
}

func worker(input chan workUnit, quit chan struct{}, wg sync.WaitGroup) {
	wg.Add(1)
	defer wg.Done()
	for {
		select {
		case w := <-input:
			renderRect(w)
		case <-quit:
			return
		}
	}

}

// Render performs the rendering task spread across 1 or more worker goroutines.
func Render(cam *camera.Camera, world *hitable.HitableSlice, canvas *image.NRGBA, numSamples int, numWorkers int) {
	nx := canvas.Bounds().Max.X
	ny := canvas.Bounds().Max.Y

	queue := make(chan workUnit, numWorkers)
	quit := make(chan struct{})
	wg := sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		go worker(queue, quit, wg)
	}

	for y := 0; y <= ny; y += 10 {
		queue <- workUnit{
			cam:        cam,
			world:      world,
			canvas:     canvas,
			numSamples: numSamples,
			x0:         0,
			x1:         nx,
			y0:         y,
			y1:         y + 10,
		}
	}

	for i := 0; i < numWorkers; i++ {
		quit <- struct{}{}
	}

	wg.Wait()
}
