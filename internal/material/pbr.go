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

// PBR represents a physically based rendering material using a metallic-roughness workflow.
type PBR struct {
	nonEmitter
	nonPathLength
	nonWorldSetter
	albedo         texture.Texture
	spectralAlbedo texture.SpectralTexture
	normalMap      texture.Texture
	roughness      texture.Texture
	metalness      texture.Texture
	// TODO: Subsurface scattering is not yet implemented.
	sss       texture.Texture
	sssRadius float64
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

// NewPBRWithSpectralAlbedo returns a new PBR material with spectral albedo support.
func NewPBRWithSpectralAlbedo(albedo texture.Texture, spectralAlbedo texture.SpectralTexture, normalMap, roughness, metalness, sss texture.Texture, sssRadius float64) *PBR {
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
func (pbr *PBR) Scatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.ScatterRecord, bool) {
	albedo := pbr.Albedo(hr.U(), hr.V(), hr.P())
	metalness := pbr.metalnessValue(hr)
	roughness := pbr.roughnessValue(hr)
	normal := pbr.computeShadingNormal(hr)

	scatteredDir, isSpecular := pbr.scatterLogic(r.Direction(), normal, roughness, metalness, random)
	scatteredRay := ray.New(hr.P(), scatteredDir, r.Time())

	var attenuation *vec3.Vec3Impl
	// For metals, specular reflection is tinted by albedo. For dielectrics, it's white.
	// We use metalness to linearly interpolate between the two.
	if isSpecular {
		dielectricTint := &vec3.Vec3Impl{X: 1.0, Y: 1.0, Z: 1.0}
		attenuation = vec3.Lerp(dielectricTint, albedo, metalness)
	} else {
		// Diffuse is always tinted by albedo (only occurs for dielectrics).
		attenuation = albedo
	}

	// NOTE: For ideal importance sampling, the PDF object should be a mixture of diffuse
	// and specular lobes. Using a pure cosine PDF is only correct for the diffuse
	// component and will be inefficient (high variance) for sharp specular reflections.
	// A custom PBR PDF sampler would be a valuable future improvement.
	// pdf := pdf.NewCosine(normal)
	pdf := pdf.NewPBR(r.Direction(), normal, roughness, metalness)

	scatterRecord := scatterrecord.New(scatteredRay, isSpecular, attenuation, nil, nil, nil, pdf)
	return scatteredRay, scatterRecord, true
}

// SpectralScatter computes how the ray bounces off the surface with spectral properties.
func (pbr *PBR) SpectralScatter(r ray.Ray, hr *hitrecord.HitRecord, random *fastrandom.LCG) (*ray.RayImpl, *scatterrecord.SpectralScatterRecord, bool) {
	lambda := r.Lambda()
	albedo := pbr.SpectralAlbedo(hr.U(), hr.V(), lambda, hr.P())
	metalness := pbr.metalnessValue(hr)
	roughness := pbr.roughnessValue(hr)
	normal := pbr.computeShadingNormal(hr)

	scatteredDir, isSpecular := pbr.scatterLogic(r.Direction(), normal, roughness, metalness, random)
	scatteredRay := ray.NewWithLambda(hr.P(), scatteredDir, r.Time(), lambda)

	var finalAlbedo float64
	if isSpecular {
		// For metals, specular is tinted. For dielectrics, it's white (reflectance 1.0).
		dielectricTint := 1.0
		finalAlbedo = dielectricTint*(1.0-metalness) + albedo*metalness
	} else {
		finalAlbedo = albedo
	}
	// NOTE: The non-physical brightness boost has been removed for physical accuracy.
	// If spectral renders appear dim, investigate light source SPD and camera response functions.

	// See the note in the RGB Scatter method regarding the PDF sampler.
	//pdf := pdf.NewCosine(normal)
	pdf := pdf.NewPBR(r.Direction(), normal, roughness, metalness)

	scatterRecord := scatterrecord.NewSpectralScatterRecord(scatteredRay, isSpecular, finalAlbedo, lambda, nil, 0.0, 0.0, pdf)
	return scatteredRay, scatterRecord, true
}

// scatterLogic contains the unified PBR scattering model.
// It decides the outgoing direction based on a probabilistic metallic-roughness workflow.
func (pbr *PBR) scatterLogic(incomingDir, normal *vec3.Vec3Impl, roughness, metalness float64, random *fastrandom.LCG) (scatteredDir *vec3.Vec3Impl, isSpecular bool) {
	// A purely specular reflection vector.
	reflected := vec3.Reflect(vec3.UnitVector(incomingDir), normal)

	// Probabilistically choose between metallic and dielectric models.
	if random.Float64() < metalness {
		// --- Metallic Path ---
		// Metals are purely specular and absorb non-reflected light.
		isSpecular = true
		fuzz := vec3.RandomInUnitSphere(random)
		scatteredDir = vec3.UnitVector(vec3.Add(reflected, vec3.ScalarMul(fuzz, roughness)))
	} else {
		// --- Dielectric Path ---
		// Dielectrics mix specular and diffuse reflections based on the Fresnel effect.
		cosTheta := math.Min(vec3.Dot(vec3.Negate(incomingDir), normal), 1.0)
		f0 := 0.04 // Base reflectance for common dielectrics.
		reflectance := f0 + (1.0-f0)*math.Pow(1.0-cosTheta, 5.0)

		if random.Float64() < reflectance {
			// Specular reflection
			isSpecular = true
			fuzz := vec3.RandomInUnitSphere(random)
			scatteredDir = vec3.UnitVector(vec3.Add(reflected, vec3.ScalarMul(fuzz, roughness)))
		} else {
			// Diffuse reflection (Lambertian)
			isSpecular = false
			uvw := onb.New()
			uvw.BuildFromW(normal)
			scatteredDir = uvw.Local(vec3.RandomCosineDirection(random))
		}
	}

	// Ensure the scattered ray is in the same hemisphere as the normal.
	if vec3.Dot(scatteredDir, normal) <= 0 {
		if isSpecular {
			return reflected, true // Fallback to perfect reflection.
		}
		return normal, false // Fallback to the normal for diffuse.
	}

	return scatteredDir, isSpecular
}

// ScatteringPDF evaluates the probability distribution function for a given scattered direction.
// This is used for Multiple Importance Sampling (MIS) when sampling light sources.
// This PDF must match the sampling logic in scatterLogic.
func (pbr *PBR) ScatteringPDF(r ray.Ray, hr *hitrecord.HitRecord, scattered ray.Ray) float64 {
	normal := pbr.computeShadingNormal(hr)
	metalness := pbr.metalnessValue(hr)
	// roughness := pbr.roughnessValue(hr) // Needed for a proper specular PDF.

	incomingDir := vec3.UnitVector(r.Direction())
	scatteredDir := vec3.UnitVector(scattered.Direction())

	cosTheta := math.Min(vec3.Dot(vec3.Negate(incomingDir), normal), 1.0)

	// PDF for the diffuse (Lambertian) lobe.
	pdfDiffuse := math.Max(0, vec3.Dot(normal, scatteredDir)/math.Pi)

	// PDF for the specular lobe.
	// For a true microfacet model (e.g., GGX), this would be the GGX distribution function D().
	// For this simplified model, we treat it as a delta function (PDF=0) for MIS purposes,
	// as its probability is only non-zero in a single direction for a given roughness.
	pdfSpecular := 0.0

	// Probability of following the dielectric path.
	probDielectric := 1.0 - metalness
	// Reflectance probability within the dielectric path (Fresnel).
	f0 := 0.04
	probReflectance := f0 + (1.0-f0)*math.Pow(1.0-cosTheta, 5.0)

	// The final PDF is the weighted sum of the PDFs of each possible path.
	pdf := metalness*pdfSpecular +
		probDielectric*(probReflectance*pdfSpecular+(1.0-probReflectance)*pdfDiffuse)

	return pdf
}

// computeShadingNormal calculates the final normal for shading, applying the normal map if present.
func (pbr *PBR) computeShadingNormal(hr *hitrecord.HitRecord) *vec3.Vec3Impl {
	if pbr.normalMap == nil {
		return hr.Normal()
	}

	// Unpack the tangent-space normal from the texture.
	normalAtUV := pbr.normalMap.Value(hr.U(), hr.V(), hr.P())
	tangentNormal := &vec3.Vec3Impl{
		X: 2.0*normalAtUV.X - 1.0,
		Y: 2.0*normalAtUV.Y - 1.0,
		Z: normalAtUV.Z,
	}

	// Create a robust orthonormal basis (TBN matrix).
	n := hr.Normal()
	// Pick an arbitrary vector that is not parallel to n.
	var up *vec3.Vec3Impl
	if math.Abs(n.X) > 0.9 {
		up = &vec3.Vec3Impl{X: 0, Y: 1, Z: 0}
	} else {
		up = &vec3.Vec3Impl{X: 1, Y: 0, Z: 0}
	}
	t := vec3.UnitVector(vec3.Cross(up, n))
	b := vec3.Cross(n, t)

	// Transform normal from tangent space to world space.
	return vec3.UnitVector(&vec3.Vec3Impl{
		X: t.X*tangentNormal.X + b.X*tangentNormal.Y + n.X*tangentNormal.Z,
		Y: t.Y*tangentNormal.X + b.Y*tangentNormal.Y + n.Y*tangentNormal.Z,
		Z: t.Z*tangentNormal.X + b.Z*tangentNormal.Y + n.Z*tangentNormal.Z,
	})
}

func (pbr *PBR) roughnessValue(hr *hitrecord.HitRecord) float64 {
	if pbr.roughness == nil {
		return 0.5 // Default to medium roughness
	}
	val := pbr.roughness.Value(hr.U(), hr.V(), hr.P())
	// Average RGB for roughness, assuming a grayscale map.
	return (val.X + val.Y + val.Z) / 3.0
}

func (pbr *PBR) metalnessValue(hr *hitrecord.HitRecord) float64 {
	if pbr.metalness == nil {
		return 0.0 // Default to non-metallic
	}
	val := pbr.metalness.Value(hr.U(), hr.V(), hr.P())
	// Average RGB for metalness, assuming a grayscale map.
	return (val.X + val.Y + val.Z) / 3.0
}

// Albedo returns the RGB albedo color at a given point.
func (pbr *PBR) Albedo(u, v float64, p *vec3.Vec3Impl) *vec3.Vec3Impl {
	return pbr.albedo.Value(u, v, p)
}

// SpectralAlbedo returns the spectral albedo value at a given point and wavelength.
func (pbr *PBR) SpectralAlbedo(u, v, lambda float64, p *vec3.Vec3Impl) float64 {
	if pbr.spectralAlbedo != nil {
		return pbr.spectralAlbedo.Value(u, v, lambda, p)
	}
	// Fallback to luminance of RGB albedo if no spectral texture is provided.
	rgbAlbedo := pbr.albedo.Value(u, v, p)
	return 0.299*rgbAlbedo.X + 0.587*rgbAlbedo.Y + 0.114*rgbAlbedo.Z
}

// NormalMap returns the normal map texture associated with this material.
func (pbr *PBR) NormalMap() texture.Texture {
	return pbr.normalMap
}
