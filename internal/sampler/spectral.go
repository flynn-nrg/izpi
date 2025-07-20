package sampler

import (
	"math"
	"sync/atomic"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitable"
	"github.com/flynn-nrg/izpi/internal/pdf"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/spectral"
)

// SpectralSampler is a separate interface for spectral rendering
type SpectralSampler interface {
	Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) float64
}

var _ SpectralSampler = (*Spectral)(nil)

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

func (s *Spectral) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int, random *fastrandom.LCG) float64 {
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
				return srec.Attenuation() * s.Sample(srec.SpecularRay(), world, lightShape, depth+1, random)
			} else {
				pLight := pdf.NewHitable(lightShape, rec.P())
				p := pdf.NewMixture(pLight, srec.PDF())
				scattered := ray.New(rec.P(), p.Generate(random), r.Time())
				pdfVal := p.Value(scattered.Direction())
				// emitted + (albedo * scatteringPDF())*spectral() / pdf
				v1 := s.Sample(scattered, world, lightShape, depth+1, random) * mat.ScatteringPDF(r, rec, scattered)
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
