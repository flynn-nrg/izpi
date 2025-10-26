// Package texture implements different types of textures.
package texture

import "github.com/flynn-nrg/go-vfx/math32/vec3"

// UV represents a UV pair.
type UV struct {
	U float32
	V float32
}

// Texture represents a texture.
type Texture interface {
	// Value returns the color values at a given point.
	Value(u float32, v float32, p vec3.Vec3Impl) vec3.Vec3Impl
}

// SpectralTexture represents a spectral texture.
type SpectralTexture interface {
	Value(u float32, v float32, lambda float32, p vec3.Vec3Impl) float32
}
