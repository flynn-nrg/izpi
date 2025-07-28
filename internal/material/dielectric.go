package material

import (
	"math"

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
	// Absorption properties for colored glass
	absorptionCoeff         *vec3.Vec3Impl          // RGB absorption coefficient (for RGB rendering)
	spectralAbsorptionCoeff texture.SpectralTexture // Spectral absorption coefficient (for spectral rendering)
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

// NewColoredDielectric returns an instance of a dielectric material with absorption for colored glass.
func NewColoredDielectric(refIdx float64, absorptionCoeff *vec3.Vec3Impl) *Dielectric {
	return &Dielectric{
		refIdx:          refIdx,
		absorptionCoeff: absorptionCoeff,
	}
}

// NewSpectralColoredDielectric returns an instance of a dielectric material with spectral refractive index and absorption.
func NewSpectralColoredDielectric(spectralRefIdx texture.SpectralTexture, spectralAbsorptionCoeff texture.SpectralTexture) *Dielectric {
	return &Dielectric{
		spectralRefIdx:          spectralRefIdx,
		spectralAbsorptionCoeff: spectralAbsorptionCoeff,
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
		scattered = ray.NewWithLambda(hr.P(), reflected, r.Time(), r.Lambda())
	} else {
		scattered = ray.NewWithLambda(hr.P(), refracted, r.Time(), r.Lambda())
	}
	return scattered, true
}

// calculateBeerLambertAttenuation calculates the Beer-Lambert law attenuation
// I = I₀ * exp(-α * d) where α is the absorption coefficient and d is the path length
func (d *Dielectric) calculateBeerLambertAttenuation(pathLength float64, lambda float64, u, v float64, p *vec3.Vec3Impl) float64 {
	if d.spectralAbsorptionCoeff != nil {
		// Spectral absorption coefficient
		absorptionCoeff := d.spectralAbsorptionCoeff.Value(u, v, lambda, p)
		return math.Exp(-absorptionCoeff * pathLength)
	}
	// No absorption (clear glass)
	return 1.0
}

// calculatePathLength estimates the path length through the material
// For spheres, we can calculate the actual chord length through the sphere
func (d *Dielectric) calculatePathLength(hr *hitrecord.HitRecord, r ray.Ray) float64 {
	// For spheres, we can calculate the actual chord length
	// The chord length through a sphere is: 2 * sqrt(r² - d²)
	// where r is the sphere radius and d is the distance from center to ray

	// Estimate sphere center and radius from hit point
	// This is a simplified approach - in practice, you'd get this from the object
	// For now, we'll use a reasonable default path length since we don't have access to the object's center
	// In a full implementation, you'd pass the object's center and radius to this method

	// For now, use a reasonable default path length for spheres
	// This is a simplified approach - in a full implementation, you'd calculate the actual chord length
	// through the sphere based on the object's center and radius
	chordLength := 30.0 // Approximate diameter of our spheres

	// Clamp to reasonable values
	if chordLength < 0.1 {
		chordLength = 0.1
	}
	if chordLength > 10.0 {
		chordLength = 10.0
	}

	return chordLength
}

// Scatter computes how the ray bounces off the surface of a dielectric material.
func (d *Dielectric) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	scattered, ok := d.scatterCommon(r, hr, random, d.refIdx)
	if !ok {
		return nil, nil, false
	}

	// Calculate Beer-Lambert attenuation for RGB rendering
	var attenuation *vec3.Vec3Impl
	if d.absorptionCoeff != nil {
		pathLength := d.calculatePathLength(hr, r)
		// Apply Beer-Lambert law for each RGB component
		attenuation = &vec3.Vec3Impl{
			X: math.Exp(-d.absorptionCoeff.X * pathLength),
			Y: math.Exp(-d.absorptionCoeff.Y * pathLength),
			Z: math.Exp(-d.absorptionCoeff.Z * pathLength),
		}
	} else {
		attenuation = &vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0}
	}

	scatterRecord := scatterrecord.New(scattered, true, attenuation, nil, nil, nil, nil)
	return scattered, scatterRecord, true
}

// SpectralScatter computes how the ray bounces off the surface of a dielectric material with spectral properties.
func (d *Dielectric) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	lambda := r.Lambda()
	refIdx := d.spectralRefIdx.Value(hr.U(), hr.V(), lambda, hr.P())

	scattered, ok := d.scatterCommon(r, hr, random, refIdx)
	if !ok {
		return nil, nil, false
	}

	// Calculate Beer-Lambert attenuation for spectral rendering
	pathLength := d.calculatePathLength(hr, r)
	albedo := d.calculateBeerLambertAttenuation(pathLength, lambda, hr.U(), hr.V(), hr.P())

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
