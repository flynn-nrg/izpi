// Package postprocess implements the postprocess pipeline.
package postprocess

import (
	"image"

	"gitlab.com/flynn-nrg/izpi/pkg/camera"
)

// Filter represents a postprocess filter.
type Filter interface {
	// Apply performs a series of changes on the supplied image.
	Apply(i image.Image, cam *camera.Camera) error
}