package material

import (
	"math/rand"

	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*Dielectric)(nil)

// Dielectric represents a dielectric material.
type Dielectric struct {
	refIdx float64
}

// NewDielectric returns an instance of a dielectric material.
func NewDielectric(reIdx float64) *Dielectric {
	return &Dielectric{
		refIdx: reIdx,
	}
}

// Scatter computes how the ray bounces off the surface of a dielectric material.
func (d *Dielectric) Scatter(r ray.Ray, hr *hitrecord.HitRecord) (*ray.RayImpl, *vec3.Vec3Impl, bool) {
	var niOverNt float64
	var cosine float64
	var reflectProb float64
	var scattered *ray.RayImpl
	var refracted *vec3.Vec3Impl
	var ok bool

	outwardNormal := &vec3.Vec3Impl{}
	reflected := reflect(r.Direction(), hr.Normal())
	attenuation := &vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0}

	if vec3.Dot(r.Direction(), hr.Normal()) > 0 {
		outwardNormal = vec3.ScalarMul(hr.Normal(), -1.0)
		niOverNt = d.refIdx
		cosine = d.refIdx * vec3.Dot(r.Direction(), hr.Normal()) / r.Direction().Length()
	} else {
		outwardNormal = hr.Normal()
		niOverNt = 1.0 / d.refIdx
		cosine = -vec3.Dot(r.Direction(), hr.Normal()) / r.Direction().Length()
	}

	if refracted, ok = refract(r.Direction(), outwardNormal, niOverNt); ok {
		reflectProb = schlick(cosine, d.refIdx)
	} else {
		scattered = ray.New(hr.P(), reflected, r.Time())
		reflectProb = 1.0
	}

	if rand.Float64() < reflectProb {
		scattered = ray.New(hr.P(), reflected, r.Time())
	} else {
		scattered = ray.New(hr.P(), refracted, r.Time())
	}

	return scattered, attenuation, true
}

// Emitted returns black for dielectrics materials.
func (d *Dielectric) Emitted(_ float64, _ float64, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{}
}
