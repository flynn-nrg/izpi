// Package mat3 implements functions to work with 3x3 matrices.
package mat3

import "github.com/flynn-nrg/izpi/internal/vec3"

type Mat3 struct {
	A11 float64
	A12 float64
	A13 float64
	A21 float64
	A22 float64
	A23 float64
	A31 float64
	A32 float64
	A33 float64
}

// NewTBN returns a new matrix made from the supplied tagent, bitangent and normal vectors.
func NewTBN(tangent, bitangent, normal vec3.Vec3Impl) Mat3 {
	return Mat3{
		A11: tangent.X,
		A12: bitangent.X,
		A13: normal.X,
		A21: tangent.Y,
		A22: bitangent.Y,
		A23: normal.Y,
		A31: tangent.Z,
		A32: bitangent.Z,
		A33: normal.Z,
	}
}

// MatrixVectorMul returns the result of axv, where a is a matrix and v is a vector.
func MatrixVectorMul(a Mat3, v vec3.Vec3Impl) vec3.Vec3Impl {
	return vec3.Vec3Impl{
		X: a.A11*v.X + a.A12*v.Y + a.A13*v.Z,
		Y: a.A21*v.X + a.A22*v.Y + a.A23*v.Z,
		Z: a.A31*v.X + a.A32*v.Y + a.A33*v.Z,
	}
}
