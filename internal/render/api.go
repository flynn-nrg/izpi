package render

import (
	"context"
	"image"
)

type Renderer interface {
	Render(ctx context.Context) image.Image
}
