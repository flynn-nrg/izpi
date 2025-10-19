package material

import (
	"math"

	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/onb"
	"github.com/flynn-nrg/izpi/internal/pdf"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/texture"
)

// Ensure interface compliance.
var _ Material = (*Lambertian)(nil)

// Lambertian represents a lambertian material.
type Lambertian struct {
	nonPBR
	nonEmitter
	nonPathLength
	nonWorldSetter
	albedo         texture.Texture
	spectralAlbedo texture.SpectralTexture
}

// NewLambertian returns an instance of the Lambert material.
func NewLambertian(albedo texture.Texture) *Lambertian {
	return &Lambertian{
		albedo: albedo,
	}
}

// NewSpectralLambertian returns an instance of the Lambert material with spectral support.
func NewSpectralLambertian(spectralAlbedo texture.SpectralTexture) *Lambertian {
	return &Lambertian{
		spectralAlbedo: spectralAlbedo,
	}
}

// scatterCommon contains the common scattering logic for both RGB and spectral rendering
func (l *Lambertian) scatterCommon(hr *hitrecord.HitRecord, random *fastrandom.XorShift, time float32) (*ray.RayImpl, *pdf.Cosine) {
	uvw := onb.New()
	uvw.BuildFromW(hr.Normal())
	direction := uvw.Local(vec3.RandomCosineDirection(random))
	scattered := ray.New(hr.P(), vec3.UnitVector(direction), time)
	pdf := pdf.NewCosine(hr.Normal())
	return scattered, pdf
}

// Scatter computes how the ray bounces off the surface of a diffuse material.
func (l *Lambertian) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.XorShift) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	scattered, pdf := l.scatterCommon(hr, random, r.Time())
	albedo := l.albedo.Value(hr.U(), hr.V(), hr.P())
	scatterRecord := scatterrecord.New(nil, false, albedo, nil, nil, nil, pdf)
	return scattered, scatterRecord, true
}

// SpectralScatter computes how the ray bounces off the surface of a diffuse material with spectral properties.
func (l *Lambertian) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.XorShift) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	scattered, pdf := l.scatterCommon(hr, random, r.Time())
	lambda := r.Lambda()
	albedo := l.spectralAlbedo.Value(hr.U(), hr.V(), lambda, hr.P())
	scatterRecord := scatterrecord.NewSpectralScatterRecord(nil, false, albedo, lambda, nil, 0.0, 0.0, pdf)

	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for diffuse materials.
func (l *Lambertian) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float32 {
	cosine := vec3.Dot(hr.Normal(), vec3.UnitVector(scattered.Direction()))
	if cosine < 0 {
		cosine = 0
	}

	return cosine / math.Pi
}

func (l *Lambertian) Albedo(u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return l.albedo.Value(u, v, p)
}

// SpectralAlbedo returns the spectral albedo at the given wavelength.
func (l *Lambertian) SpectralAlbedo(u float32, v float32, lambda float32, p *vec3.Vec3Impl) float32 {
	return l.spectralAlbedo.Value(u, v, lambda, p)
}
