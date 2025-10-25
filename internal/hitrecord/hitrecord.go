package hitrecord

import "github.com/flynn-nrg/izpi/internal/vec3"

// HitRecord contains data related to an intersection between a ray and an object.
type HitRecord struct {
	u      float64
	v      float64
	t      float64
	p      vec3.Vec3Impl
	normal vec3.Vec3Impl
}

func New(t float64, u float64, v float64, p vec3.Vec3Impl, normal vec3.Vec3Impl) *HitRecord {
	return &HitRecord{
		u:      u,
		v:      v,
		t:      t,
		p:      p,
		normal: normal,
	}
}

// Normal returns the normal vector at the intersection point.
func (hr *HitRecord) Normal() vec3.Vec3Impl {
	return hr.normal
}

// P returns the intersection point.
func (hr *HitRecord) P() vec3.Vec3Impl {
	return hr.p
}

// T returns the t value.
func (hr *HitRecord) T() float64 {
	return hr.t
}

// U returns the u value.
func (hr *HitRecord) U() float64 {
	return hr.u
}

// V returns the V value.
func (hr *HitRecord) V() float64 {
	return hr.v
}
