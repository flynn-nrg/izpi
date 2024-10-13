// Package display implements an Fyne-backed preview window.
package display

import (
	"image"
	"image/color"
	"math"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"

	"github.com/flynn-nrg/izpi/pkg/common"
)

var _ Display = (*FyneDisplay)(nil)

type FyneDisplay struct {
	name   string
	window fyne.Window
	canvas *canvas.Image
	image  *image.RGBA
	width  int
	height int
	input  chan DisplayTile
	quit   chan struct{}
	wg     sync.WaitGroup
}

func NewFyneDisplay(name string, width int, height int, input chan DisplayTile) *FyneDisplay {
	myApp := app.New()
	window := myApp.NewWindow("Izpi render output")
	window.Resize(fyne.NewSize(float32(width), float32(height)))
	window.SetMaster()
	window.SetFixedSize(true)

	backdrop := image.NewRGBA(image.Rectangle{
		Min: image.Point{
			X: 0,
			Y: 0,
		},
		Max: image.Point{
			X: width,
			Y: height,
		},
	})

	c := canvas.NewImageFromImage(backdrop)

	window.SetContent(container.NewStack(c))

	return &FyneDisplay{
		name:   name,
		window: window,
		image:  backdrop,
		canvas: c,
		width:  width,
		height: height,
		input:  input,
		quit:   make(chan struct{}),
	}
}

func (fd *FyneDisplay) Start() {
	go fd.busyLoop()

	fd.window.ShowAndRun()

	fd.quit <- struct{}{}
	fd.wg.Wait()
}

func (fd *FyneDisplay) Wait() {
	fd.wg.Wait()
}

func (fd *FyneDisplay) busyLoop() {
	fd.wg.Add(1)
	defer fd.wg.Done()

	fd.makeBackdrop()

	for {
		select {
		case in := <-fd.input:
			i := 0
			for y := in.PosY; y < in.PosY+in.Height; y++ {
				for x := in.PosX; x < in.PosX+in.Width; x++ {
					b := floatToUint8(in.Pixels[i])
					i++
					g := floatToUint8(in.Pixels[i])
					i++
					r := floatToUint8(in.Pixels[i])
					i++
					a := floatToUint8(in.Pixels[i])
					i++

					pixel := color.RGBA{
						R: r,
						G: g,
						B: b,
						A: a,
					}

					fd.image.Set(x, y, pixel)
				}
			}

			fd.canvas.Refresh()
		case <-fd.quit:
			return
		}
	}
}

func (fd *FyneDisplay) makeBackdrop() {
	cols := []color.Color{
		color.RGBA{
			R: 255,
			G: 255,
			B: 255,
			A: 255,
		},
		color.RGBA{
			R: 128,
			G: 128,
			B: 128,
			A: 255,
		},
	}

	chosen := 0
	stepSizeX, stepSizeY := common.Tiles(int(fd.width), int(fd.height))

	for y := 0; y < fd.height; y++ {
		for x := 0; x < fd.width; x++ {
			if x%stepSizeX == 0 {
				chosen ^= 1
			}
			// RGBA
			fd.image.Set(x, y, cols[chosen])
		}
		if y%stepSizeY == 0 {
			chosen ^= 1
		}
	}
}

func floatToUint8(in float64) uint8 {
	// Gamma 2.0
	in = math.Sqrt(in)
	p := int(in * 255)
	if p > 255 {
		return 255
	}

	return uint8(p)
}
