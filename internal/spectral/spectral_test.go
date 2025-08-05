package spectral

import (
	"math"
	"testing"
)

func TestNeutralMaterialConversion(t *testing.T) {
	// Test that neutral materials (constant reflectance) produce equal RGB values
	testCases := []float64{0.1, 0.5, 0.73, 0.9}

	for _, reflectance := range testCases {
		t.Run("reflectance_"+string(rune(int(reflectance*100))), func(t *testing.T) {
			spd := NewEmptyCIESPD()

			// Set all wavelengths to the same reflectance value
			for i := range spd.values {
				spd.SetValue(i, reflectance)
			}

			r, g, b := SPDToRGB(spd)

			// For neutral materials, RGB values should be approximately equal
			tolerance := 0.01
			if math.Abs(r-g) > tolerance || math.Abs(g-b) > tolerance || math.Abs(r-b) > tolerance {
				t.Errorf("Neutral material (reflectance=%.2f) produced unequal RGB values: R=%.3f, G=%.3f, B=%.3f",
					reflectance, r, g, b)
				t.Errorf("Differences: R-G=%.3f, G-B=%.3f, R-B=%.3f",
					math.Abs(r-g), math.Abs(g-b), math.Abs(r-b))
			}

			// RGB values should be approximately equal to the reflectance
			if math.Abs(r-reflectance) > tolerance {
				t.Errorf("RGB value (%.3f) differs significantly from reflectance (%.2f)", r, reflectance)
			}
		})
	}
}

func TestWavelengthToRGB(t *testing.T) {
	// Test that WavelengthToRGB produces reasonable values for known wavelengths
	testCases := []struct {
		wavelength float64
		name       string
		expectedR  float64
		expectedG  float64
		expectedB  float64
		tolerance  float64
	}{
		{450, "Blue", 0.0, 0.0, 1.0, 0.3},   // Blue wavelength should be mostly blue
		{550, "Green", 0.0, 1.0, 0.0, 0.3},  // Green wavelength should be mostly green
		{650, "Red", 1.0, 0.0, 0.0, 0.3},    // Red wavelength should be mostly red
		{580, "Yellow", 1.0, 1.0, 0.0, 0.4}, // Yellow wavelength should be red+green
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			r, g, b := WavelengthToRGB(tc.wavelength)

			// Check that values are in [0,1] range
			if r < 0 || r > 1 || g < 0 || g > 1 || b < 0 || b > 1 {
				t.Errorf("RGB values out of range [0,1]: R=%.3f, G=%.3f, B=%.3f", r, g, b)
			}

			// Check that the dominant color is as expected
			if math.Abs(r-tc.expectedR) > tc.tolerance ||
				math.Abs(g-tc.expectedG) > tc.tolerance ||
				math.Abs(b-tc.expectedB) > tc.tolerance {
				t.Errorf("Wavelength %.0fnm: expected RGB≈(%.1f,%.1f,%.1f), got (%.3f,%.3f,%.3f)",
					tc.wavelength, tc.expectedR, tc.expectedG, tc.expectedB, r, g, b)
			}
		})
	}
}

func TestGetCIEValues(t *testing.T) {
	// Test that GetCIEValues returns reasonable values
	testCases := []struct {
		wavelength float64
		expectedY  float64
		tolerance  float64
	}{
		{380, 0.0, 0.01}, // Start of visible spectrum
		{555, 1.0, 0.1},  // Peak of CIE Y function
		{750, 0.0, 0.01}, // End of visible spectrum
	}

	for _, tc := range testCases {
		t.Run("wavelength_"+string(rune(int(tc.wavelength))), func(t *testing.T) {
			x, y, z := GetCIEValues(tc.wavelength)

			// Check that values are non-negative
			if x < 0 || y < 0 || z < 0 {
				t.Errorf("Negative CIE values: X=%.3f, Y=%.3f, Z=%.3f", x, y, z)
			}

			// Check that Y value is as expected
			if math.Abs(y-tc.expectedY) > tc.tolerance {
				t.Errorf("Y value for wavelength %.0fnm: expected %.1f, got %.3f",
					tc.wavelength, tc.expectedY, y)
			}
		})
	}
}

func TestSpectralPowerDistribution(t *testing.T) {
	// Test SPD basic functionality
	spd := NewEmptyCIESPD()

	// Test that we have the expected number of wavelengths
	expectedWavelengths := 75 // 380-750nm in 5nm steps
	if spd.NumWavelengths() != expectedWavelengths {
		t.Errorf("Expected %d wavelengths, got %d", expectedWavelengths, spd.NumWavelengths())
	}

	// Test wavelength bounds
	if spd.Wavelength(0) != 380 {
		t.Errorf("First wavelength should be 380nm, got %.0fnm", spd.Wavelength(0))
	}
	if spd.Wavelength(spd.NumWavelengths()-1) != 750 {
		t.Errorf("Last wavelength should be 750nm, got %.0fnm", spd.Wavelength(spd.NumWavelengths()-1))
	}

	// Test value setting and getting
	testValue := 0.5
	spd.SetValue(10, testValue)
	if spd.Values()[10] != testValue {
		t.Errorf("Set value %.1f at index 10, but got %.1f", testValue, spd.Values()[10])
	}

	// Test normalization
	spd.SetValue(0, 2.0)
	spd.SetValue(1, 4.0)
	spd.Normalise(2)
	if spd.Values()[0] != 1.0 || spd.Values()[1] != 2.0 {
		t.Errorf("Normalization failed: expected (1.0, 2.0), got (%.1f, %.1f)",
			spd.Values()[0], spd.Values()[1])
	}
}
