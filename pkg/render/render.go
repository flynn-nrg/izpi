// Package render implements the main rendering loop.
package render

import (
	"image"
	"math"
	"math/rand"
	"sync"

	"gitlab.com/flynn-nrg/izpi/pkg/camera"
	"gitlab.com/flynn-nrg/izpi/pkg/colour"
	"gitlab.com/flynn-nrg/izpi/pkg/floatimage"
	"gitlab.com/flynn-nrg/izpi/pkg/hitable"
	"gitlab.com/flynn-nrg/izpi/pkg/pdf"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"

	pb "github.com/cheggaaa/pb/v3"
)

// Renderer represents a renderer config.
type Renderer struct {
	cam        *camera.Camera
	world      *hitable.HitableSlice
	canvas     *floatimage.FloatNRGBA
	sizeX      int
	sizeY      int
	numSamples int
	numWorkers int
	stepSize   int
	verbose    bool
}

type workUnit struct {
	cam        *camera.Camera
	world      *hitable.HitableSlice
	canvas     *floatimage.FloatNRGBA
	numSamples int
	x0         int
	x1         int
	y0         int
	y1         int
}

func computeColour(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		_, srec, ok := mat.Scatter(r, rec)
		emitted := mat.Emitted(r, rec, rec.U(), rec.V(), rec.P())
		if depth < 50 && ok {
			if srec.IsSpecular() {
				// srec.Attenuation() * colour(...)
				return vec3.Mul(srec.Attenuation(), computeColour(srec.SpecularRay(), world, lightShape, depth+1))
			} else {
				pLight := pdf.NewHitable(lightShape, rec.P())
				p := pdf.NewMixture(pLight, srec.PDF())
				scattered := ray.New(rec.P(), p.Generate(), r.Time())
				pdfVal := p.Value(scattered.Direction())
				// emitted + (albedo * scatteringPDF())*colour() / pdf
				v1 := vec3.ScalarMul(computeColour(scattered, world, lightShape, depth+1), mat.ScatteringPDF(r, rec, scattered))
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
				col = vec3.Add(col, vec3.DeNAN(computeColour(r, w.world, hList, 0)))
			}

			col = vec3.ScalarDiv(col, float64(w.numSamples))
			// gamma 2
			col = &vec3.Vec3Impl{X: math.Sqrt(col.X), Y: math.Sqrt(col.Y), Z: math.Sqrt(col.Z)}
			w.canvas.Set(x, ny-y, colour.FloatNRGBA{R: col.X, G: col.Y, B: col.Z, A: 1.0})
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

// New returns a new instance of a renderer.
func New(cam *camera.Camera, world *hitable.HitableSlice, sizeX int, sizeY int, numSamples int, numWorkers int, stepSize int, verbose bool) *Renderer {
	return &Renderer{
		cam:        cam,
		world:      world,
		canvas:     floatimage.NewFloatNRGBA(image.Rect(0, 0, sizeX, sizeY)),
		sizeX:      sizeX,
		sizeY:      sizeY,
		numSamples: numSamples,
		numWorkers: numWorkers,
		stepSize:   stepSize,
		verbose:    verbose,
	}
}

// Render performs the rendering task spread across 1 or more worker goroutines.
// It returns a FloatNRGBA image that can be further processed before output or fed to an output directly.
func (r *Renderer) Render() image.Image {
	queue := make(chan workUnit, r.numWorkers)
	quit := make(chan struct{})
	wg := sync.WaitGroup{}

	bar := pb.StartNew(r.sizeY)

	for i := 0; i < r.numWorkers; i++ {
		go worker(queue, quit, wg)
	}

	for y := 0; y < r.sizeY; y += r.stepSize {
		if r.verbose {
			bar.Add(r.stepSize)
		}
		queue <- workUnit{
			cam:        r.cam,
			world:      r.world,
			canvas:     r.canvas,
			numSamples: r.numSamples,
			x0:         0,
			x1:         r.sizeX,
			y0:         y,
			y1:         y + r.stepSize,
		}
	}

	for i := 0; i < r.numWorkers; i++ {
		quit <- struct{}{}
	}

	wg.Wait()
	bar.Finish()

	return r.canvas
}
