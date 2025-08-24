// Package vec3 provides utility functions to work with vectors.
package vec3

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
)

// Vec3Impl defines a vector with its position and colour.
type Vec3Impl struct {
	X float32
	Y float32
	Z float32
	R float32
	G float32
	B float32
}

// Length returns the length of this vector.
func (v *Vec3Impl) Length() float32 {
	return float32(math.Sqrt(float64((v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z))))
}

// SquaredLength returns the squared length of this vector.
func (v *Vec3Impl) SquaredLength() float32 {
	return (v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z)
}

// MakeUnitVector transform the vector into its unit representation.
func (v *Vec3Impl) MakeUnitVector() {
	l := v.Length()
	v.X = v.X / l
	v.Y = v.Y / l
	v.Z = v.Z / l
}

// Add returns the sum of two or more vectors.
func Add(v1 *Vec3Impl, args ...*Vec3Impl) *Vec3Impl {
	sum := &Vec3Impl{
		X: v1.X,
		Y: v1.Y,
		Z: v1.Z,
	}

	for i := range args {
		sum.X += args[i].X
		sum.Y += args[i].Y
		sum.Z += args[i].Z
	}

	return sum
}

// Sub returns the subtraction of two or more vectors.
func Sub(v1 *Vec3Impl, args ...*Vec3Impl) *Vec3Impl {
	res := &Vec3Impl{
		X: v1.X,
		Y: v1.Y,
		Z: v1.Z,
	}

	for i := range args {
		res.X -= args[i].X
		res.Y -= args[i].Y
		res.Z -= args[i].Z
	}

	return res
}

// Mul returns the multiplication of two vectors.
func Mul(v1 *Vec3Impl, v2 *Vec3Impl) *Vec3Impl {
	return &Vec3Impl{
		X: v1.X * v2.X,
		Y: v1.Y * v2.Y,
		Z: v1.Z * v2.Z,
	}
}

// Div returns the division of two vectors.
func Div(v1 *Vec3Impl, v2 *Vec3Impl) *Vec3Impl {
	return &Vec3Impl{
		X: v1.X / v2.X,
		Y: v1.Y / v2.Y,
		Z: v1.Z / v2.Z,
	}
}

// ScalarMul returns the scalar multiplication of the given vector and scalar values.
func ScalarMul(v1 *Vec3Impl, t float32) *Vec3Impl {
	return &Vec3Impl{
		X: v1.X * t,
		Y: v1.Y * t,
		Z: v1.Z * t,
	}
}

// ScalarMul returns the scalar division of the given vector and scalar values.
func ScalarDiv(v1 *Vec3Impl, t float32) *Vec3Impl {
	return &Vec3Impl{
		X: v1.X / t,
		Y: v1.Y / t,
		Z: v1.Z / t,
	}
}

// Dot computes the dot product of the two supplied vectors.
func Dot(v1 *Vec3Impl, v2 *Vec3Impl) float32 {
	return (v1.X * v2.X) + (v1.Y * v2.Y) + (v1.Z * v2.Z)
}

// Cross computes the cross product of the two supplied vectors.
func Cross(v1 *Vec3Impl, v2 *Vec3Impl) *Vec3Impl {
	return &Vec3Impl{
		X: (v1.Y * v2.Z) - (v1.Z * v2.Y),
		Y: -((v1.X * v2.Z) - (v1.Z * v2.X)),
		Z: (v1.X * v2.Y) - (v1.Y * v2.X),
	}
}

// UnitVector returns a unit vector representation of the supplied vector.
func UnitVector(v *Vec3Impl) *Vec3Impl {
	return ScalarDiv(v, v.Length())
}

// RandomCosineDirection returns a vector with a random cosine direction.
func RandomCosineDirection(random *fastrandom.XorShift) *Vec3Impl {
	r1 := random.Float32()
	r2 := random.Float32()
	z := float32(math.Sqrt(float64(1 - r2)))
	phi := 2 * math.Pi * r1
	x := float32(math.Cos(float64(phi)) * 2 * math.Sqrt(float64(r2)))
	y := float32(math.Sin(float64(phi)) * 2 * math.Sqrt(float64(r2)))
	return &Vec3Impl{X: x, Y: y, Z: z}
}

// RandomToSphere returns a new random sphere of the given radius at the given distance.
func RandomToSphere(radius float32, distanceSquared float32, random *fastrandom.XorShift) *Vec3Impl {
	r1 := random.Float32()
	r2 := random.Float32()
	z := 1 + r2*(float32(math.Sqrt(float64(1-radius*radius/distanceSquared)))-1)
	phi := 2 * math.Pi * r1
	x := float32(math.Cos(float64(phi)) * math.Sqrt(float64(1-z*z)))
	y := float32(math.Sin(float64(phi)) * math.Sqrt(float64(1-z*z)))
	return &Vec3Impl{X: x, Y: y, Z: z}
}

// DeNAN ensures that the vector elements are numbers.
func DeNAN(v *Vec3Impl) *Vec3Impl {
	x := v.X
	y := v.Y
	z := v.Z
	if math.IsNaN(float64(x)) || math.IsInf(float64(x), -1) || math.IsInf(float64(x), 1) {
		x = 0
	}

	if math.IsNaN(float64(y)) || math.IsInf(float64(y), -1) || math.IsInf(float64(y), 1) {
		y = 0
	}

	if math.IsNaN(float64(z)) || math.IsInf(float64(z), -1) || math.IsInf(float64(z), 1) {
		z = 0
	}

	return &Vec3Impl{X: x, Y: y, Z: z}
}

// Min3 returns a new vector with the minimum coordinates among the supplied ones.
func Min3(v0 *Vec3Impl, v1 *Vec3Impl, v2 *Vec3Impl) *Vec3Impl {
	xMin := float32(math.MaxFloat32)
	yMin := float32(math.MaxFloat32)
	zMin := float32(math.MaxFloat32)

	if v0.X < xMin {
		xMin = v0.X
	}

	if v1.X < xMin {
		xMin = v1.X
	}

	if v2.X < xMin {
		xMin = v2.X
	}

	if v0.Y < yMin {
		yMin = v0.Y
	}

	if v1.Y < yMin {
		yMin = v1.Y
	}

	if v2.Y < yMin {
		yMin = v2.Y
	}

	if v0.Z < zMin {
		zMin = v0.Z
	}

	if v1.Z < zMin {
		zMin = v1.Z
	}

	if v2.Z < zMin {
		zMin = v2.Z
	}

	return &Vec3Impl{X: xMin, Y: yMin, Z: zMin}
}

// Max3 returns a new vector with the maximum coordinates among the supplied ones.
func Max3(v0 *Vec3Impl, v1 *Vec3Impl, v2 *Vec3Impl) *Vec3Impl {
	xMax := float32(-math.MaxFloat32)
	yMax := float32(-math.MaxFloat32)
	zMax := float32(-math.MaxFloat32)

	if v0.X > xMax {
		xMax = v0.X
	}

	if v1.X > xMax {
		xMax = v1.X
	}

	if v2.X > xMax {
		xMax = v2.X
	}

	if v0.Y > yMax {
		yMax = v0.Y
	}

	if v1.Y > yMax {
		yMax = v1.Y
	}

	if v2.Y > yMax {
		yMax = v2.Y
	}

	if v0.Z > zMax {
		zMax = v0.Z
	}

	if v1.Z > zMax {
		zMax = v1.Z
	}

	if v2.Z > zMax {
		zMax = v2.Z
	}

	return &Vec3Impl{X: xMax, Y: yMax, Z: zMax}
}

// Lerp performs a linear interpolation between the two provided vectors.
func Lerp(v0, v1 *Vec3Impl, t float32) *Vec3Impl {
	return &Vec3Impl{
		X: (1-t)*v0.X + t*v1.X,
		Y: (1-t)*v0.Y + t*v1.Y,
		Z: (1-t)*v0.Z + t*v1.Z,
	}
}

// Equals returns whether two vectors are the same.
func Equals(v0, v1 *Vec3Impl) bool {
	return v0.X == v1.X &&
		v0.Y == v1.Y &&
		v0.Z == v1.Z
}
