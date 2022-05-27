package pdf

import (
	"math/rand"

	"github.com/flynn-nrg/izpi/pkg/vec3"
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

func (m *Mixture) Value(direction *vec3.Vec3Impl) float64 {
	return 0.5*m.p[0].Value(direction) + 0.5*m.p[1].Value(direction)
}

func (m *Mixture) Generate() *vec3.Vec3Impl {
	if rand.Float64() < 0.5 {
		return m.p[0].Generate()
	}

	return m.p[1].Generate()
}
