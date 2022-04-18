package material

import (
	"gitlab.com/flynn-nrg/izpi/pkg/hitrecord"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/texture"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Material = (*Lambertian)(nil)

// Lambertian represents a diffuse material.
type Lambertian struct {
	albedo texture.Texture
}

// NewLambertian returns an instance of the Lambert material.
func NewLambertian(albedo texture.Texture) *Lambertian {
	return &Lambertian{
		albedo: albedo,
	}
}

// Scatter computes how the ray bounces off the surface of a diffuse material.
func (l *Lambertian) Scatter(r ray.Ray, hr *hitrecord.HitRecord) (*ray.RayImpl, *vec3.Vec3Impl, bool) {
	target := vec3.Add(hr.P(), hr.Normal(), randomInUnitSphere())
	return ray.New(hr.P(), vec3.Sub(target, hr.P()), r.Time()), l.albedo.Value(hr.U(), hr.V(), hr.P()), true
}

// Emitted returns black for Lambertian materials.
func (l *Lambertian) Emitted(_ float64, _ float64, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	return &vec3.Vec3Impl{}
}
