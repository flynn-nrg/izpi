package material

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/pdf"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Material = (*Isotropic)(nil)

// Isotropic represents an isotropic material.
type Isotropic struct {
	nonPBR
	nonEmitter
	nonPathLength
	nonWorldSetter
	albedo texture.Texture
}

// NewIsotropic returns a new instances of the isotropic material.
func NewIsotropic(albedo texture.Texture) *Isotropic {
	return &Isotropic{
		albedo: albedo,
	}
}

// Scatter computes how the ray bounces off the surface of a diffuse material.
func (i *Isotropic) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	scattered := ray.New(hr.P(), randomInUnitSphere(random), r.Time())
	attenuation := i.albedo.Value(hr.U(), hr.V(), hr.P())
	pdf := pdf.NewCosine(hr.Normal())
	scatterRecord := scatterrecord.New(nil, false, attenuation, nil, nil, nil, pdf)
	return scattered, scatterRecord, true
}

// SpectralScatter computes how the ray bounces off the surface of an isotropic material with spectral properties.
func (i *Isotropic) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	scattered := ray.NewWithLambda(hr.P(), randomInUnitSphere(random), r.Time(), r.Lambda())
	lambda := r.Lambda()
	albedo := i.albedo.Value(hr.U(), hr.V(), hr.P()).X // Use red component as approximation
	pdf := pdf.NewCosine(hr.Normal())
	scatterRecord := scatterrecord.NewSpectralScatterRecord(nil, false, albedo, lambda, nil, 0.0, 0.0, pdf)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for isotropic materials.
func (i *Isotropic) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	return 0
}

func (i *Isotropic) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return i.albedo.Value(u, v, p)
}

// SpectralAlbedo returns the spectral albedo at the given wavelength.
func (i *Isotropic) SpectralAlbedo(u float64, v float64, lambda float64, p *vec3.Vec3Impl) float64 {
	return i.albedo.Value(u, v, p).X // Use red component as approximation
}
