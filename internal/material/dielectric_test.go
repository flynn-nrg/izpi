package material

import (
	"testing"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	"github.com/flynn-nrg/izpi/internal/hitrecord"
	"github.com/flynn-nrg/izpi/internal/ray"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

func TestNewSpectralColoredDielectric(t *testing.T) {
	spectralRefIdx := texture.NewSpectralConstant(1.5, 550.0, 50.0)
	spectralAbsorptionCoeff := texture.NewSpectralConstant(0.5, 480.0, 60.0)

	dielectric := NewSpectralColoredDielectric(spectralRefIdx, spectralAbsorptionCoeff)

	if dielectric.spectralRefIdx == nil {
		t.Error("Expected spectral refractive index to be set")
	}

	if dielectric.spectralAbsorptionCoeff == nil {
		t.Error("Expected spectral absorption coefficient to be set")
	}
}

func TestBeerLambertAttenuation(t *testing.T) {
	spectralAbsorptionCoeff := texture.NewSpectralConstant(0.5, 480.0, 60.0)
	dielectric := NewSpectralColoredDielectric(nil, spectralAbsorptionCoeff)

	// Test at the peak wavelength (480nm)
	attenuation := dielectric.calculateBeerLambertAttenuation(1.0, 480.0, 0.0, 0.0, vec3.Vec3Impl{})
	expected := 0.6065 // exp(-0.5 * 1.0)

	if abs(attenuation-expected) > 0.001 {
		t.Errorf("Expected attenuation %f, got %f", expected, attenuation)
	}

	// Test at a wavelength far from the peak (should have lower absorption)
	attenuation2 := dielectric.calculateBeerLambertAttenuation(1.0, 700.0, 0.0, 0.0, vec3.Vec3Impl{})
	if attenuation2 <= attenuation {
		t.Error("Expected higher attenuation (less absorption) at wavelength far from peak")
	}
}

func TestPathLengthCalculation(t *testing.T) {
	dielectric := NewDielectric(1.5)

	// Create a hit record
	hitPoint := vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 0.5}
	normal := vec3.Vec3Impl{X: 0.577, Y: 0.577, Z: 0.577} // Normalized
	hr := hitrecord.New(1.0, 0.0, 0.0, hitPoint, normal)

	// Create a ray
	origin := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: -2.0}
	direction := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	r := ray.New(origin, direction, 0.0)

	// Create a scattered ray
	scattered := ray.New(origin, direction, 0.0)

	// Create a mock world geometry for testing
	mockWorld := &mockSceneGeometry{}

	pathLength := dielectric.calculatePathLength(r, hr, scattered, mockWorld)

	// For a sphere with radius ~0.866 (distance from origin to hit point)
	// and ray passing through center, chord length should be ~1.732
	if pathLength < 0.1 || pathLength > 10.0 {
		t.Errorf("Path length %f is outside reasonable bounds", pathLength)
	}
}

// mockSceneGeometry is a simple mock for testing
type mockSceneGeometry struct{}

func (m *mockSceneGeometry) Hit(r ray.Ray, tMin float64, tMax float64) (*hitrecord.HitRecord, Material, bool) {
	// Return a mock exit point for testing
	exitPoint := vec3.Vec3Impl{X: 0.5, Y: 0.5, Z: 1.5}
	normal := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	hr := hitrecord.New(2.0, 0.0, 0.0, exitPoint, normal)
	return hr, nil, true
}

func TestColoredGlassScattering(t *testing.T) {
	absorptionCoeff := vec3.Vec3Impl{X: 0.1, Y: 0.2, Z: 0.3}
	dielectric := NewColoredDielectric(1.5, absorptionCoeff)

	// Set up a mock world for path length calculation
	mockWorld := &mockSceneGeometry{}
	dielectric.SetWorld(mockWorld)

	// Create a hit record
	hitPoint := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	normal := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	hr := hitrecord.New(1.0, 0.0, 0.0, hitPoint, normal)

	// Create a ray
	origin := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: -1.0}
	direction := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	r := ray.New(origin, direction, 0.0)

	// Use a fixed seed for deterministic testing (avoid CI flakiness)
	// We need enough iterations to ensure we hit both reflection and transmission cases
	random := fastrandom.New(12345, 4294967296, 1664525, 1013904223)

	// Test multiple times to account for random reflection/transmission
	foundAttenuation := false
	foundNoAttenuation := false

	for i := 0; i < 1000; i++ {
		scattered, scatterRecord, ok := dielectric.Scatter(r, hr, random)

		if !ok {
			t.Error("Expected scattering to succeed")
		}

		if scattered == nil {
			t.Error("Expected scattered ray to be non-nil")
		}

		if scatterRecord == nil {
			t.Error("Expected scatter record to be non-nil")
		}

		// Check attenuation values
		attenuation := scatterRecord.Attenuation()
		// Use epsilon for floating point comparison
		epsilon := 1e-6
		if attenuation.X < (1.0-epsilon) || attenuation.Y < (1.0-epsilon) || attenuation.Z < (1.0-epsilon) {
			foundAttenuation = true
		}
		if attenuation.X >= (1.0-epsilon) && attenuation.Y >= (1.0-epsilon) && attenuation.Z >= (1.0-epsilon) {
			foundNoAttenuation = true
		}

		// If we found both cases, we can break early
		if foundAttenuation && foundNoAttenuation {
			break
		}
	}

	// Check that we found both attenuated (transmitted) and non-attenuated (reflected) rays
	if !foundAttenuation {
		t.Error("Expected to find some transmitted rays with attenuation < 1.0 for colored glass")
	}
	if !foundNoAttenuation {
		t.Error("Expected to find some reflected rays with attenuation = 1.0")
	}
}

func TestSpectralColoredGlassScattering(t *testing.T) {
	spectralRefIdx := texture.NewSpectralConstant(1.5, 550.0, 50.0)
	spectralAbsorptionCoeff := texture.NewSpectralConstant(0.5, 480.0, 60.0)
	dielectric := NewSpectralColoredDielectric(spectralRefIdx, spectralAbsorptionCoeff)

	// Set up a mock world for path length calculation
	mockWorld := &mockSceneGeometry{}
	dielectric.SetWorld(mockWorld)

	// Create a hit record
	hitPoint := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	normal := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	hr := hitrecord.New(1.0, 0.0, 0.0, hitPoint, normal)

	// Create a ray with wavelength
	origin := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: -1.0}
	direction := vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 1.0}
	r := ray.NewWithLambda(origin, direction, 0.0, 480.0) // Blue wavelength

	// Use a fixed seed for deterministic testing (avoid CI flakiness)
	// We need enough iterations to ensure we hit both reflection and transmission cases
	random := fastrandom.New(12345, 4294967296, 1664525, 1013904223)

	// Test multiple times to account for random reflection/transmission
	foundAttenuation := false
	foundNoAttenuation := false

	for i := 0; i < 1000; i++ {
		scattered, scatterRecord, ok := dielectric.SpectralScatter(r, hr, random)

		if !ok {
			t.Error("Expected spectral scattering to succeed")
		}

		if scattered == nil {
			t.Error("Expected scattered ray to be non-nil")
		}

		if scatterRecord == nil {
			t.Error("Expected scatter record to be non-nil")
		}

		attenuation := scatterRecord.Attenuation()
		if attenuation < 1.0 {
			foundAttenuation = true
		}
		if attenuation >= 1.0 {
			foundNoAttenuation = true
		}

		// If we found both cases, we can break early
		if foundAttenuation && foundNoAttenuation {
			break
		}
	}

	// Check that we found both attenuated (transmitted) and non-attenuated (reflected) rays
	if !foundAttenuation {
		t.Error("Expected to find some transmitted rays with attenuation < 1.0 for colored glass")
	}
	if !foundNoAttenuation {
		t.Error("Expected to find some reflected rays with attenuation = 1.0")
	}
}

// Helper function for floating point comparison
func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}
