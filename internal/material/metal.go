package material

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Material = (*Metal)(nil)

// Metal represents metallic materials.
type Metal struct {
	nonEmitter
	nonPBR
	nonSpectral
	albedo *vec3.Vec3Impl
	fuzz   float64
}

// NewMetal returns an instance of the metal material.
func NewMetal(albedo *vec3.Vec3Impl, fuzz float64) *Metal {
	return &Metal{
		albedo: albedo,
		fuzz:   fuzz,
	}
}

// Scatter computes how the ray bounces off the surface of a metallic object.
func (m *Metal) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	reflected := reflect(vec3.UnitVector(r.Direction()), hr.Normal())
	specular := ray.New(hr.P(), vec3.Add(reflected, vec3.ScalarMul(randomInUnitSphere(random), m.fuzz)), r.Time())
	attenuation := m.albedo
	scatterRecord := scatterrecord.New(specular, true, attenuation, nil, nil, nil, nil)
	return nil, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for metals.
func (m *Metal) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	return 0
}

func (m *Metal) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	a := *m.albedo
	return &a
}
