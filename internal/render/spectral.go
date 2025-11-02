package render

import (
	"sync"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
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
			cieX, cieY, cieZ := RenderPixelSpectral(w.numSamples, x, y, nx, ny, w.scene, w.sampler, random)

			// Canvas information is in CIE XYZ space.
			w.canvas.Set(x, ny-y, colour.Float64NRGBA{R: cieX, G: cieY, B: cieZ, A: 1.0})

			exposure := w.scene.Exposure
			r, g, b := w.scene.WhiteBalance.Matrix.Apply(cieX*exposure, cieY*exposure, cieZ*exposure)

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

func RenderPixelSpectral(numSamples int, x, y, nx, ny int, scene *scene.Scene, sampler sampler.Sampler, random *fastrandom.LCG) (float64, float64, float64) {
	// Initialize XYZ accumulators for the pixel
	var sumX, sumY, sumZ float64

	for range numSamples {
		// Importance sample a wavelength AND its PDF
		lambda, pdf := spectral.SampleWavelength(random.Float64())
		if pdf == 0 {
			continue
		}

		// Get camera ray for this specific wavelength
		u := (float64(x) + random.Float64()) / float64(nx)
		v := (float64(y) + random.Float64()) / float64(ny)
		r := scene.Camera.GetRayWithLambda(u, v, lambda)

		// Trace the path to get radiance at this wavelength
		radiance := sampler.SampleSpectral(r, scene.World, scene.Lights, 0, random)

		// Convert sample to an XYZ contribution and add it to the pixel's
		// accumulators using the unbiased estimator.
		cieX_val, cieY_val, cieZ_val := spectral.GetCIEValues(lambda)

		sumX += (radiance * cieX_val) / pdf
		sumY += (radiance * cieY_val) / pdf
		sumZ += (radiance * cieZ_val) / pdf
	}

	// Average the accumulated XYZ values
	invNumSamples := 1.0 / float64(numSamples)
	finalX := sumX * invNumSamples
	finalY := sumY * invNumSamples
	finalZ := sumZ * invNumSamples

	// Convert the final XYZ color to linear sRGB using the white balance matrix
	//exposure := 1.0
	//r, g, b := scene.WhiteBalance.Matrix.Apply(finalX*exposure, finalY*exposure, finalZ*exposure)

	return finalX, finalY, finalZ
}
