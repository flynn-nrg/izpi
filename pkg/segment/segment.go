package segment

import (
	"github.com/flynn-nrg/izpi/pkg/vec3"
)

// Segment represents a segment in the 3d space.
type Segment struct {
	A *vec3.Vec3Impl
	B *vec3.Vec3Impl
}

func Belongs(s *Segment, c *vec3.Vec3Impl) bool {
	ab := vec3.Sub(s.B, s.A)
	ac := vec3.Sub(c, s.A)

	// Check if they're colinear.
	if vec3.Cross(ab, ac).Length() != 0 {
		return false
	}

	kac := vec3.Dot(ab, ac)
	kab := vec3.Dot(ab, ab)

	if kac < 0 || kac > kab {
		return false
	}

	return true
}
