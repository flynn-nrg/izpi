// Package hitabletarget implements the methods used to extract PDF data from hitables.
// This is done avoid a circular dependency between pdf, hitable and material.
package hitabletarget

import "github.com/flynn-nrg/izpi/pkg/vec3"

// HitableTarget defines the methods used to embed hitables in a PDF.
type HitableTarget interface {
	PDFValue(o *vec3.Vec3Impl, v *vec3.Vec3Impl) float64
	Random(o *vec3.Vec3Impl) *vec3.Vec3Impl
}
