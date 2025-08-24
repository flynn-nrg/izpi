package pdf

import (
	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ PDF = (*Mixture)(nil)

// Mixture represents a mixture of two PDFs.
type Mixture struct {
	p [2]PDF
}

// NewMixture returns an instance of the mixture PDF.
func NewMixture(p0 PDF, p1 PDF) *Mixture {
	return &Mixture{
		p: [2]PDF{p0, p1},
	}
}

func (m *Mixture) Value(direction *vec3.Vec3Impl) float32 {
	return 0.5*m.p[0].Value(direction) + 0.5*m.p[1].Value(direction)
}

func (m *Mixture) Generate(random *fastrandom.XorShift) *vec3.Vec3Impl {
	if random.Float32() < 0.5 {
		return m.p[0].Generate(random)
	}

	return m.p[1].Generate(random)
}
