package material

import (
	"github.com/flynn-nrg/go-vfx/math32"

	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
)

func randomInUnitSphere(random *fastrandom.XorShift) vec3.Vec3Impl {
	for {
		p := vec3.Sub(vec3.ScalarMul(vec3.Vec3Impl{X: random.Float32(), Y: random.Float32(), Z: random.Float32()}, 2.0),
			vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0})
		if p.SquaredLength() < 1.0 {
			return p
		}
	}
}

func reflect(v vec3.Vec3Impl, n vec3.Vec3Impl) vec3.Vec3Impl {
	// v - 2*dot(v,n)*n
	return vec3.Sub(v, vec3.ScalarMul(n, 2*vec3.Dot(v, n)))
}

func refract(v vec3.Vec3Impl, n vec3.Vec3Impl, niOverNt float32) (vec3.Vec3Impl, bool) {
	uv := vec3.UnitVector(v)

	dt := vec3.Dot(uv, n)
	discriminant := 1.0 - niOverNt*niOverNt*(1-dt*dt)
	if discriminant > 0 {
		// niOverNt * (uv - n*dt) - n*sqrt(discriminant)
		refracted := vec3.Sub(vec3.ScalarMul(vec3.Sub(uv, vec3.ScalarMul(n, dt)), niOverNt),
			vec3.ScalarMul(n, math32.Sqrt(discriminant)))
		return refracted, true
	}
	return vec3.Vec3Impl{}, false
}

func schlick(cosine float32, refIdx float32) float32 {
	r0 := (1.0 - refIdx) / (1.0 + refIdx)
	r0 = r0 * r0
	return r0 + (1.0-r0)*math32.Pow((1.0-cosine), 5)
}
