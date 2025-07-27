package ray

import "github.com/flynn-nrg/izpi/internal/vec3"

// Ensure interface compliance.
var _ Ray = (*RayImpl)(nil)

// RayImpl implements the Ray interface.
type RayImpl struct {
	origin    *vec3.Vec3Impl
	direction *vec3.Vec3Impl
	lambda    float64
	time      float64
}

// New returns a new ray with the supplied origin and direction vectors and time.
func New(origin *vec3.Vec3Impl, direction *vec3.Vec3Impl, time float64) *RayImpl {
	return &RayImpl{
		origin:    origin,
		direction: direction,
		time:      time,
	}
}

// NewWithLambda returns a new ray with the supplied origin and direction vectors, time, and wavelength.
func NewWithLambda(origin *vec3.Vec3Impl, direction *vec3.Vec3Impl, time float64, lambda float64) *RayImpl {
	return &RayImpl{
		origin:    origin,
		direction: direction,
		lambda:    lambda,
		time:      time,
	}
}

// Origin returns the origin vector of this ray.
func (r *RayImpl) Origin() *vec3.Vec3Impl {
	return r.origin
}

// Direction returns the direction vector of this ray.
func (r *RayImpl) Direction() *vec3.Vec3Impl {
	return r.direction
}

// PointAtParameter is used to traverse the ray.
func (r *RayImpl) PointAtParameter(t float64) *vec3.Vec3Impl {
	return vec3.Add(r.origin, vec3.ScalarMul(r.direction, t))
}

// Time returns the time associated with this ray.
func (r *RayImpl) Time() float64 {
	return r.time
}

// Lambda returns the wavelength associated with this ray.
func (r *RayImpl) Lambda() float64 {
	return r.lambda
}

// SetLambda sets the wavelength associated with this ray.
func (r *RayImpl) SetLambda(lambda float64) {
	r.lambda = lambda
}
