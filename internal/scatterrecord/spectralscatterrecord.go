// Package scatterrecord implements the scatter record.
package scatterrecord

import (
	"github.com/flynn-nrg/izpi/internal/pdf"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// SpectralScatterRecord represents a spectral scatter record.
type SpectralScatterRecord struct {
	specularRay ray.Ray
	isSpecular  bool
	albedo      float32 // Spectral albedo at wavelength lambda
	lambda      float32 // Wavelength in nanometers
	normal      *vec3.Vec3Impl
	roughness   float32 // Spectral roughness at wavelength lambda
	metalness   float32 // Spectral metalness at wavelength lambda
	pdf         pdf.PDF
}

// New returns an instance of a spectral scatter record.
func NewSpectralScatterRecord(specularRay ray.Ray, isSpecular bool,
	albedo float32, lambda float32, normal *vec3.Vec3Impl, roughness float32, metalness float32, pdf pdf.PDF) *SpectralScatterRecord {
	return &SpectralScatterRecord{
		specularRay: specularRay,
		isSpecular:  isSpecular,
		albedo:      albedo,
		lambda:      lambda,
		normal:      normal,
		roughness:   roughness,
		metalness:   metalness,
		pdf:         pdf,
	}
}

// SpecularRay() returns the specular ray from this scatter record.
func (ssr *SpectralScatterRecord) SpecularRay() ray.Ray {
	return ssr.specularRay
}

// IsSpecular() returns whether this material is specular.
func (ssr *SpectralScatterRecord) IsSpecular() bool {
	return ssr.isSpecular
}

// Attenuation returns the spectral attenuation value for this material at wavelength lambda.
func (ssr *SpectralScatterRecord) Attenuation() float32 {
	return ssr.albedo
}

// Wavelength returns the wavelength (lambda) associated with this scatter record.
func (ssr *SpectralScatterRecord) Wavelength() float32 {
	return ssr.lambda
}

// Normal returns the normal for this material at this point.
func (ssr *SpectralScatterRecord) Normal() *vec3.Vec3Impl {
	return ssr.normal
}

// Roughness returns the spectral roughness value for this material at wavelength lambda.
func (ssr *SpectralScatterRecord) Roughness() float32 {
	return ssr.roughness
}

// Metalness returns the spectral metalness value for this material at wavelength lambda.
func (ssr *SpectralScatterRecord) Metalness() float32 {
	return ssr.metalness
}

// PDF returns the PDF associated with this material.
func (ssr *SpectralScatterRecord) PDF() pdf.PDF {
	return ssr.pdf
}
