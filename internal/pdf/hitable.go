package pdf

import (
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/hitabletarget"
)

// Ensure interface compliance.
var _ PDF = (*Hitable)(nil)

// Hitable represents a hitable PDF.
type Hitable struct {
	o       vec3.Vec3Impl
	hitable hitabletarget.HitableTarget
}

// NewHitable returns an instance of a hitable PDF.
func NewHitable(p hitabletarget.HitableTarget, origin vec3.Vec3Impl) *Hitable {
	return &Hitable{
		o:       origin,
		hitable: p,
	}
}

func (h *Hitable) Value(direction vec3.Vec3Impl) float32 {
	return h.hitable.PDFValue(h.o, direction)
}

func (h *Hitable) Generate(random *fastrandom.XorShift) vec3.Vec3Impl {
	return h.hitable.Random(h.o, random)
}
