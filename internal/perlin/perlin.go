// Package perlin implements functions to generate Perlin noise.
package perlin

import (
	"math"
	"math/rand"

	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Perlin represents an instance of a Perlin noise generator.
type Perlin struct {
	ranVec []vec3.Vec3Impl
	permX  []int
	permY  []int
	permZ  []int
}

func New() *Perlin {
	return &Perlin{
		ranVec: perlinGenerate(),
		permX:  perlinGeneratePerm(),
		permY:  perlinGeneratePerm(),
		permZ:  perlinGeneratePerm(),
	}
}

// Noise returns the noise value at a given position.
func (pl *Perlin) Noise(p vec3.Vec3Impl) float64 {
	var c [2][2][2]vec3.Vec3Impl

	u := p.X - math.Floor(p.X)
	v := p.Y - math.Floor(p.Y)
	w := p.Z - math.Floor(p.Z)
	i := int(math.Floor(p.X))
	j := int(math.Floor(p.Y))
	k := int(math.Floor(p.Z))

	for di := 0; di < 2; di++ {
		for dj := 0; dj < 2; dj++ {
			for dk := 0; dk < 2; dk++ {
				c[di][dj][dk] = pl.ranVec[pl.permX[(i+di)&255]^pl.permY[(j+dj)&255]^pl.permZ[(k+dk)&255]]
			}
		}
	}
	return trilinearInterp(c, u, v, w)
}

// Turb applies turbulence to this instance of Perlin noise.
func (pl *Perlin) Turb(p vec3.Vec3Impl, depth int) float64 {
	var accum float64

	tempP := p

	weight := float64(1.0)

	for i := 0; i < depth; i++ {
		accum += weight * pl.Noise(tempP)
		weight *= 0.5
		tempP = vec3.ScalarMul(tempP, 2.0)
	}

	return math.Abs(accum)
}

func perlinGenerate() []vec3.Vec3Impl {
	p := make([]vec3.Vec3Impl, 256)
	for i := range p {
		p[i] = vec3.UnitVector(vec3.Vec3Impl{X: -1 + 2*rand.Float64(), Y: -1 + 2*rand.Float64(), Z: -1 + 2*rand.Float64()})
	}

	return p
}

func permute(p []int) []int {
	for i := (len(p) - 1); i > 0; i-- {
		target := int(rand.Float64() * float64(i+1))
		tmp := p[i]
		p[i] = p[target]
		p[target] = tmp
	}

	return p
}

func perlinGeneratePerm() []int {
	p := make([]int, 256)
	for i := range p {
		p[i] = i
	}

	return permute(p)
}

func trilinearInterp(c [2][2][2]vec3.Vec3Impl, u float64, v float64, w float64) float64 {
	var accum float64

	uu := u * u * (3 - 2*u)
	vv := v * v * (3 - 2*v)
	ww := w * w * (3 - 2*w)

	for i := 0; i < 2; i++ {
		for j := 0; j < 2; j++ {
			for k := 0; k < 2; k++ {
				weightV := vec3.Vec3Impl{X: u - float64(i), Y: v - float64(j), Z: w - float64(k)}
				accum += (float64(i)*uu + (1.0-float64(i))*(1.0-uu)) *
					(float64(j)*vv + (1.0-float64(j))*(1.0-vv)) *
					(float64(k)*ww + (1.0-float64(k))*(1.0-ww)) * vec3.Dot(c[i][j][k], weightV)
			}
		}
	}

	return accum
}
