package texture

import (
	"fmt"
	"testing"

	"github.com/flynn-nrg/izpi/internal/vec3"
)

func TestSpectralImageDebug(t *testing.T) {
	// Create a simple 1x1 red image
	si := NewSpectralImageFromRawData(1, 1, []float64{1.0, 0.0, 0.0, 1.0})

	// Test different wavelengths for red pixel
	testWavelengths := []float64{380, 450, 550, 650, 750}

	fmt.Println("Testing red pixel (1,0,0) at different wavelengths:")
	for _, lambda := range testWavelengths {
		value := si.Value(0.5, 0.5, lambda, &vec3.Vec3Impl{})
		fmt.Printf("  Wavelength %.0fnm: %.3f\n", lambda, value)
	}

	// Test the rgbToSpectralValue function directly
	fmt.Println("\nTesting rgbToSpectralValue function directly:")
	for _, lambda := range testWavelengths {
		value := si.rgbToSpectralValue(1.0, 0.0, 0.0, lambda)
		fmt.Printf("  Wavelength %.0fnm: %.3f\n", lambda, value)
	}
}

func TestSpectralImageUVMapping(t *testing.T) {
	// Create a 2x2 image with different colors in each corner
	si := NewSpectralImageFromRawData(2, 2, []float64{
		// Red pixel (1, 0, 0) - top-left
		1.0, 0.0, 0.0, 1.0,
		// Green pixel (0, 1, 0) - top-right
		0.0, 1.0, 0.0, 1.0,
		// Blue pixel (0, 0, 1) - bottom-left
		0.0, 0.0, 1.0, 1.0,
		// White pixel (1, 1, 1) - bottom-right
		1.0, 1.0, 1.0, 1.0,
	})

	fmt.Println("Testing UV to pixel mapping:")

	// Test each corner
	testUVs := []struct {
		name string
		u, v float64
	}{
		{"Top-left (red)", 0.0, 1.0},
		{"Top-right (green)", 1.0, 1.0},
		{"Bottom-left (blue)", 0.0, 0.0},
		{"Bottom-right (white)", 1.0, 0.0},
	}

	for _, test := range testUVs {
		// Calculate pixel coordinates
		i := int(test.u * float64(si.sizeX))
		j := int((1 - test.v) * (float64(si.sizeY) - 0.001))

		// Clamp coordinates
		if i < 0 {
			i = 0
		}
		if j < 0 {
			j = 0
		}
		if i > (si.sizeX - 1) {
			i = si.sizeX - 1
		}
		if j > (si.sizeY - 1) {
			j = si.sizeY - 1
		}

		pixelIndex := j*si.sizeX + i

		fmt.Printf("  %s: UV(%.1f,%.1f) -> pixel(%d,%d) -> index %d\n",
			test.name, test.u, test.v, i, j, pixelIndex)

		// Test spectral value at red wavelength
		value := si.Value(test.u, test.v, 650.0, &vec3.Vec3Impl{})
		fmt.Printf("    Spectral value at 650nm: %.3f\n", value)
	}
}

func TestSpectralImageFromRawData(t *testing.T) {
	// Create a simple 2x2 image with red, green, blue, and white pixels
	width, height := 2, 2
	data := []float64{
		// Red pixel (1, 0, 0)
		1.0, 0.0, 0.0, 1.0,
		// Green pixel (0, 1, 0)
		0.0, 1.0, 0.0, 1.0,
		// Blue pixel (0, 0, 1)
		0.0, 0.0, 1.0, 1.0,
		// White pixel (1, 1, 1)
		1.0, 1.0, 1.0, 1.0,
	}

	si := NewSpectralImageFromRawData(width, height, data)

	// Test that we have the correct number of wavelengths
	expectedWavelengths := 75 // 380-750nm in 5nm steps
	if len(si.wavelengths) != expectedWavelengths {
		t.Errorf("Expected %d wavelengths, got %d", expectedWavelengths, len(si.wavelengths))
	}

	// Test that we have the correct number of spectral data buckets
	if len(si.spectralData) != expectedWavelengths {
		t.Errorf("Expected %d spectral data buckets, got %d", expectedWavelengths, len(si.spectralData))
	}

	// Test that each bucket has the correct number of pixels
	expectedPixels := width * height
	for i, bucket := range si.spectralData {
		if len(bucket) != expectedPixels {
			t.Errorf("Bucket %d: expected %d pixels, got %d", i, expectedPixels, len(bucket))
		}
	}

	// Test spectral values at different wavelengths
	testCases := []struct {
		name      string
		u, v      float64
		lambda    float64
		expected  float64
		tolerance float64
	}{
		{
			name:      "Red pixel at red wavelength",
			u:         0.0, // Red pixel (top-left)
			v:         1.0,
			lambda:    650.0, // Red wavelength
			expected:  0.6,   // Should be high for red
			tolerance: 0.2,
		},
		{
			name:      "Green pixel at green wavelength",
			u:         1.0, // Green pixel (top-right)
			v:         1.0,
			lambda:    550.0, // Green wavelength
			expected:  0.8,   // Should be high for green
			tolerance: 0.3,
		},
		{
			name:      "Blue pixel at blue wavelength",
			u:         0.0, // Blue pixel (bottom-left)
			v:         0.0,
			lambda:    450.0, // Blue wavelength
			expected:  0.7,   // Should be high for blue
			tolerance: 0.3,
		},
		{
			name:      "White pixel at any wavelength",
			u:         1.0, // White pixel (bottom-right)
			v:         0.0,
			lambda:    550.0, // Green wavelength
			expected:  0.8,   // Should be high for white
			tolerance: 0.3,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			value := si.Value(tc.u, tc.v, tc.lambda, &vec3.Vec3Impl{})
			if value < tc.expected-tc.tolerance || value > tc.expected+tc.tolerance {
				t.Errorf("Expected value around %f, got %f", tc.expected, value)
			}
		})
	}
}

func TestSpectralImageWavelengthIndex(t *testing.T) {
	// Create a simple 1x1 image
	si := NewSpectralImageFromRawData(1, 1, []float64{1.0, 1.0, 1.0, 1.0})

	testCases := []struct {
		lambda        float64
		expectedIndex int
		expectedValue float64
	}{
		{380.0, 0, 380.0},  // First wavelength
		{385.0, 1, 385.0},  // Second wavelength
		{750.0, 74, 750.0}, // Last wavelength
		{400.0, 4, 400.0},  // Middle wavelength
		{375.0, 0, 380.0},  // Below range, should clamp to first
		{755.0, 74, 750.0}, // Above range, should clamp to last
	}

	for _, tc := range testCases {
		index := si.findWavelengthIndex(tc.lambda)
		if index != tc.expectedIndex {
			t.Errorf("For lambda %f: expected index %d, got %d", tc.lambda, tc.expectedIndex, index)
		}
		if si.wavelengths[index] != tc.expectedValue {
			t.Errorf("For lambda %f: expected wavelength %f, got %f", tc.lambda, tc.expectedValue, si.wavelengths[index])
		}
	}
}
