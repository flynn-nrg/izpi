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
	albedo      *vec3.Vec3Impl
	normal      *vec3.Vec3Impl
	roughness   *vec3.Vec3Impl
	metalness   *vec3.Vec3Impl
	pdf         pdf.PDF
}

// New returns an instance of a scatter record.
func New(specularRay ray.Ray, isSpecular bool,
	albedo, normal, roughness, metalness *vec3.Vec3Impl, pdf pdf.PDF) *ScatterRecord {
	return &ScatterRecord{
		specularRay: specularRay,
		isSpecular:  isSpecular,
		albedo:      albedo,
		normal:      normal,
		roughness:   roughness,
		metalness:   metalness,
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
	return sr.albedo
}

// Normal returns the normal for this material at this point.
func (sr *ScatterRecord) Normal() *vec3.Vec3Impl {
	return sr.normal
}

// Roughness returns the roughness value for this material at this point.
func (sr *ScatterRecord) Roughness() *vec3.Vec3Impl {
	return sr.albedo
}

// Metalness returns the metalness value for this material at this point.
func (sr *ScatterRecord) Metalness() *vec3.Vec3Impl {
	return sr.albedo
}

// PDF returns the PDF associated with this material.
func (sr *ScatterRecord) PDF() pdf.PDF {
	return sr.pdf
}
