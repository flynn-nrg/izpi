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
	"gitlab.com/flynn-nrg/izpi/pkg/pdf"
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

func colour(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		_, srec, ok := mat.Scatter(r, rec)
		emitted := mat.Emitted(r, rec, rec.U(), rec.V(), rec.P())
		if depth < 50 && ok {
			if srec.IsSpecular() {
				// srec.Attenuation() * colour(...)
				return vec3.Mul(srec.Attenuation(), colour(srec.SpecularRay(), world, lightShape, depth+1))
			} else {
				pLight := pdf.NewHitable(lightShape, rec.P())
				p := pdf.NewMixture(pLight, srec.PDF())
				scattered := ray.New(rec.P(), p.Generate(), r.Time())
				pdfVal := p.Value(scattered.Direction())
				// emitted + (albedo * scatteringPDF())*colour() / pdf
				v1 := vec3.ScalarMul(colour(scattered, world, lightShape, depth+1), mat.ScatteringPDF(r, rec, scattered))
				v2 := vec3.Mul(srec.Attenuation(), v1)
				v3 := vec3.ScalarDiv(v2, pdfVal)
				res := vec3.Add(emitted, v3)
				return res
			}
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
				lightShape := hitable.NewXZRect(213, 343, 227, 332, 554, nil)
				glassSphere := hitable.NewSphere(&vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, &vec3.Vec3Impl{X: 190, Y: 90, Z: 190}, 0, 1, 90, nil)
				hList := hitable.NewSlice([]hitable.Hitable{lightShape, glassSphere})
				col = vec3.Add(col, vec3.DeNAN(colour(r, w.world, hList, 0)))
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
