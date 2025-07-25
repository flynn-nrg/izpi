package material

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Material = (*Dielectric)(nil)

// Dielectric represents a dielectric material.
type Dielectric struct {
	nonPBR
	refIdx float64
}

// NewDielectric returns an instance of a dielectric material.
func NewDielectric(reIdx float64) *Dielectric {
	return &Dielectric{
		refIdx: reIdx,
	}
}

// Scatter computes how the ray bounces off the surface of a dielectric material.
func (d *Dielectric) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	var niOverNt float64
	var cosine float64
	var reflectProb float64
	var scattered *ray.RayImpl
	var refracted *vec3.Vec3Impl
	var ok bool
	var outwardNormal *vec3.Vec3Impl

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
		reflectProb = 1.0
	}

	if random.Float64() < reflectProb {
		scattered = ray.New(hr.P(), reflected, r.Time())
	} else {
		scattered = ray.New(hr.P(), refracted, r.Time())
	}
	scatterRecord := scatterrecord.New(scattered, true, attenuation, nil, nil, nil, nil)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for dielectric materials.
func (d *Dielectric) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	return 0
}

// IsEmitter() is true for dielectric materials as they reflect light and can be considered emitters.
func (d *Dielectric) IsEmitter() bool {
	return true
}

func (d *Dielectric) Emitted(_ ray.Ray, _ *hitrecord.HitRecord, _ float64, _ float64, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{}
}

func (d *Dielectric) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0}
}
