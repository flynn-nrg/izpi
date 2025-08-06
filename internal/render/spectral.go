package render

import (
	"math/rand/v2"
	"sync"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/spectral"
)

func renderRectSpectral(w workUnit, random *fastrandom.LCG) {
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
			col := spectral.NewEmptyCIESPD()
			for s := 0; s < w.numSamples; s++ {
				// Choose a wavelength.
				samplingIndex := int(float64(col.NumWavelengths()) * rand.Float64())
				lambda := col.Wavelength(samplingIndex)
				u := (float64(x) + rand.Float64()) / float64(nx)
				v := (float64(y) + rand.Float64()) / float64(ny)
				r := w.scene.Camera.GetRayWithLambda(u, v, lambda)
				sampled := w.sampler.SampleSpectral(r, w.scene.World, w.scene.Lights, 0, random)
				col.AddValue(samplingIndex, sampled)
			}

			// Normalise the spectral power distribution
			col.Normalise(w.numSamples)
			// Convert to RGB.
			r, g, b := spectral.SPDToRGB(col, 20.0)
			w.canvas.Set(x, ny-y, colour.FloatNRGBA{R: r, G: g, B: b, A: 1.0})
			if w.preview {
				tile.Pixels[i] = b
				tile.Pixels[i+1] = g
				tile.Pixels[i+2] = r
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

func workerSpectral(input chan workUnit, quit chan struct{}, random *fastrandom.LCG, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case w := <-input:
			renderRectSpectral(w, random)
		case <-quit:
			return
		}
	}
}
