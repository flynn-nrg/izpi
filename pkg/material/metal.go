package material

import (
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*Metal)(nil)

// Metal represents metallic materials.
type Metal struct {
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
func (m *Metal) Scatter(r ray.Ray, hr *hitrecord.HitRecord) (*ray.RayImpl, *vec3.Vec3Impl, bool) {
	reflected := reflect(vec3.UnitVector(r.Direction()), hr.Normal())
	scattered := ray.New(hr.P(), vec3.Add(reflected, vec3.ScalarMul(randomInUnitSphere(), m.fuzz)), r.Time())
	attenuation := m.albedo
	return scattered, attenuation, (vec3.Dot(scattered.Direction(), hr.Normal()) > 0)
}

// Emitted returns black for metallic materials.
func (m *Metal) Emitted(_ float64, _ float64, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{}
}
