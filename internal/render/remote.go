package render

import (
	"context"
	"io"
	"sync"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/izpi/internal/display"

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	log "github.com/sirupsen/logrus"
)

func renderRectRemote(ctx context.Context, w workUnit, client pb_control.RenderControlServiceClient) {
	var tile display.DisplayTile

	ny := w.canvas.Bounds().Max.Y

	if w.preview {
		tile = display.DisplayTile{
			Width:  w.x1 - w.x0 + 1,
			Height: 1,
			PosX:   w.x0,
			Pixels: make([]float64, (w.x1-w.x0+1)*4),
		}
	}

	request := &pb_control.RenderTileRequest{
		StripHeight: 1,
		X0:          uint32(w.x0),
		Y0:          uint32(w.y0),
		X1:          uint32(w.x1),
		Y1:          uint32(w.y1),
	}

	stream, err := client.RenderTile(ctx, request)
	if err != nil {
		log.Errorf("Failed to render tile: %v", err)
		return
	}
	defer stream.CloseSend()

	for {
		reply, err := stream.Recv()
		if err != nil {
			if err == io.EOF {
				break
			}

			log.Errorf("Failed to receive tile: %v", err)
			return
		}

		posX := int(reply.GetPosX())
		posY := int(reply.GetPosY())
		width := int(reply.GetWidth())
		pixels := reply.GetPixels()

		tile.PosY = ny - posY

		i := 0
		for x := posX; x < posX+width; x++ {
			w.canvas.Set(x, ny-posY, colour.FloatNRGBA{B: pixels[i], G: pixels[i+1], R: pixels[i+2], A: pixels[i+3]})
			if w.preview {
				tile.Pixels[i] = pixels[i]
				tile.Pixels[i+1] = pixels[i+1]
				tile.Pixels[i+2] = pixels[i+2]
				tile.Pixels[i+3] = pixels[i+3]
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

func remoteWorker(ctx context.Context, input chan workUnit, quit chan struct{}, wg *sync.WaitGroup, config *RemoteWorkerConfig) {
	defer wg.Done()
	for {
		select {
		case w := <-input:
			renderRectRemote(ctx, w, config.Client)
		case <-quit:
			log.Debug("Remote worker exiting")
			return
		}
	}
}
