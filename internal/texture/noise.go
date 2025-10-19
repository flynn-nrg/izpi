package texture

import (
	"github.com/flynn-nrg/go-vfx/math32"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/perlin"
)

// Ensure interface compliance.
var _ Texture = (*Noise)(nil)

// Noise represents a noise texture.
type Noise struct {
	perlin *perlin.Perlin
	scale  float32
}

// NewNoise returns an instance of the noise texture.
func NewNoise(scale float32) *Noise {
	return &Noise{
		perlin: perlin.New(),
		scale:  scale,
	}
}

func (n *Noise) Value(_ float32, _ float32, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return vec3.ScalarMul(&vec3.Vec3Impl{X: 1, Y: 1, Z: 1}, 0.5*(1+math32.Sin(n.scale*p.Z+10*n.perlin.Turb(p, 7))))
}
