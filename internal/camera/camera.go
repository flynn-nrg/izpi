// Package camera implements a set of functions to work with cameras.
package camera

import (
	"github.com/flynn-nrg/go-vfx/math32"
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
	"github.com/flynn-nrg/izpi/internal/ray"
)

// Camera represents a camera in the world.
type Camera struct {
	random          *fastrandom.XorShift
	lensRadius      float32
	time0           float32
	time1           float32
	u               *vec3.Vec3Impl
	v               *vec3.Vec3Impl
	origin          *vec3.Vec3Impl
	lowerLeftCorner *vec3.Vec3Impl
	horizontal      *vec3.Vec3Impl
	vertical        *vec3.Vec3Impl
}

// New returns an instance of a camera.
func New(lookFrom *vec3.Vec3Impl, lookAt *vec3.Vec3Impl, vup *vec3.Vec3Impl,
	vfov float32, aspect float32, aperture float32, focusDist float32, time0 float32, time1 float32) *Camera {

	lensRadius := aperture / 2.0
	theta := vfov * math32.Pi / 180
	halfHeight := math32.Tan(theta / 2.0)
	halfWidth := aspect * halfHeight
	w := vec3.UnitVector(vec3.Sub(lookFrom, lookAt))
	u := vec3.UnitVector(vec3.Cross(vup, w))
	v := vec3.Cross(w, u)

	// origin - halfWidth*focusDist*u - halfHeight*focusDist*v - focusDist*w
	lowerLeftCorner := vec3.Sub(lookFrom, vec3.ScalarMul(u, halfWidth*focusDist), vec3.ScalarMul(v, halfHeight*focusDist), vec3.ScalarMul(w, focusDist))
	horizontal := vec3.ScalarMul(u, 2.0*halfWidth*focusDist)
	vertical := vec3.ScalarMul(v, 2.0*halfHeight*focusDist)
	origin := lookFrom

	return &Camera{
		random:          fastrandom.NewWithDefaults(),
		lensRadius:      lensRadius,
		time0:           time0,
		time1:           time1,
		u:               u,
		v:               v,
		lowerLeftCorner: lowerLeftCorner,
		horizontal:      horizontal,
		vertical:        vertical,
		origin:          origin,
	}
}

// GetRay returns the ray associated for the supplied u and v.
func (c *Camera) GetRay(s float32, t float32) *ray.RayImpl {
	rd := vec3.ScalarMul(c.randomInUnitDisc(), c.lensRadius)
	offset := vec3.Add(vec3.ScalarMul(c.u, rd.X), vec3.ScalarMul(c.v, rd.Y))
	time := c.time0 + c.random.Float32()*(c.time1-c.time0)
	return ray.New(vec3.Add(c.origin, offset),
		// lowerLeftCorner + s*horizontal + t*vertical - origin - offset
		vec3.Sub(vec3.Add(c.lowerLeftCorner, vec3.ScalarMul(c.horizontal, s),
			vec3.ScalarMul(c.vertical, t)), c.origin, offset), time)
}

// GetRayWithLambda returns the ray associated for the supplied u and v with a specific wavelength.
func (c *Camera) GetRayWithLambda(s float32, t float32, lambda float32) *ray.RayImpl {
	rd := vec3.ScalarMul(c.randomInUnitDisc(), c.lensRadius)
	offset := vec3.Add(vec3.ScalarMul(c.u, rd.X), vec3.ScalarMul(c.v, rd.Y))
	time := c.time0 + c.random.Float32()*(c.time1-c.time0)
	return ray.NewWithLambda(vec3.Add(c.origin, offset),
		// lowerLeftCorner + s*horizontal + t*vertical - origin - offset
		vec3.Sub(vec3.Add(c.lowerLeftCorner, vec3.ScalarMul(c.horizontal, s),
			vec3.ScalarMul(c.vertical, t)), c.origin, offset), time, lambda)
}

func (c *Camera) randomInUnitDisc() *vec3.Vec3Impl {
	for {
		p := vec3.Sub(vec3.ScalarMul(&vec3.Vec3Impl{X: c.random.Float32(), Y: c.random.Float32()}, 2.0), &vec3.Vec3Impl{X: 1.0, Y: 1.0})
		if vec3.Dot(p, p) < 1.0 {
			return p
		}
	}
}
