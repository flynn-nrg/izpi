// Package display implements an SDL-backed preview window.
// The way SDL is shoehorned into Go makes for some intereasting limitations in regards to
// how multithreading is managed. This will probably be replaced by something Vulkan-backed
// in the future.
package display

import (
	"log"
	"sync"
	"time"

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
			sd.texture.Destroy()
			sd.renderer.Destroy()
			sd.window.Destroy()
		})
	})
}

func (sd *SDLDisplay) Wait() {
	sd.wg.Wait()
}

func (sd *SDLDisplay) busyLoop() {
	sd.wg.Add(1)
	defer sd.wg.Done()
	for {
		select {
		case in := <-sd.input:
			rect := &sdl.Rect{
				X: int32(in.PosX),
				Y: int32(in.PosY),
				W: int32(in.Width),
				H: int32(in.Height),
			}

			pixels := make([]byte, len(in.Pixels)*4)
			for i := 0; i < len(in.Pixels); i++ {
				pixels[i] = floatToByte(in.Pixels[i])
			}

			sd.texture.Update(rect, pixels, int(sd.pitch))
			sd.renderer.Copy(sd.texture, nil, nil)
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

func floatToByte(in float64) byte {
	p := int(in * 255)
	if p > 255 {
		return 255
	}

	return byte(p)
}
