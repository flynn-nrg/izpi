// Package hitabletarget implements the methods used to extract PDF data from hitables.
// This is done avoid a circular dependency between pdf, hitable and material.
package hitabletarget

import (
	"github.com/flynn-nrg/go-vfx/math32/fastrandom"
	"github.com/flynn-nrg/go-vfx/math32/vec3"
)

// HitableTarget defines the methods used to embed hitables in a PDF.
type HitableTarget interface {
	PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float32
	Random(o *vec3.Vec3Impl, random *fastrandom.XorShift) *vec3.Vec3Impl
}
