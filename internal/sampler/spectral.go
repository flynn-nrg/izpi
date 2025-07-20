package sampler

import (
	"math"
	"sync/atomic"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/pdf"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/spectral"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

var _ Sampler = (*Spectral)(nil)

type Spectral struct {
	maxDepth   int
	numRays    *uint64
	background *spectral.SpectralPowerDistribution
}

func NewSpectral(maxDepth int, background *spectral.SpectralPowerDistribution, numRays *uint64) *Spectral {
	return &Spectral{
		maxDepth:   maxDepth,
		numRays:    numRays,
		background: background,
	}
}

// NonSpectral is a stub implementation for RGB samplers
// that don't support spectral rendering. It provides the SampleSpectral method
// that RGB samplers can embed to satisfy the Sampler interface.
type NonSpectral struct{}

func NewNonSpectral() *NonSpectral {
	return &NonSpectral{}
}

// SampleSpectral implements the Sampler interface for RGB samplers
// Returns a neutral value (0.5) since RGB samplers don't support spectral rendering
func (ns *NonSpectral) SampleSpectral(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) float64 {
	// Return neutral value for RGB samplers that don't support spectral rendering
	return 0.5
}

func (s *Spectral) SampleSpectral(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) float64 {
	if depth >= s.maxDepth {
		// Return the background spectral power distribution at the wavelength of the ray
		return s.background.Value(r.Lambda())
	}

	atomic.AddUint64(s.numRays, 1)

	// L(λ) = Le(λ) + ∫ f(λ) * L(λ) * cos(θ) / p(ω) dω
	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		_, srec, ok := mat.SpectralScatter(r, rec, random)
		emitted := mat.EmittedSpectral(r, rec, rec.U(), rec.V(), r.Lambda(), rec.P())
		if depth < s.maxDepth && ok {
			if srec.IsSpecular() {
				return srec.Attenuation() * s.SampleSpectral(srec.SpecularRay(), world, lightShape, depth+1, random)
			} else {
				pLight := pdf.NewHitable(lightShape, rec.P())
				p := pdf.NewMixture(pLight, srec.PDF())
				scattered := ray.New(rec.P(), p.Generate(random), r.Time())
				pdfVal := p.Value(scattered.Direction())
				// emitted + (albedo * scatteringPDF())*spectral() / pdf
				v1 := s.SampleSpectral(scattered, world, lightShape, depth+1, random) * mat.ScatteringPDF(r, rec, scattered)
				v2 := srec.Attenuation() * v1
				v3 := v2 / pdfVal
				res := emitted + v3
				return res
			}
		} else {
			return emitted
		}
	}

	return s.background.Value(r.Lambda())
}

// Sample implements the Sampler interface for RGB rendering
// Converts spectral rendering to RGB by sampling at the ray's wavelength
// and converting to RGB using the spectral conversion utilities
func (s *Spectral) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) *vec3.Vec3Impl {
	// For RGB rendering, we need to convert spectral values to RGB
	// This is a simplified approach - in practice, you'd want to sample multiple wavelengths
	spectralValue := s.SampleSpectral(r, world, lightShape, depth, random)

	// Convert single wavelength to RGB using the spectral conversion
	red, green, blue := spectral.WavelengthToRGB(r.Lambda())

	// Scale by the spectral value
	return &vec3.Vec3Impl{
		X: red * spectralValue,
		Y: green * spectralValue,
		Z: blue * spectralValue,
	}
}
