package spectral

import (
	"testing"

	"github.com/flynn-nrg/go-vfx/math32"
)

func TestComputeWhitePointFromTemperature(t *testing.T) {
	tests := []struct {
		name        string
		temperature float32
		wantCloseTo WhitePointXYZ // Approximate expected values
		tolerance   float32
	}{
		{
			name:        "D65 (6500K)",
			temperature: 6500,
			wantCloseTo: D65,
			tolerance:   0.05, // Allow 5% deviation
		},
		{
			name:        "Incandescent (2800K)",
			temperature: 2800,
			wantCloseTo: WhitePointXYZ{X: 1.09, Y: 1.0, Z: 0.35}, // Approximate incandescent
			tolerance:   0.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ComputeWhitePointFromTemperature(tt.temperature)

			// Check Y is normalized to 1.0
			if math32.Abs(got.Y-1.0) > 0.001 {
				t.Errorf("ComputeWhitePointFromTemperature() Y = %v, want 1.0", got.Y)
			}

			// Check values are within tolerance
			xDiff := math32.Abs(got.X-tt.wantCloseTo.X) / tt.wantCloseTo.X
			zDiff := math32.Abs(got.Z-tt.wantCloseTo.Z) / math32.Max(tt.wantCloseTo.Z, 0.1) // Avoid division by very small numbers

			if xDiff > tt.tolerance {
				t.Errorf("ComputeWhitePointFromTemperature() X = %v, want approximately %v", got.X, tt.wantCloseTo.X)
			}
			if zDiff > tt.tolerance {
				t.Errorf("ComputeWhitePointFromTemperature() Z = %v, want approximately %v", got.Z, tt.wantCloseTo.Z)
			}
		})
	}
}

func TestComputeWhitePointFromSPD(t *testing.T) {
	// Test with a D65-like SPD
	d65SPD := NewBlackbodySPD(6500)
	whitePoint := ComputeWhitePointFromSPD(d65SPD)

	// Y should be normalized to 1.0
	if math32.Abs(whitePoint.Y-1.0) > 0.001 {
		t.Errorf("ComputeWhitePointFromSPD() Y = %v, want 1.0", whitePoint.Y)
	}

	// Should be close to D65
	if math32.Abs(whitePoint.X-D65.X) > 0.1 {
		t.Errorf("ComputeWhitePointFromSPD() X = %v, want approximately %v", whitePoint.X, D65.X)
	}
}

func TestXYZToRGBMatrixApply(t *testing.T) {
	// Test the standard sRGB matrix with a white point
	matrix := sRGBD65Matrix

	// D65 white point should convert to white in RGB (1, 1, 1) approximately
	r, g, b := matrix.Apply(D65.X, D65.Y, D65.Z)

	// Check that RGB values are close to 1.0
	tolerance := float32(0.05)
	if math32.Abs(r-1.0) > tolerance || math32.Abs(g-1.0) > tolerance || math32.Abs(b-1.0) > tolerance {
		t.Errorf("sRGBD65Matrix.Apply(D65) = (%v, %v, %v), want approximately (1.0, 1.0, 1.0)", r, g, b)
	}
}

func TestComputeAdaptedXYZToRGBMatrix(t *testing.T) {
	// Test that D65 white point returns the standard matrix
	matrix := ComputeAdaptedXYZToRGBMatrix(D65)

	// Compare with sRGB D65 matrix
	for i := range 3 {
		for j := range 3 {
			if math32.Abs(matrix[i][j]-sRGBD65Matrix[i][j]) > 0.0001 {
				t.Errorf("ComputeAdaptedXYZToRGBMatrix(D65)[%d][%d] = %v, want %v", i, j, matrix[i][j], sRGBD65Matrix[i][j])
			}
		}
	}

	// Test with an incandescent white point (2800K)
	incandescentWP := ComputeWhitePointFromTemperature(2800)
	incandescentMatrix := ComputeAdaptedXYZToRGBMatrix(incandescentWP)

	// The matrix should be different from D65
	hasDifference := false
	for i := range 3 {
		for j := range 3 {
			if math32.Abs(incandescentMatrix[i][j]-sRGBD65Matrix[i][j]) > 0.01 {
				hasDifference = true
				break
			}
		}
	}

	if !hasDifference {
		t.Error("ComputeAdaptedXYZToRGBMatrix(incandescent) should differ from D65 matrix")
	}
}

func TestNewWhiteBalanceFromTemperature(t *testing.T) {
	tests := []struct {
		name        string
		temperature float32
		wantErr     bool
	}{
		{
			name:        "Valid D65",
			temperature: 6500,
			wantErr:     false,
		},
		{
			name:        "Valid incandescent",
			temperature: 2800,
			wantErr:     false,
		},
		{
			name:        "Too low",
			temperature: 500,
			wantErr:     true,
		},
		{
			name:        "Too high",
			temperature: 30000,
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config, err := NewWhiteBalanceFromTemperature(tt.temperature)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewWhiteBalanceFromTemperature() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if config == nil {
					t.Error("NewWhiteBalanceFromTemperature() returned nil config without error")
				}
				if config.Description == "" {
					t.Error("NewWhiteBalanceFromTemperature() config has empty description")
				}
			}
		})
	}
}

func TestNewWhiteBalanceFromSPD(t *testing.T) {
	// Test with a valid SPD
	spd := NewBlackbodySPD(5000)
	config, err := NewWhiteBalanceFromSPD(spd, "Test SPD")

	if err != nil {
		t.Errorf("NewWhiteBalanceFromSPD() unexpected error = %v", err)
	}
	if config == nil {
		t.Fatal("NewWhiteBalanceFromSPD() returned nil config")
	}
	if config.Description != "Test SPD" {
		t.Errorf("NewWhiteBalanceFromSPD() description = %v, want 'Test SPD'", config.Description)
	}

	// Test with nil SPD
	_, err = NewWhiteBalanceFromSPD(nil, "Test")
	if err == nil {
		t.Error("NewWhiteBalanceFromSPD(nil) should return error")
	}
}

func TestNewWhiteBalanceDefault(t *testing.T) {
	config := NewWhiteBalanceDefault()

	if config == nil {
		t.Fatal("NewWhiteBalanceDefault() returned nil")
	}

	// Should use D65
	if config.WhitePoint.X != D65.X || config.WhitePoint.Y != D65.Y || config.WhitePoint.Z != D65.Z {
		t.Errorf("NewWhiteBalanceDefault() white point = %+v, want D65 = %+v", config.WhitePoint, D65)
	}

	// Matrix should be the standard sRGB matrix
	for i := range 3 {
		for j := range 3 {
			if math32.Abs(config.Matrix[i][j]-sRGBD65Matrix[i][j]) > 0.0001 {
				t.Errorf("NewWhiteBalanceDefault() matrix[%d][%d] = %v, want %v", i, j, config.Matrix[i][j], sRGBD65Matrix[i][j])
			}
		}
	}
}

func TestChromaticAdaptation(t *testing.T) {
	// Test that chromatic adaptation from D65 to D65 is identity-like
	adaptMatrix := ComputeChromaticAdaptationMatrix(D65, D65)

	// Should be close to identity matrix
	for i := range 3 {
		for j := range 3 {
			expected := float32(0.0)
			if i == j {
				expected = 1.0
			}
			if math32.Abs(adaptMatrix[i][j]-expected) > 0.01 {
				t.Errorf("ComputeChromaticAdaptationMatrix(D65, D65)[%d][%d] = %v, want %v", i, j, adaptMatrix[i][j], expected)
			}
		}
	}

	// Test that adapting from incandescent to D65 produces a different matrix
	incandescentWP := ComputeWhitePointFromTemperature(2800)
	adaptMatrix = ComputeChromaticAdaptationMatrix(incandescentWP, D65)

	// Should not be identity
	isIdentity := true
	for i := range 3 {
		for j := range 3 {
			expected := float32(0.0)
			if i == j {
				expected = 1.0
			}
			if math32.Abs(adaptMatrix[i][j]-expected) > 0.1 {
				isIdentity = false
				break
			}
		}
	}

	if isIdentity {
		t.Error("ComputeChromaticAdaptationMatrix(incandescent, D65) should not be identity")
	}
}
