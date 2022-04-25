// Package onb implements methods to work with ortho-normal bases
package onb

import (
	"math"

	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Onb represents an ortho-normal base.
type Onb struct {
	axis []*vec3.Vec3Impl
}

// New returns an instance of an ortho-normal base.
func New() *Onb {
	return &Onb{
		axis: make([]*vec3.Vec3Impl, 3),
	}
}

// U returns the first axis of the ortho-normal base.
func (o *Onb) U() *vec3.Vec3Impl {
	return o.axis[0]
}

// U returns the second axis of the ortho-normal base.
func (o *Onb) V() *vec3.Vec3Impl {
	return o.axis[1]
}

// U returns the third axis of the ortho-normal base.
func (o *Onb) W() *vec3.Vec3Impl {
	return o.axis[2]
}

// BuildFromW constructs the ortho-normal base from the provided vector.
func (o *Onb) BuildFromW(n *vec3.Vec3Impl) {
	// W
	o.axis[2] = vec3.UnitVector(n)
	var a *vec3.Vec3Impl
	if math.Abs(o.W().X) > 0.9 {
		a = &vec3.Vec3Impl{Y: 1}
	} else {
		a = &vec3.Vec3Impl{X: 1}
	}
	// V
	o.axis[1] = vec3.UnitVector(vec3.Cross(o.W(), a))
	// U
	o.axis[0] = vec3.Cross(o.W(), o.V())
}

// ScalarLocal returns the ortho-normal base local to the supplied position.
func (o *Onb) ScalarLocal(a, b, c float64) *vec3.Vec3Impl {
	// a*u + b*v + c*w
	return vec3.Add(vec3.ScalarMul(o.U(), a),
		vec3.ScalarMul(o.V(), b),
		vec3.ScalarMul(o.W(), c))

}

// Local returns the ortho-normal base local to the supplied position.
func (o *Onb) Local(a *vec3.Vec3Impl) *vec3.Vec3Impl {
	// a.x*u + a.y*v + a.z*w
	return vec3.Add(vec3.ScalarMul(o.U(), a.X),
		vec3.ScalarMul(o.V(), a.Y),
		vec3.ScalarMul(o.W(), a.Z))
}
