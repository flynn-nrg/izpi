package texture

import "github.com/flynn-nrg/go-vfx/math32/vec3"

// Ensure interface compliance.
var _ Texture = (*Constant)(nil)

// Constant represents a constant texture.
type Constant struct {
	color *vec3.Vec3Impl
}

// NewConstant returns an instance of the constant texture.
func NewConstant(color *vec3.Vec3Impl) *Constant {
	return &Constant{
		color: color,
	}
}

func (c *Constant) Value(_ float32, _ float32, _ *vec3.Vec3Impl) *vec3.Vec3Impl {
	col := *c.color
	return &col
}
