// Package pdf implements methods to work with probability density functions.
package pdf

import "gitlab.com/flynn-nrg/izpi/pkg/vec3"

// PDF represents a probability density function.
type PDF interface {
	// Value computes the probability density function at a given point.
	Value(direction *vec3.Vec3Impl) float64
	// Generate generates a probability density function.
	Generate() *vec3.Vec3Impl
}
