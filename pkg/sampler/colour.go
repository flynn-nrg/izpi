package sampler

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/hitable"
	"github.com/flynn-nrg/izpi/pkg/pdf"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Ensure interface compliance.
var _ Sampler = (*Colour)(nil)

type Colour struct {
	maxDepth   int
	background *vec3.Vec3Impl
}

func NewColour(maxDepth int, background *vec3.Vec3Impl) *Colour {
	return &Colour{
		maxDepth:   maxDepth,
		background: background,
	}
}

func (cs *Colour) Sample(r ray.Ray, world *hitable.HitableSlice, lightShape hitable.Hitable, depth int) *vec3.Vec3Impl {
	if depth >= cs.maxDepth {
		return &vec3.Vec3Impl{Z: 1.0}
	}

	if rec, mat, ok := world.Hit(r, 0.001, math.MaxFloat64); ok {
		_, srec, ok := mat.Scatter(r, rec)
		emitted := mat.Emitted(r, rec, rec.U(), rec.V(), rec.P())
		if depth < cs.maxDepth && ok {
			if srec.IsSpecular() {
				// srec.Attenuation() * colour(...)
				return vec3.Mul(srec.Attenuation(), cs.Sample(srec.SpecularRay(), world, lightShape, depth+1))
			} else {
				pLight := pdf.NewHitable(lightShape, rec.P())
				p := pdf.NewMixture(pLight, srec.PDF())
				scattered := ray.New(rec.P(), p.Generate(), r.Time())
				pdfVal := p.Value(scattered.Direction())
				// emitted + (albedo * scatteringPDF())*colour() / pdf
				v1 := vec3.ScalarMul(cs.Sample(scattered, world, lightShape, depth+1), mat.ScatteringPDF(r, rec, scattered))
				v2 := vec3.Mul(srec.Attenuation(), v1)
				v3 := vec3.ScalarDiv(v2, pdfVal)
				res := vec3.Add(emitted, v3)
				return res
			}
		} else {
			return emitted
		}
	}

	b := *cs.background
	return &b
}
