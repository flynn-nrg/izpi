// Package ray implements the interface and methods to work with rays.
package ray

import "github.com/flynn-nrg/go-vfx/math32/vec3"

// Ray defines the methods used to work with rays.
type Ray interface {
	Origin() vec3.Vec3Impl
	Direction() vec3.Vec3Impl
	PointAtParameter(t float32) vec3.Vec3Impl
	Time() float32
	Lambda() float32
	SetLambda(float32)
}
