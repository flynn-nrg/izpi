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

type PBR struct {
	nonEmitter
	albedo    texture.Texture
	normalMap texture.Texture
	roughness texture.Texture
	metalness texture.Texture
	sss       texture.Texture // Subsurface scattering strength
	sssRadius float64         // Subsurface scattering radius
}

// NewPBR returns a new PBR material with the supplied textures.
func NewPBR(albedo, normalMap, roughness, metalness, sss texture.Texture, sssRadius float64) *PBR {
	return &PBR{
		albedo:    albedo,
		normalMap: normalMap,
		roughness: roughness,
		metalness: metalness,
		sss:       sss,
		sssRadius: sssRadius,
	}
}

// Scatter computes how the ray bounces off the surface of a PBR material.
func (pbr *PBR) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
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

	// Adjust for better appearance
	adjustedMetalness := metalnessValue * (1.0 - roughnessValue*0.5)

	// Create ONB for local space calculations
	uvw := onb.New()
	uvw.BuildFromW(normal)

	// Calculate reflection vector for specular component
	reflected := reflect(vec3.UnitVector(r.Direction()), normal)

	// Adjust roughness factor based on material type
	roughnessFactor := math.Max(0.2, roughnessValue)
	randomDir := randomInUnitSphere(random)
	specularDir := vec3.Add(reflected, vec3.ScalarMul(randomDir, roughnessFactor))
	specularDir = vec3.UnitVector(specularDir)

	// Calculate diffuse direction
	diffuseDir := uvw.Local(vec3.RandomCosineDirection(random))
	diffuseDir = vec3.UnitVector(diffuseDir)

	isSpecular := false

	// Blend between diffuse and specular based on adjusted metalness and roughness
	var scatteredDir *vec3.Vec3Impl
	if random.Float64() < adjustedMetalness {
		if random.Float64() < roughnessValue*0.3 {
			scatteredDir = diffuseDir
			isSpecular = false
		} else {
			scatteredDir = specularDir
			isSpecular = true
		}
	} else {
		// Non-metallic parts with material-specific specularity
		// Base specular probability increases with smoothness (1-roughness)
		// Range from 5% to 15% specular probability for non-metallic materials
		specularProb := 0.05 + (1.0-roughnessValue)*0.10
		if random.Float64() < specularProb {
			scatteredDir = specularDir
			isSpecular = true
		} else {
			scatteredDir = diffuseDir
			isSpecular = false
		}
	}

	scattered := ray.New(hr.P(), scatteredDir, r.Time())
	pdf := pdf.NewCosine(normal)

	scatterRecord := scatterrecord.New(scattered, isSpecular, albedo, normal, roughness, metalness, pdf)
	return scattered, scatterRecord, true
}

// ScatteringPDF implements the probability distribution function for PBR materials.
func (pbr *PBR) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
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

func (pbr *PBR) Albedo(u float64, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return pbr.albedo.Value(u, v, p)
}
