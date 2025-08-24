package material

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/onb"
	"github.com/flynn-nrg/izpi/internal/pdf"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/scatterrecord"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance.
var _ Material = (*PBR)(nil)

// PBR represents a physically based rendering material.
type PBR struct {
	nonEmitter
	nonPathLength
	nonWorldSetter
	albedo         texture.Texture
	spectralAlbedo texture.SpectralTexture
	normalMap      texture.Texture
	roughness      texture.Texture
	metalness      texture.Texture
	sss            texture.Texture // Subsurface scattering strength
	sssRadius      float32         // Subsurface scattering radius
}

// NewPBR returns a new PBR material with the supplied textures.
func NewPBR(albedo, normalMap, roughness, metalness, sss texture.Texture, sssRadius float32) *PBR {
	return &PBR{
		albedo:    albedo,
		normalMap: normalMap,
		roughness: roughness,
		metalness: metalness,
		sss:       sss,
		sssRadius: sssRadius,
	}
}

// NewPBRWithSpectralAlbedo returns a new PBR material with spectral albedo support.
func NewPBRWithSpectralAlbedo(albedo texture.Texture, spectralAlbedo texture.SpectralTexture, normalMap, roughness, metalness, sss texture.Texture, sssRadius float32) *PBR {
	return &PBR{
		albedo:         albedo,
		spectralAlbedo: spectralAlbedo,
		normalMap:      normalMap,
		roughness:      roughness,
		metalness:      metalness,
		sss:            sss,
		sssRadius:      sssRadius,
	}
}

// Scatter computes how the ray bounces off the surface of a PBR material.
func (pbr *PBR) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.XorShift) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	albedo := pbr.albedo.Value(hr.U(), hr.V(), hr.P())

	// Handle normal map - convert from tangent space to world space
	var normal *vec3.Vec3Impl
	if pbr.normalMap != nil {
		normalAtUV := pbr.normalMap.Value(hr.U(), hr.V(), hr.P())
		tangentNormal := &vec3.Vec3Impl{
			X: 2.0*normalAtUV.X - 1.0,
			Y: 2.0*normalAtUV.Y - 1.0,
			Z: normalAtUV.Z,
		}

		n := hr.Normal()
		t := vec3.Cross(n, &vec3.Vec3Impl{X: 0, Y: 1, Z: 0})
		if vec3.Dot(t, t) < 0.001 {
			t = vec3.Cross(n, &vec3.Vec3Impl{X: 1, Y: 0, Z: 0})
		}

		t.MakeUnitVector()
		b := vec3.Cross(n, t)
		b.MakeUnitVector()

		// Transform normal from tangent space to world space
		normal = &vec3.Vec3Impl{
			X: t.X*tangentNormal.X + b.X*tangentNormal.Y + n.X*tangentNormal.Z,
			Y: t.Y*tangentNormal.X + b.Y*tangentNormal.Y + n.Y*tangentNormal.Z,
			Z: t.Z*tangentNormal.X + b.Z*tangentNormal.Y + n.Z*tangentNormal.Z,
		}
		normal.MakeUnitVector()
	} else {
		normal = hr.Normal()
	}

	var roughness *vec3.Vec3Impl
	if pbr.roughness != nil {
		roughness = pbr.roughness.Value(hr.U(), hr.V(), hr.P())
	} else {
		roughness = &vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 0.5} // Default to medium roughness
	}

	var metalness *vec3.Vec3Impl
	if pbr.metalness != nil {
		metalness = pbr.metalness.Value(hr.U(), hr.V(), hr.P())
	} else {
		metalness = &vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 0.0} // Default to non-metallic
	}

	// Calculate average values
	roughnessValue := (roughness.X + roughness.Y + roughness.Z) / 3.0
	metalnessValue := (metalness.X + metalness.Y + metalness.Z) / 3.0

	// Create ONB for local space calculations
	uvw := onb.New()
	uvw.BuildFromW(normal)

	// Calculate reflection vector for specular component
	reflected := reflect(vec3.UnitVector(r.Direction()), normal)

	// Balanced PBR scattering logic using roughness to control specular probability
	var finalDir *vec3.Vec3Impl
	var isSpecular bool

	// Calculate Fresnel effect (simplified)
	cosTheta := float32(math.Abs(float64(vec3.Dot(vec3.UnitVector(r.Direction()), normal))))
	fresnel := float32(0.04 + (1.0-0.04)*math.Pow(float64(1.0-cosTheta), 5.0))

	// Adjust fresnel based on metalness
	fresnel = fresnel + (metalnessValue * 0.5)

	// Use roughness to control specular probability
	// Lower roughness = higher chance of specular reflection
	specularProbability := fresnel * (1.0 - roughnessValue)

	if random.Float32() < specularProbability {
		// Specular reflection
		roughnessFactor := float32(math.Max(0.01, float64(roughnessValue*0.3)))
		randomDir := randomInUnitSphere(random)
		specularDir := vec3.Add(reflected, vec3.ScalarMul(randomDir, roughnessFactor))
		finalDir = vec3.UnitVector(specularDir)
		isSpecular = true
	} else {
		// Diffuse reflection
		diffuseDir := uvw.Local(vec3.RandomCosineDirection(random))
		finalDir = vec3.UnitVector(diffuseDir)
		isSpecular = false
	}

	scattered := ray.New(hr.P(), finalDir, r.Time())

	// Create PDF for importance sampling
	// TODO: Use a more accurate PDF for PBR materials.
	pdf := pdf.NewCosine(normal)

	scatterRecord := scatterrecord.New(scattered, isSpecular, albedo, nil, nil, nil, pdf)
	return scattered, scatterRecord, true
}

// SpectralScatter computes how the ray bounces off the surface of a PBR material with spectral properties.
func (pbr *PBR) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.XorShift) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	// Use spectral albedo if available, otherwise fall back to RGB albedo
	lambda := r.Lambda()
	albedo := pbr.SpectralAlbedo(hr.U(), hr.V(), lambda, hr.P())

	// Handle normal map - convert from tangent space to world space
	var normal *vec3.Vec3Impl
	if pbr.normalMap != nil {
		normalAtUV := pbr.normalMap.Value(hr.U(), hr.V(), hr.P())
		tangentNormal := &vec3.Vec3Impl{
			X: 2.0*normalAtUV.X - 1.0,
			Y: 2.0*normalAtUV.Y - 1.0,
			Z: normalAtUV.Z,
		}

		n := hr.Normal()
		t := vec3.Cross(n, &vec3.Vec3Impl{X: 0, Y: 1, Z: 0})
		if vec3.Dot(t, t) < 0.001 {
			t = vec3.Cross(n, &vec3.Vec3Impl{X: 1, Y: 0, Z: 0})
		}

		t.MakeUnitVector()
		b := vec3.Cross(n, t)
		b.MakeUnitVector()

		// Transform normal from tangent space to world space
		normal = &vec3.Vec3Impl{
			X: t.X*tangentNormal.X + b.X*tangentNormal.Y + n.X*tangentNormal.Z,
			Y: t.Y*tangentNormal.X + b.Y*tangentNormal.Y + n.Y*tangentNormal.Z,
			Z: t.Z*tangentNormal.X + b.Z*tangentNormal.Y + n.Z*tangentNormal.Z,
		}
		normal.MakeUnitVector()
	} else {
		normal = hr.Normal()
	}

	var roughness *vec3.Vec3Impl
	if pbr.roughness != nil {
		roughness = pbr.roughness.Value(hr.U(), hr.V(), hr.P())
	} else {
		roughness = &vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 0.5} // Default to medium roughness
	}

	var metalness *vec3.Vec3Impl
	if pbr.metalness != nil {
		metalness = pbr.metalness.Value(hr.U(), hr.V(), hr.P())
	} else {
		metalness = &vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 0.0} // Default to non-metallic
	}

	// Calculate average values
	roughnessValue := (roughness.X + roughness.Y + roughness.Z) / 3.0
	metalnessValue := (metalness.X + metalness.Y + metalness.Z) / 3.0

	// Create ONB for local space calculations
	uvw := onb.New()
	uvw.BuildFromW(normal)

	// Calculate reflection vector for specular component
	reflected := reflect(vec3.UnitVector(r.Direction()), normal)

	// Balanced PBR scattering logic using roughness to control specular probability
	var finalDir *vec3.Vec3Impl
	var isSpecular bool

	// Calculate Fresnel effect (simplified)
	cosTheta := float32(math.Abs(float64(vec3.Dot(vec3.UnitVector(r.Direction()), normal))))
	fresnel := float32(0.04 + (1.0-0.04)*math.Pow(float64(1.0-cosTheta), 5.0))

	// Adjust fresnel based on metalness
	fresnel = fresnel + (metalnessValue * 0.5)

	// Use roughness to control specular probability
	// Lower roughness = higher chance of specular reflection
	specularProbability := fresnel * (1.0 - roughnessValue)

	if random.Float32() < specularProbability {
		// Specular reflection
		roughnessFactor := float32(math.Max(0.01, float64(roughnessValue*0.3)))
		randomDir := randomInUnitSphere(random)
		specularDir := vec3.Add(reflected, vec3.ScalarMul(randomDir, roughnessFactor))
		finalDir = vec3.UnitVector(specularDir)
		isSpecular = true
	} else {
		// Diffuse reflection
		diffuseDir := uvw.Local(vec3.RandomCosineDirection(random))
		finalDir = vec3.UnitVector(diffuseDir)
		isSpecular = false
	}

	scattered := ray.NewWithLambda(hr.P(), finalDir, r.Time(), lambda)

	// Create PDF for importance sampling
	// TODO: Use a more accurate PDF for PBR materials.
	pdf := pdf.NewCosine(normal)

	// Boost specular reflection brightness to match RGB reference
	var finalAlbedo float32
	if isSpecular {
		// Increase the brightness of specular reflections specifically
		finalAlbedo = albedo * 1.5 // 50% boost for specular reflections
	} else {
		finalAlbedo = albedo
	}

	scatterRecord := scatterrecord.NewSpectralScatterRecord(scattered, isSpecular, finalAlbedo, lambda, nil, 0.0, 0.0, pdf)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for PBR materials.
func (pbr *PBR) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float32 {
	cosine := vec3.Dot(hr.Normal(), vec3.UnitVector(scattered.Direction()))
	if cosine < 0 {
		cosine = 0
	}

	return cosine / math.Pi
}

// NormalMap() returns the normal map associated with this material.
func (pbr *PBR) NormalMap() texture.Texture {
	return pbr.normalMap
}

func (pbr *PBR) Albedo(u float32, v float32, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return pbr.albedo.Value(u, v, p)
}

// SpectralAlbedo returns the spectral albedo at the given wavelength.
func (pbr *PBR) SpectralAlbedo(u float32, v float32, lambda float32, p *vec3.Vec3Impl) float32 {
	if pbr.spectralAlbedo != nil {
		return pbr.spectralAlbedo.Value(u, v, lambda, p)
	}
	// Fallback to RGB albedo if no spectral albedo is provided
	// Use luminance-weighted average to preserve overall brightness
	rgbAlbedo := pbr.albedo.Value(u, v, p)
	return 0.299*rgbAlbedo.X + 0.587*rgbAlbedo.Y + 0.114*rgbAlbedo.Z
}
