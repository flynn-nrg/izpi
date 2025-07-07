package texture

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/perlin"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Texture = (*Noise)(nil)

// Noise represents a noise texture.
type Noise struct {
	perlin *perlin.Perlin
	scale  float64
}

// NewNoise returns an instance of the noise texture.
func NewNoise(scale float64) *Noise {
	return &Noise{
		perlin: perlin.New(),
		scale:  scale,
	}
}

func (n *Noise) Value(_ float64, _ float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return vec3.ScalarMul(&vec3.Vec3Impl{X: 1, Y: 1, Z: 1}, 0.5*(1+math.Sin(n.scale*p.Z+10*n.perlin.Turb(p, 7))))
}
