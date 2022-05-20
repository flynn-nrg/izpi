// Package ray implements the interface and methods to work with rays.
package ray

import "github.com/flynn-nrg/izpi/pkg/vec3"

// Ray defines the methods used to work with rays.
type Ray interface {
	Origin() *vec3.Vec3Impl
	Direction() *vec3.Vec3Impl
	PointAtParameter(t float64) *vec3.Vec3Impl
	Time() float64
}
