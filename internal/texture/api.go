// Package texture implements different types of textures.
package texture

import "github.com/flynn-nrg/izpi/internal/vec3"

// Texture represents a texture.
type Texture interface {
	// Value returns the color values at a given point.
	Value(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl
}

// UV represents a UV pair.
type UV struct {
	U float64
	V float64
}
