package material

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/vec3"
)

func refract(v *vec3.Vec3Impl, n *vec3.Vec3Impl, niOverNt float64) (*vec3.Vec3Impl, bool) {
	uv := vec3.UnitVector(v)

	dt := vec3.Dot(uv, n)
	discriminant := 1.0 - niOverNt*niOverNt*(1-dt*dt)
	if discriminant > 0 {
		// niOverNt * (uv - n*dt) - n*sqrt(discriminant)
		refracted := vec3.Sub(vec3.ScalarMul(vec3.Sub(uv, vec3.ScalarMul(n, dt)), niOverNt),
			vec3.ScalarMul(n, math.Sqrt(discriminant)))
		return refracted, true
	}
	return nil, false
}

func schlick(cosine float64, refIdx float64) float64 {
	r0 := (1.0 - refIdx) / (1.0 + refIdx)
	r0 = r0 * r0
	return r0 + (1.0-r0)*math.Pow((1.0-cosine), 5)
}
