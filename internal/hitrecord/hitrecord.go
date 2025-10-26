package hitrecord

import "github.com/flynn-nrg/go-vfx/math32/vec3"

// HitRecord contains data related to an intersection between a ray and an object.
type HitRecord struct {
	u      float32
	v      float32
	t      float32
	p      vec3.Vec3Impl
	normal vec3.Vec3Impl
}

func New(t float32, u float32, v float32, p vec3.Vec3Impl, normal vec3.Vec3Impl) *HitRecord {
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
func (hr *HitRecord) T() float32 {
	return hr.t
}

// U returns the u value.
func (hr *HitRecord) U() float32 {
	return hr.u
}

// V returns the V value.
func (hr *HitRecord) V() float32 {
	return hr.v
}
