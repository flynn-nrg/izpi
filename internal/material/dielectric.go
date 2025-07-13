package material

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Material = (*Dielectric)(nil)

// Dielectric represents a dielectric material.
type Dielectric struct {
	nonPBR
	nonEmitter
	refIdx         float64
	spectralRefIdx texture.SpectralTexture
}

// NewDielectric returns an instance of a dielectric material.
func NewDielectric(reIdx float64) *Dielectric {
	return &Dielectric{
		refIdx: reIdx,
	}
}

// NewSpectralDielectric returns an instance of a dielectric material with spectral refractive index.
func NewSpectralDielectric(spectralRefIdx texture.SpectralTexture) *Dielectric {
	return &Dielectric{
		spectralRefIdx: spectralRefIdx,
	}
}

// scatterCommon contains the common scattering logic for both RGB and spectral rendering
func (d *Dielectric) scatterCommon(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG, refIdx float64) (*ray.RayImpl, bool) {
	var niOverNt float64
	var cosine float64
	var reflectProb float64
	var scattered *ray.RayImpl
	var refracted *vec3.Vec3Impl
	var ok bool
	var outwardNormal *vec3.Vec3Impl

	reflected := reflect(r.Direction(), hr.Normal())

	if vec3.Dot(r.Direction(), hr.Normal()) > 0 {
		outwardNormal = vec3.ScalarMul(hr.Normal(), -1.0)
		niOverNt = refIdx
		cosine = refIdx * vec3.Dot(r.Direction(), hr.Normal()) / r.Direction().Length()
	} else {
		outwardNormal = hr.Normal()
		niOverNt = 1.0 / refIdx
		cosine = -vec3.Dot(r.Direction(), hr.Normal()) / r.Direction().Length()
	}

	if refracted, ok = refract(r.Direction(), outwardNormal, niOverNt); ok {
		reflectProb = schlick(cosine, refIdx)
	} else {
		reflectProb = 1.0
	}

	if random.Float64() < reflectProb {
		scattered = ray.New(hr.P(), reflected, r.Time())
	} else {
		scattered = ray.New(hr.P(), refracted, r.Time())
	}
	return scattered, true
}

// Scatter computes how the ray bounces off the surface of a dielectric material.
func (d *Dielectric) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	scattered, ok := d.scatterCommon(r, hr, random, d.refIdx)
	if !ok {
		return nil, nil, false
	}
	attenuation := &vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0}
	scatterRecord := scatterrecord.New(scattered, true, attenuation, nil, nil, nil, nil)
	return scattered, scatterRecord, true
}

// SpectralScatter computes how the ray bounces off the surface of a dielectric material with spectral properties.
func (d *Dielectric) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	// For spectral rendering, we need to get the wavelength from the ray
	// Assuming the ray has a wavelength field or we need to sample one
	lambda := 550.0 // Default wavelength, should be extracted from ray or sampled

	// Get wavelength-dependent refractive index
	refIdx := d.spectralRefIdx.Value(hr.U(), hr.V(), lambda, hr.P())

	// Use the common scattering logic with spectral refractive index
	scattered, ok := d.scatterCommon(r, hr, random, refIdx)
	if !ok {
		return nil, nil, false
	}

	albedo := 1.0 // Dielectrics have no absorption in the visible spectrum
	scatterRecord := scatterrecord.NewSpectralScatterRecord(scattered, true, albedo, lambda, nil, 0.0, 0.0, nil)
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

// SpectralAlbedo returns the spectral albedo at the given wavelength.
// Dielectrics have no absorption in the visible spectrum, so this returns 1.0.
func (d *Dielectric) SpectralAlbedo(u float64, v float64, lambda float64, p *vec3.Vec3Impl) float64 {
	return 1.0
}
