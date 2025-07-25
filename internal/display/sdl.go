// Package display implements an SDL-backed preview window.
// The way SDL is shoehorned into Go makes for some intereasting limitations in regards to
// how multithreading is managed. This will probably be replaced by something Vulkan-backed
// in the future.
// nolint
// TODO: Reimplement this using fyne.
package display

import (
	"math"
	"sync"
	"time"

	"github.com/flynn-nrg/izpi/internal/common"
	log "github.com/sirupsen/logrus"
	"github.com/veandco/go-sdl2/sdl"
)

var _ Display = (*SDLDisplay)(nil)

type SDLDisplay struct {
	name     string
	width    int32
	height   int32
	pitch    int32
	window   *sdl.Window
	renderer *sdl.Renderer
	texture  *sdl.Texture
	input    chan DisplayTile
	quit     chan struct{}
	wg       sync.WaitGroup
}

func NewSDLDisplay(name string, width int, height int, input chan DisplayTile) *SDLDisplay {
	return &SDLDisplay{
		name:   name,
		width:  int32(width),
		height: int32(height),
		input:  input,
		quit:   make(chan struct{}),
	}
}

func (sd *SDLDisplay) Start() {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		log.Fatal(err)
	}

	sdl.Main(func() {
		var window *sdl.Window
		var renderer *sdl.Renderer
		var texture *sdl.Texture
		var err error

		sdl.Do(func() {
			window, err = sdl.CreateWindow(sd.name, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, sd.width, sd.height, sdl.WINDOW_OPENGL)
			if err != nil {
				log.Fatal(err)
			}
		})
		sd.window = window

		sdl.Do(func() {
			renderer, err = sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED)
			if err != nil {
				log.Fatal(err)
			}
		})

		sd.renderer = renderer

		var surface *sdl.Surface

		sdl.Do(func() {
			surface, err = sdl.CreateRGBSurfaceWithFormat(0, sd.width, sd.height, 32, sdl.PIXELFORMAT_RGBA8888)
			if err != nil {
				log.Fatal(err)
			}
		})

		sd.pitch = surface.Pitch

		sdl.Do(func() {
			texture, err = sd.renderer.CreateTextureFromSurface(surface)
			if err != nil {
				log.Fatal(err)
			}
		})

		sd.texture = texture
		go sd.busyLoop()
		sd.poll()
		sd.quit <- struct{}{}
		sd.wg.Wait()
		sdl.Do(func() {
			err := sd.texture.Destroy()
			if err != nil {
				log.Error(err)
			}
			err = sd.renderer.Destroy()
			if err != nil {
				log.Error(err)
			}
			err = sd.window.Destroy()
			if err != nil {
				log.Error(err)
			}
		})
	})
}

func (sd *SDLDisplay) Wait() {
	sd.wg.Wait()
}

func (sd *SDLDisplay) busyLoop() {
	sd.wg.Add(1)
	defer sd.wg.Done()

	sd.makeBackdrop()

	for {
		select {
		case in := <-sd.input:
			rect := &sdl.Rect{
				X: int32(in.PosX),
				Y: int32(in.PosY),
				W: int32(in.Width),
				H: int32(in.Height),
			}

			pixels := make([]uint32, len(in.Pixels)/4)
			for i := 0; i < len(in.Pixels); i += 4 {
				pixels[i/4] = floatToUint32(in.Pixels[i]) | (floatToUint32(in.Pixels[i+1]) << 8) | (floatToUint32(in.Pixels[i+2]) << 16) | (floatToUint32(in.Pixels[i+3]) << 24)
			}

			err := sd.texture.UpdateRGBA(rect, pixels, int(sd.pitch))
			if err != nil {
				log.Error(err)
			}
			err = sd.renderer.Copy(sd.texture, nil, nil)
			if err != nil {
				log.Error(err)
			}
			sd.renderer.Present()

		case <-sd.quit:
			return
		}
	}
}

func (sd *SDLDisplay) poll() {
	sdl.Do(func() {
		for {
			for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
				switch event.(type) {
				case *sdl.QuitEvent:
					return
				}
			}
			time.Sleep(100 * time.Millisecond)
		}
	})
}

func (sd *SDLDisplay) makeBackdrop() {
	canvas := make([]uint32, sd.height*4*sd.width)

	cols := []uint32{
		255 << 24,
		(128) | (128 << 8) | (128 << 16) | (255 << 24),
	}

	chosen := 0
	stepSizeX, stepSizeY := common.Tiles(int(sd.width), int(sd.height))
	stepSizeY *= 4

	i := 0

	for y := 0; y < int(sd.height)*4; y++ {
		for x := 0; x < int(sd.width); x++ {
			if x%stepSizeX == 0 {
				chosen ^= 1
			}
			// RGBA
			canvas[i] = cols[chosen]
			i++
		}
		if y%stepSizeY == 0 {
			chosen ^= 1
		}
	}

	rect := &sdl.Rect{
		X: 0,
		Y: 0,
		W: sd.width,
		H: sd.height,
	}

	err := sd.texture.UpdateRGBA(rect, canvas, int(sd.pitch))
	if err != nil {
		log.Error(err)
	}
	err = sd.renderer.Copy(sd.texture, nil, nil)
	if err != nil {
		log.Error(err)
	}
	sd.renderer.Present()

}

func floatToUint32(in float64) uint32 {
	// Gamma 2.0
	in = math.Sqrt(in)
	p := int(in * 255)
	if p > 255 {
		return 255
	}

	return uint32(p)
}
