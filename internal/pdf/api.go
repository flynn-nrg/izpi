// Package pdf implements methods to work with probability density functions.
package pdf

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// PDF represents a probability density function.
type PDF interface {
	// Value computes the probability density function at a given point.
	Value(direction *vec3.Vec3Impl) float32
	// Generate generates a probability density function.
	Generate(random *fastrandom.XorShift) *vec3.Vec3Impl
}
