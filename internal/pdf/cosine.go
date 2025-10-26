package pdf

import (
	"github.com/flynn-nrg/go-vfx/math32"

	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/onb"
)

// Ensure interface compliance.
var _ PDF = (*Cosine)(nil)

type Cosine struct {
	uvw *onb.Onb
}

// NewCosine returns an instance of a cosine PDF.
func NewCosine(w vec3.Vec3Impl) *Cosine {
	o := onb.New()
	o.BuildFromW(w)
	return &Cosine{
		uvw: o,
	}
}

func (c *Cosine) Value(direction vec3.Vec3Impl) float32 {
	cosine := vec3.Dot(vec3.UnitVector(direction), c.uvw.W())
	if cosine > 0 {
		return cosine / math32.Pi
	}

	return 0
}

func (c *Cosine) Generate(random *fastrandom.XorShift) vec3.Vec3Impl {
	return c.uvw.Local(vec3.RandomCosineDirection(random))
}
