package render

import (
	"math/rand/v2"
	"sync"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

func renderRectRGB(w workUnit, random *fastrandom.LCG) {
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
			w.canvas.Set(x, ny-y, colour.Float64NRGBA{R: col.X, G: col.Y, B: col.Z, A: 1.0})
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

func workerRGB(input chan workUnit, quit chan struct{}, random *fastrandom.LCG, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case w := <-input:
			renderRectRGB(w, random)
		case <-quit:
			return
		}
	}

}
