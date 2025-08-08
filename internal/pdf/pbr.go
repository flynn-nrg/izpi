package pdf

import (
	"math"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/onb"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// Ensure interface compliance

// PBR_PDF handles importance sampling for a PBR material.
type PBR_PDF struct {
	uvw         *onb.Onb
	incomingDir *vec3.Vec3Impl
	normal      *vec3.Vec3Impl
	roughness   float64
	metalness   float64
}

// NewPBR returns a new PBR PDF sampler.
func NewPBR(incomingDir, normal *vec3.Vec3Impl, roughness, metalness float64) *PBR_PDF {
	p := &PBR_PDF{
		uvw:         onb.New(),
		incomingDir: incomingDir,
		normal:      normal,
		roughness:   roughness,
		metalness:   metalness,
	}
	p.uvw.BuildFromW(normal)
	return p
}

// Generate a direction based on the PBR model.
func (p *PBR_PDF) Generate(random *fastrandom.LCG) *vec3.Vec3Impl {
	// Probabilistically choose between metallic and dielectric models.
	if random.Float64() < p.metalness {
		// --- Metallic Path: Purely specular ---
		reflected := vec3.Reflect(vec3.UnitVector(p.incomingDir), p.normal)
		fuzz := vec3.RandomInUnitSphere(random)
		return vec3.UnitVector(vec3.Add(reflected, vec3.ScalarMul(fuzz, p.roughness)))
	}

	// --- Dielectric Path: Mix of diffuse and specular ---
	cosTheta := math.Min(vec3.Dot(vec3.Negate(p.incomingDir), p.normal), 1.0)
	f0 := 0.04
	reflectance := f0 + (1.0-f0)*math.Pow(1.0-cosTheta, 5.0)

	if random.Float64() < reflectance {
		// Specular reflection
		reflected := vec3.Reflect(vec3.UnitVector(p.incomingDir), p.normal)
		fuzz := vec3.RandomInUnitSphere(random)
		return vec3.UnitVector(vec3.Add(reflected, vec3.ScalarMul(fuzz, p.roughness)))
	}

	// Diffuse reflection (Lambertian)
	return p.uvw.Local(vec3.RandomCosineDirection(random))
}

// Value returns the probability of generating the given direction.
func (p *PBR_PDF) Value(direction *vec3.Vec3Impl) float64 {
	unitDir := vec3.UnitVector(direction)
	cosTheta := vec3.Dot(p.normal, unitDir)

	if cosTheta <= 0 {
		return 0
	}

	// --- PDF for the Diffuse Lobe ---
	pdfDiffuse := cosTheta / math.Pi

	// --- PDF for the Specular Lobe (Cosine-Power Approximation) ---
	reflected := vec3.Reflect(vec3.UnitVector(p.incomingDir), p.normal)
	cosAlpha := vec3.Dot(reflected, unitDir)
	var pdfSpecular float64
	if cosAlpha > 0 {
		// Map roughness (0-1) to an exponent (high to low).
		// A roughness of 0 gives a near-infinite exponent (perfect mirror).
		// A roughness of 1 gives an exponent of 0 (nearly uniform lobe).
		clampedRoughness := math.Max(0.001, p.roughness) // Avoid division by zero
		exponent := 2.0/(clampedRoughness*clampedRoughness) - 2.0
		pdfSpecular = ((exponent + 1.0) / (2.0 * math.Pi)) * math.Pow(cosAlpha, exponent)
	}

	// --- Weigh the PDFs based on the material properties ---
	incomingCosTheta := math.Min(vec3.Dot(vec3.Negate(p.incomingDir), p.normal), 1.0)
	f0 := 0.04
	reflectance := f0 + (1.0-f0)*math.Pow(1.0-incomingCosTheta, 5.0)

	// Final PDF is the weighted average of the individual PDFs
	return p.metalness*pdfSpecular + (1.0-p.metalness)*((1.0-reflectance)*pdfDiffuse+reflectance*pdfSpecular)
}
