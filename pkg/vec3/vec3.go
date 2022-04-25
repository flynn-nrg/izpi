// Package vec3 provides utility functions to work with vectors.
package vec3

import (
	"math"
	"math/rand"
)

// Vec3Impl defines a vector with its position and colour.
type Vec3Impl struct {
	X float64
	Y float64
	Z float64
	R float64
	G float64
	B float64
}

// Length returns the length of this vector.
func (v *Vec3Impl) Length() float64 {
	return math.Sqrt((v.X * v.X) + (v.Y * v.Y) + (v.Z * v.Z))
}

// SquaredLength returns the squared length of this vector.
func (v *Vec3Impl) SquaredLength() float64 {
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
func ScalarMul(v1 *Vec3Impl, t float64) *Vec3Impl {
	return &Vec3Impl{
		X: v1.X * t,
		Y: v1.Y * t,
		Z: v1.Z * t,
	}
}

// ScalarMul returns the scalar division of the given vector and scalar values.
func ScalarDiv(v1 *Vec3Impl, t float64) *Vec3Impl {
	return &Vec3Impl{
		X: v1.X / t,
		Y: v1.Y / t,
		Z: v1.Z / t,
	}
}

// Dot computes the dot product of the two supplied vectors.
func Dot(v1 *Vec3Impl, v2 *Vec3Impl) float64 {
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
func RandomCosineDirection() *Vec3Impl {
	r1 := rand.Float64()
	r2 := rand.Float64()
	z := math.Sqrt(1 - r2)
	phi := 2 * math.Pi * r1
	x := math.Cos(phi) * 2 * math.Sqrt(r2)
	y := math.Sin(phi) * 2 * math.Sqrt(r2)
	return &Vec3Impl{X: x, Y: y, Z: z}
}

// RandomToSphere returns a new random sphere of the given radius at the given distance.
func RandomToSphere(radius float64, distanceSquared float64) *Vec3Impl {
	r1 := rand.Float64()
	r2 := rand.Float64()
	z := 1 + r2*(math.Sqrt(1-radius*radius/distanceSquared)-1)
	phi := 2 * math.Pi * r1
	x := math.Cos(phi) * math.Sqrt(1-z*z)
	y := math.Sin(phi) * math.Sqrt(1-z*z)
	return &Vec3Impl{X: x, Y: y, Z: z}
}

// DeNAN ensures that the vector elements are numbers.
func DeNAN(v *Vec3Impl) *Vec3Impl {
	x := v.X
	y := v.Y
	z := v.Z
	if math.IsNaN(x) || math.IsInf(x, -1) || math.IsInf(x, 1) {
		x = 0
	}

	if math.IsNaN(y) || math.IsInf(y, -1) || math.IsInf(y, 1) {
		y = 0
	}

	if math.IsNaN(z) || math.IsInf(z, -1) || math.IsInf(z, 1) {
		z = 0
	}

	return &Vec3Impl{X: x, Y: y, Z: z}
}
