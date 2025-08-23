package texture

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Texture = (*Checker)(nil)

// Checker represents a checker board pattern texture.
type Checker struct {
	odd  Texture
	even Texture
}

// NewChecker returns a new instance of the Checker texture.
func NewChecker(odd Texture, even Texture) *Checker {
	return &Checker{
		odd:  odd,
		even: even,
	}
}

func (c *Checker) Value(u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	sines := float32(math.Sin(float64(10.0*p.X)) * math.Sin(float64(10.0*p.Y)) * math.Sin(float64(10.0*p.Z)))
	if sines < 0 {
		return c.odd.Value(u, v, p)
	}

	return c.even.Value(u, v, p)
}
