// Package scatterrecord implements the scatter record.
package scatterrecord

import (
	"gitlab.com/flynn-nrg/izpi/pkg/pdf"
	"gitlab.com/flynn-nrg/izpi/pkg/ray"
	"gitlab.com/flynn-nrg/izpi/pkg/vec3"
)

// ScatterRecord represents a scatter record.
type ScatterRecord struct {
	specularRay ray.Ray
	isSpecular  bool
	attenuation *vec3.Vec3Impl
	pdf         pdf.PDF
}

// New returns an instance of a scatter record.
func New(specularRay ray.Ray, isSpecular bool, attenuation *vec3.Vec3Impl, pdf pdf.PDF) *ScatterRecord {
	return &ScatterRecord{
		specularRay: specularRay,
		isSpecular:  isSpecular,
		attenuation: attenuation,
		pdf:         pdf,
	}
}

// SpecularRay() returns the specular ray from this scatter record.
func (sr *ScatterRecord) SpecularRay() ray.Ray {
	return sr.specularRay
}

// IsSpecular() returns whether this material is specular.
func (sr *ScatterRecord) IsSpecular() bool {
	return sr.isSpecular
}

// Attenuation returns the attenuation value for this material.
func (sr *ScatterRecord) Attenuation() *vec3.Vec3Impl {
	return sr.attenuation
}

func (sr *ScatterRecord) PDF() pdf.PDF {
	return sr.pdf
}
