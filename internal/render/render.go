// Package render implements the main rendering loop.
package render

import (
	"image"
	"math/rand"
	"sync"
	"time"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/izpi/internal/common"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/grid"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/vec3"

	pb "github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
)

// Renderer represents a renderer config.
type Renderer struct {
	scene       *scene.Scene
	numRays     uint64
	canvas      *floatimage.FloatNRGBA
	previewChan chan display.DisplayTile
	maxDepth    int
	background  *vec3.Vec3Impl
	ink         *vec3.Vec3Impl
	preview     bool
	samplerType sampler.SamplerType
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
	sampler     sampler.Sampler
	previewChan chan display.DisplayTile
	preview     bool
	verbose     bool
	numSamples  int
	x0          int
	x1          int
	y0          int
	y1          int
}

func renderRect(w workUnit, random *fastrandom.LCG) {
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
				col = vec3.Add(col, vec3.DeNAN(w.sampler.Sample(r, w.scene.World, w.scene.Lights, 0, random)))
			}

			// Linear colour space.
			col = vec3.ScalarDiv(col, float64(w.numSamples))
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

func worker(input chan workUnit, quit chan struct{}, random *fastrandom.LCG, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case w := <-input:
			renderRect(w, random)
		case <-quit:
			return
		}
	}

}

// New returns a new instance of a renderer.
func New(scene *scene.Scene, sizeX int, sizeY int, numSamples int, maxDepth int, background *vec3.Vec3Impl, ink *vec3.Vec3Impl,
	numWorkers int, verbose bool, previewChan chan display.DisplayTile, preview bool, samplerType sampler.SamplerType) *Renderer {
	return &Renderer{
		scene:       scene,
		canvas:      floatimage.NewFloatNRGBA(image.Rect(0, 0, sizeX, sizeY), make([]float64, sizeX*sizeY*4)),
		previewChan: previewChan,
		maxDepth:    maxDepth,
		background:  background,
		ink:         ink,
		preview:     preview,
		samplerType: samplerType,
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

	var bar *pb.ProgressBar

	queue := make(chan workUnit)
	quit := make(chan struct{})
	wg := &sync.WaitGroup{}

	stepSizeX, stepSizeY := common.Tiles(r.sizeX, r.sizeY)

	numTiles := (r.sizeX / stepSizeX) * (r.sizeY / stepSizeY)
	if r.verbose {
		bar = pb.StartNew(numTiles)
	}

	for i := 0; i < r.numWorkers; i++ {
		random := fastrandom.NewWithDefaults()
		wg.Add(1)
		go worker(queue, quit, random, wg)
	}

	gridSizeX := r.sizeX / stepSizeX
	gridSizeY := r.sizeY / stepSizeY
	path := grid.WalkGrid(gridSizeX, gridSizeY, grid.PATTERN_SPIRAL)

	var s sampler.Sampler
	switch r.samplerType {
	case sampler.ColourSampler:
		s = sampler.NewColour(r.maxDepth, r.background, &r.numRays)
	case sampler.NormalSampler:
		s = sampler.NewNormal(&r.numRays)
	case sampler.WireFrameSampler:
		s = sampler.NewWireFrame(r.background, r.ink, &r.numRays)
	case sampler.AlbedoSampler:
		s = sampler.NewAlbedo(&r.numRays)
	default:
		log.Fatalf("invalid sampler type %v", r.samplerType)
	}

	log.Infof("Begin rendering using %v worker threads", r.numWorkers)
	startTime := time.Now()

	for _, t := range path {
		queue <- workUnit{
			scene:       r.scene,
			canvas:      r.canvas,
			bar:         bar,
			sampler:     s,
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

	log.Infof("Rendering completed in %v using %v rays", time.Since(startTime), r.numRays)
	return r.canvas
}
