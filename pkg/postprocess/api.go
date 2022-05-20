// Package postprocess implements the postprocess pipeline.
package postprocess

import (
	"image"

	"github.com/flynn-nrg/izpi/pkg/scene"
)

// Filter represents a postprocess filter.
type Filter interface {
	// Apply performs a series of changes on the supplied image.
	Apply(i image.Image, scene *scene.Scene) error
}
