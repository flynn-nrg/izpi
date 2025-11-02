// Package render implements the main rendering loop.
package render

import (
	"context"
	"image"
	"sync"
	"time"

	"github.com/flynn-nrg/floatimage/floatimage"
	"github.com/flynn-nrg/izpi/internal/common"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/grid"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/spectral"
	"github.com/flynn-nrg/izpi/internal/vec3"

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"

	pb "github.com/cheggaaa/pb/v3"
	log "github.com/sirupsen/logrus"
)

type Renderer interface {
	Render(ctx context.Context) image.Image
}

// RendererImpl represents an RGB renderer config.
type RendererImpl struct {
	scene              *scene.Scene
	numRays            uint64
	remoteWorkers      []*RemoteWorkerConfig
	canvas             *floatimage.Float64NRGBA
	previewChan        chan display.DisplayTile
	maxDepth           int
	background         vec3.Vec3Impl
	ink                vec3.Vec3Impl
	spectralBackground *spectral.SpectralPowerDistribution
	exposure           float64
	preview            bool
	samplerType        sampler.SamplerType
	sizeX              int
	sizeY              int
	numSamples         int
	numWorkers         int
	verbose            bool
}

type RemoteWorkerConfig struct {
	Client   pb_control.RenderControlServiceClient
	NumCores int
}

type workUnit struct {
	scene       *scene.Scene
	canvas      *floatimage.Float64NRGBA
	bar         *pb.ProgressBar
	sampler     sampler.Sampler
	previewChan chan display.DisplayTile
	preview     bool
	verbose     bool
	stripHeight int
	numSamples  int
	x0          int
	x1          int
	y0          int
	y1          int
}

// New returns a new instance of a renderer.
func New(
	scene *scene.Scene,
	sizeX int, sizeY int,
	numSamples int, maxDepth int,
	background vec3.Vec3Impl, ink vec3.Vec3Impl,
	spectralBackground *spectral.SpectralPowerDistribution,
	numLocalWorkers int,
	remoteWorkers []*RemoteWorkerConfig,
	verbose bool,
	previewChan chan display.DisplayTile,
	preview bool,
	samplerType sampler.SamplerType,
) *RendererImpl {
	return &RendererImpl{
		scene:              scene,
		remoteWorkers:      remoteWorkers,
		canvas:             floatimage.NewFloat64NRGBA(image.Rect(0, 0, sizeX, sizeY), make([]float64, sizeX*sizeY*4)),
		previewChan:        previewChan,
		maxDepth:           maxDepth,
		background:         background,
		ink:                ink,
		spectralBackground: spectralBackground,
		exposure:           scene.Exposure,
		preview:            preview,
		samplerType:        samplerType,
		sizeX:              sizeX,
		sizeY:              sizeY,
		numSamples:         numSamples,
		numWorkers:         numLocalWorkers,
		verbose:            verbose,
	}
}

// Render performs the rendering task spread across 1 or more worker goroutines.
// It returns a FloatNRGBA image that can be further processed before output or fed to an output directly.
func (r *RendererImpl) Render(ctx context.Context) image.Image {

	var bar *pb.ProgressBar

	queue := make(chan workUnit)
	quit := make(chan struct{})
	wg := &sync.WaitGroup{}

	stepSizeX, stepSizeY := common.Tiles(r.sizeX, r.sizeY)

	numTiles := (r.sizeX / stepSizeX) * (r.sizeY / stepSizeY)
	if r.verbose {
		bar = pb.StartNew(numTiles)
	}

	totalWorkers := 0

	// Local workers
	for range r.numWorkers {
		totalWorkers++
		random := fastrandom.NewWithDefaults()
		wg.Add(1)
		switch r.samplerType {
		case sampler.ColourSampler, sampler.NormalSampler, sampler.WireFrameSampler, sampler.AlbedoSampler:
			go workerRGB(queue, quit, random, wg)
		case sampler.SpectralSampler:
			go workerSpectral(queue, quit, random, wg)
		default:
			log.Fatalf("invalid sampler type %v", r.samplerType)
		}
	}

	// Remote workers
	for _, worker := range r.remoteWorkers {
		for range worker.NumCores {
			totalWorkers++
			wg.Add(1)
			go remoteWorker(ctx, queue, quit, wg, worker)
		}
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
	case sampler.SpectralSampler:
		s = sampler.NewSpectral(r.maxDepth, r.spectralBackground, &r.numRays)
	default:
		log.Fatalf("invalid sampler type %v", r.samplerType)
	}

	log.Infof("Begin rendering using %v worker threads: %v local, %v remote", totalWorkers, r.numWorkers, totalWorkers-r.numWorkers)
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
			stripHeight: 1,
			numSamples:  r.numSamples,
			x0:          t.X * stepSizeX,
			x1:          t.X*stepSizeX + (stepSizeX - 1),
			y0:          t.Y * stepSizeY,
			y1:          t.Y*stepSizeY + (stepSizeY - 1),
		}
	}

	for range totalWorkers {
		quit <- struct{}{}
	}

	log.Debugf("Rendering done. Waiting for workers threads to exit")

	wg.Wait()

	if r.verbose {
		bar.Finish()
	}

	// If there are any remote workers, call them to collect stats and free up resources.
	for _, worker := range r.remoteWorkers {
		report, err := worker.Client.RenderEnd(ctx, &pb_control.RenderEndRequest{})
		if err != nil {
			log.Errorf("Failed to get render end report: %v", err)
		}

		workerNumRays := report.GetTotalRaysTraced()
		r.numRays += workerNumRays
	}

	log.Infof("Rendering completed in %v using %v rays", time.Since(startTime), r.numRays)

	// If spectral rendering is enabled, perform firefly rejection and convert to sRGB.
	return r.canvas
}
