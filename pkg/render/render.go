// Package render implements the main rendering loop.
package render

import (
	"image"
	"math"
	"math/rand"
	"sync"

	"github.com/flynn-nrg/izpi/pkg/colour"
	"github.com/flynn-nrg/izpi/pkg/display"
	"github.com/flynn-nrg/izpi/pkg/floatimage"
	"github.com/flynn-nrg/izpi/pkg/grid"
	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/pdf"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/scene"
	"github.com/flynn-nrg/izpi/pkg/vec3"

	pb "github.com/cheggaaa/pb/v3"
)

// Renderer represents a renderer config.
type Renderer struct {
	scene       *scene.Scene
	canvas      *floatimage.FloatNRGBA
	previewChan chan display.DisplayTile
	preview     bool
	sizeX       int
	sizeY       int
	numSamples  int
	numWorkers  int
	verbose     bool
}

type workUnit struct {
	scene       *scene.Scene
	canvas      *floatimage.FloatNRGBA
	bar         *pb.ProgressBar
	previewChan chan display.DisplayTile
	preview     bool
	verbose     bool
	numSamples  int
	x0          int
	x1          int
	y0          int
	y1          int
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
	var tile display.DisplayTile

	nx := w.canvas.Bounds().Max.X
	ny := w.canvas.Bounds().Max.Y

	if w.preview {
		tile = display.DisplayTile{
			Width:  w.x1 - w.x0 + 1,
			Height: 1,
			PosX:   w.x0,
			Pixels: make([]float64, (w.x1-w.x0+1)*4),
		}
	}

	for y := w.y0; y <= w.y1; y++ {
		i := 0
		tile.PosY = ny - y
		for x := w.x0; x <= w.x1; x++ {
			col := &vec3.Vec3Impl{}
			for s := 0; s < w.numSamples; s++ {
				u := (float64(x) + rand.Float64()) / float64(nx)
				v := (float64(y) + rand.Float64()) / float64(ny)
				r := w.scene.Camera.GetRay(u, v)
				col = vec3.Add(col, vec3.DeNAN(computeColour(r, w.scene.World, w.scene.Lights, 0)))
			}

			col = vec3.ScalarDiv(col, float64(w.numSamples))
			// gamma 2
			col = &vec3.Vec3Impl{X: math.Sqrt(col.X), Y: math.Sqrt(col.Y), Z: math.Sqrt(col.Z)}
			w.canvas.Set(x, ny-y, colour.FloatNRGBA{R: col.X, G: col.Y, B: col.Z, A: 1.0})
			if w.preview {
				tile.Pixels[i] = col.Z
				tile.Pixels[i+1] = col.Y
				tile.Pixels[i+2] = col.X
				tile.Pixels[i+3] = 1.0
				i += 4
			}
		}
		if w.preview {
			w.previewChan <- tile
		}
	}
	if w.verbose {
		w.bar.Increment()
	}
}

func worker(input chan workUnit, quit chan struct{}, wg *sync.WaitGroup) {
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
func New(scene *scene.Scene, sizeX int, sizeY int, numSamples int,
	numWorkers int, verbose bool, previewChan chan display.DisplayTile, preview bool) *Renderer {
	return &Renderer{
		scene:       scene,
		canvas:      floatimage.NewFloatNRGBA(image.Rect(0, 0, sizeX, sizeY)),
		previewChan: previewChan,
		preview:     preview,
		sizeX:       sizeX,
		sizeY:       sizeY,
		numSamples:  numSamples,
		numWorkers:  numWorkers,
		verbose:     verbose,
	}
}

// Render performs the rendering task spread across 1 or more worker goroutines.
// It returns a FloatNRGBA image that can be further processed before output or fed to an output directly.
func (r *Renderer) Render() image.Image {
	stepSizes := []int{32, 25, 24, 20, 16, 12, 10, 8, 5, 4}

	var bar *pb.ProgressBar
	var stepSizeX, stepSizeY int

	queue := make(chan workUnit)
	quit := make(chan struct{})
	wg := &sync.WaitGroup{}

	for _, size := range stepSizes {
		if r.sizeX%size == 0 {
			stepSizeX = size
			break
		}
	}

	for _, size := range stepSizes {
		if r.sizeY%size == 0 {
			stepSizeY = size
			break
		}
	}

	numTiles := (r.sizeX / stepSizeX) * (r.sizeY / stepSizeY)
	if r.verbose {
		bar = pb.StartNew(numTiles)
	}

	for i := 0; i < r.numWorkers; i++ {
		go worker(queue, quit, wg)
	}

	gridSizeX := r.sizeX / stepSizeX
	gridSizeY := r.sizeY / stepSizeY
	path := grid.WalkGrid(gridSizeX, gridSizeY, grid.PATTERN_SPIRAL)
	for _, t := range path {
		queue <- workUnit{
			scene:       r.scene,
			canvas:      r.canvas,
			bar:         bar,
			previewChan: r.previewChan,
			preview:     r.preview,
			verbose:     r.verbose,
			numSamples:  r.numSamples,
			x0:          t.X * stepSizeX,
			x1:          t.X*stepSizeX + (stepSizeX - 1),
			y0:          t.Y * stepSizeY,
			y1:          t.Y*stepSizeY + (stepSizeY - 1),
		}
	}
	for i := 0; i < r.numWorkers; i++ {
		quit <- struct{}{}
	}

	wg.Wait()

	if r.verbose {
		bar.Finish()
	}

	return r.canvas
}
