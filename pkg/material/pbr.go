package material

import (
	"math"

	"github.com/flynn-nrg/izpi/pkg/fastrandom"
	"github.com/flynn-nrg/izpi/pkg/hitrecord"
	"github.com/flynn-nrg/izpi/pkg/onb"
	"github.com/flynn-nrg/izpi/pkg/pdf"
	"github.com/flynn-nrg/izpi/pkg/ray"
	"github.com/flynn-nrg/izpi/pkg/scatterrecord"
	"github.com/flynn-nrg/izpi/pkg/texture"
	"github.com/flynn-nrg/izpi/pkg/vec3"
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
	// Get material properties at hit point
	albedo := pbr.albedo.Value(hr.U(), hr.V(), hr.P())

	// Handle normal map - use surface normal if normal map is nil
	var normalAtUV *vec3.Vec3Impl
	if pbr.normalMap != nil {
		normalAtUV = pbr.normalMap.Value(hr.U(), hr.V(), hr.P())
	} else {
		normalAtUV = hr.Normal()
	}

	// Handle roughness texture - use default value if nil
	var roughness *vec3.Vec3Impl
	if pbr.roughness != nil {
		roughness = pbr.roughness.Value(hr.U(), hr.V(), hr.P())
	} else {
		roughness = &vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 0.5} // Default to medium roughness
	}

	// Handle metalness texture - use default value if nil
	var metalness *vec3.Vec3Impl
	if pbr.metalness != nil {
		metalness = pbr.metalness.Value(hr.U(), hr.V(), hr.P())
	} else {
		metalness = &vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 0.0} // Default to non-metallic
	}

	// Handle subsurface scattering texture - use default value if nil
	var sssStrength *vec3.Vec3Impl
	if pbr.sss != nil {
		sssStrength = pbr.sss.Value(hr.U(), hr.V(), hr.P())
	} else {
		sssStrength = &vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 0.0} // Default to no subsurface scattering
	}

	// Calculate average values
	roughnessValue := (roughness.X + roughness.Y + roughness.Z) / 3.0
	metalnessValue := (metalness.X + metalness.Y + metalness.Z) / 3.0
	sssValue := (sssStrength.X + sssStrength.Y + sssStrength.Z) / 3.0

	// Create ONB for local space calculations
	uvw := onb.New()
	uvw.BuildFromW(hr.Normal())

	// Calculate reflection vector for specular component
	reflected := reflect(vec3.UnitVector(r.Direction()), hr.Normal())

	// Calculate diffuse direction
	diffuseDir := uvw.Local(vec3.RandomCosineDirection(random))
	diffuseDir = vec3.UnitVector(diffuseDir)

	// Calculate specular direction with roughness
	roughnessFactor := roughnessValue * roughnessValue // Square roughness for more natural look
	randomDir := randomInUnitSphere(random)
	specularDir := vec3.Add(reflected, vec3.ScalarMul(randomDir, roughnessFactor))
	specularDir = vec3.UnitVector(specularDir)

	// Calculate subsurface scattering direction if enabled
	var sssDir *vec3.Vec3Impl
	if sssValue > 0 {
		// Generate a random point within the subsurface radius
		offset := vec3.ScalarMul(randomInUnitSphere(random), pbr.sssRadius)
		sssDir = vec3.UnitVector(vec3.Add(hr.P(), offset))
	}

	// Blend between diffuse, specular, and subsurface based on material properties
	t := smoothstep(0.0, 0.5, metalnessValue)
	scatteredDir := vec3.Add(
		vec3.ScalarMul(diffuseDir, 1.0-t),
		vec3.ScalarMul(specularDir, t),
	)

	// Add subsurface scattering contribution
	if sssValue > 0 {
		sssBlend := smoothstep(0.0, 1.0, sssValue)
		scatteredDir = vec3.Add(
			vec3.ScalarMul(scatteredDir, 1.0-sssBlend),
			vec3.ScalarMul(sssDir, sssBlend),
		)
	}

	scatteredDir = vec3.UnitVector(scatteredDir)
	scattered := ray.New(hr.P(), scatteredDir, r.Time())
	pdf := pdf.NewCosine(hr.Normal())
	scatterRecord := scatterrecord.New(scattered, t > 0.5, albedo, normalAtUV, roughness, metalness, pdf)
	return scattered, scatterRecord, true
}

// smoothstep performs smooth interpolation between 0 and 1
func smoothstep(edge0, edge1, x float64) float64 {
	// Clamp x to 0..1 range
	if x < edge0 {
		return 0
	}
	if x > edge1 {
		return 1
	}
	x = (x - edge0) / (edge1 - edge0)
	return x * x * (3 - 2*x)
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
