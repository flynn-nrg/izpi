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

	// Handle normal map - convert from tangent space to world space
	var normal *vec3.Vec3Impl
	if pbr.normalMap != nil {
		normalAtUV := pbr.normalMap.Value(hr.U(), hr.V(), hr.P())
		// Convert from [0,1] to [-1,1] range
		tangentNormal := &vec3.Vec3Impl{
			X: 2.0*normalAtUV.X - 1.0,
			Y: 2.0*normalAtUV.Y - 1.0,
			Z: normalAtUV.Z,
		}
		// Create TBN matrix
		N := hr.Normal()
		T := vec3.Cross(N, &vec3.Vec3Impl{X: 0, Y: 1, Z: 0})
		if vec3.Dot(T, T) < 0.001 {
			T = vec3.Cross(N, &vec3.Vec3Impl{X: 1, Y: 0, Z: 0})
		}
		T.MakeUnitVector()
		B := vec3.Cross(N, T)
		B.MakeUnitVector()

		// Transform normal from tangent space to world space
		normal = &vec3.Vec3Impl{
			X: T.X*tangentNormal.X + B.X*tangentNormal.Y + N.X*tangentNormal.Z,
			Y: T.Y*tangentNormal.X + B.Y*tangentNormal.Y + N.Y*tangentNormal.Z,
			Z: T.Z*tangentNormal.X + B.Z*tangentNormal.Y + N.Z*tangentNormal.Z,
		}
		normal.MakeUnitVector()
	} else {
		normal = hr.Normal()
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

	// Calculate average values
	roughnessValue := (roughness.X + roughness.Y + roughness.Z) / 3.0
	metalnessValue := (metalness.X + metalness.Y + metalness.Z) / 3.0

	// For brushed metal, roughness should affect the metallic appearance
	// Higher roughness (brushed areas) should reduce the metallic reflection even more
	adjustedMetalness := metalnessValue * (1.0 - roughnessValue*0.8) * 0.7 // Stronger roughness influence and overall reduction

	// Create ONB for local space calculations
	uvw := onb.New()
	uvw.BuildFromW(normal)

	// Calculate reflection vector for specular component
	reflected := reflect(vec3.UnitVector(r.Direction()), normal)

	// For brushed metal, we want more controlled roughness
	roughnessFactor := math.Max(0.4, roughnessValue) // Even higher minimum roughness
	randomDir := randomInUnitSphere(random)
	specularDir := vec3.Add(reflected, vec3.ScalarMul(randomDir, roughnessFactor))
	specularDir = vec3.UnitVector(specularDir)

	// Calculate diffuse direction
	diffuseDir := uvw.Local(vec3.RandomCosineDirection(random))
	diffuseDir = vec3.UnitVector(diffuseDir)

	// Blend between diffuse and specular based on adjusted metalness and roughness
	var scatteredDir *vec3.Vec3Impl
	if random.Float64() < adjustedMetalness {
		// Metallic reflection with increased diffuse mixing
		if random.Float64() < roughnessValue*0.5 { // Increased diffuse probability in rough areas
			scatteredDir = diffuseDir
		} else {
			scatteredDir = specularDir
		}
	} else {
		// Non-metallic parts heavily favor diffuse
		if random.Float64() < 0.02+(1.0-roughnessValue)*0.15 { // Very low base specularity
			scatteredDir = specularDir
		} else {
			scatteredDir = diffuseDir
		}
	}

	scattered := ray.New(hr.P(), scatteredDir, r.Time())
	pdf := pdf.NewCosine(normal)

	scatterRecord := scatterrecord.New(scattered, metalnessValue > 0.5, albedo, normal, roughness, metalness, pdf)
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
