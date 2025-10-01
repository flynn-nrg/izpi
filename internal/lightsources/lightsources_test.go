package lightsources

import (
	"sort"
	"strings"
	"testing"
)

func TestLightSourcesPopulated(t *testing.T) {
	sources := ListLightSources()
	if len(sources) == 0 {
		t.Fatal("No light sources found")
	}

	t.Logf("Found %d light sources", len(sources))

	// Test that we can retrieve each one
	for _, key := range sources {
		spd, ok := GetLightSource(key)
		if !ok {
			t.Errorf("Failed to get light source: %s", key)
		}
		if spd == nil {
			t.Errorf("Light source %s is nil", key)
		}
		if spd.NumWavelengths() != 75 {
			t.Errorf("Light source %s has %d wavelengths, expected 75", key, spd.NumWavelengths())
		}
	}
}

func TestListLightSources(t *testing.T) {
	sources := ListLightSources()
	sort.Strings(sources)

	t.Log("Available light sources:")
	for _, key := range sources {
		t.Logf("  - %s", key)
	}
}

func TestIncandescentSources(t *testing.T) {
	tests := []struct {
		name string
		key  string
	}{
		{"Standard Incandescent", "incandescent_2800k"},
		{"Halogen", "halogen_3200k"},
		{"CIE Illuminant A", "cie_illuminant_a_2856k"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spd, ok := GetLightSource(tt.key)
			if !ok {
				t.Errorf("Failed to get %s light source", tt.name)
			}
			if spd == nil {
				t.Errorf("%s light source is nil", tt.name)
			}
			if spd.NumWavelengths() != 75 {
				t.Errorf("%s has %d wavelengths, expected 75", tt.name, spd.NumWavelengths())
			}

			// Check that values are reasonable (normalized, should have max near 1.0)
			maxVal := 0.0
			minVal := 1.0
			for i := 0; i < spd.NumWavelengths(); i++ {
				val := spd.Value(spd.Wavelength(i))
				if val > maxVal {
					maxVal = val
				}
				if val < minVal {
					minVal = val
				}
			}

			t.Logf("%s: max=%.4f, min=%.4f", tt.name, maxVal, minVal)

			if maxVal < 0.9 || maxVal > 1.1 {
				t.Errorf("%s: expected max near 1.0, got %.4f", tt.name, maxVal)
			}
			if minVal < 0.0 {
				t.Errorf("%s: expected all values >= 0, got min %.4f", tt.name, minVal)
			}
		})
	}
}

func TestFluorescentSources(t *testing.T) {
	tests := []struct {
		name string
		key  string
		cct  string
		cri  string
	}{
		{"F1 Daylight", "cie_f1_daylight_fluorescent", "6430K", "76"},
		{"F2 Cool White", "cie_f2_cool_white_fluorescent", "4230K", "64"},
		{"F3 White", "cie_f3_white_fluorescent", "3450K", "57"},
		{"F4 Warm White", "cie_f4_warm_white_fluorescent", "2940K", "51"},
		{"F5 Daylight", "cie_f5_daylight_fluorescent", "6350K", "72"},
		{"F6 Lite White", "cie_f6_lite_white_fluorescent", "4150K", "59"},
		{"F7 Broadband Daylight", "cie_f7_broadband_daylight", "6500K", "90"},
		{"F8 Broadband Cool White", "cie_f8_broadband_cool_white", "5000K", "95"},
		{"F9 Broadband Cool White Deluxe", "cie_f9_broadband_cool_white_deluxe", "4150K", "90"},
		{"F10 Narrowband", "cie_f10_narrowband_5000k", "5000K", "81"},
		{"F11 Narrowband", "cie_f11_narrowband_4000k", "4000K", "83"},
		{"F12 Narrowband", "cie_f12_narrowband_3000k", "3000K", "83"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spd, ok := GetLightSource(tt.key)
			if !ok {
				t.Errorf("Failed to get %s light source", tt.name)
			}
			if spd == nil {
				t.Errorf("%s light source is nil", tt.name)
			}
			if spd.NumWavelengths() != 75 {
				t.Errorf("%s has %d wavelengths, expected 75", tt.name, spd.NumWavelengths())
			}

			// Check that values are reasonable (normalized, should have max near 1.0)
			maxVal := 0.0
			minVal := 1.0
			for i := 0; i < spd.NumWavelengths(); i++ {
				val := spd.Value(spd.Wavelength(i))
				if val > maxVal {
					maxVal = val
				}
				if val < minVal {
					minVal = val
				}
			}

			t.Logf("%s (%s, CRI %s): max=%.4f, min=%.4f", tt.name, tt.cct, tt.cri, maxVal, minVal)

			if maxVal < 0.9 || maxVal > 1.1 {
				t.Errorf("%s: expected max near 1.0, got %.4f", tt.name, maxVal)
			}
			if minVal < 0.0 {
				t.Errorf("%s: expected all values >= 0, got min %.4f", tt.name, minVal)
			}
		})
	}
}

func TestLightSourceCategories(t *testing.T) {
	sources := ListLightSources()

	categories := map[string]int{
		"LED":          0,
		"Incandescent": 0,
		"Fluorescent":  0,
	}

	for _, key := range sources {
		switch {
		case strings.HasPrefix(key, "cie_f"):
			categories["Fluorescent"]++
		case strings.HasPrefix(key, "incandescent_") || strings.HasPrefix(key, "halogen_") || strings.HasPrefix(key, "cie_illuminant_a"):
			categories["Incandescent"]++
		default:
			categories["LED"]++
		}
	}

	t.Logf("Light source categories:")
	t.Logf("  LEDs: %d", categories["LED"])
	t.Logf("  Incandescent: %d", categories["Incandescent"])
	t.Logf("  Fluorescent (CIE F-series): %d", categories["Fluorescent"])
	t.Logf("  Total: %d", len(sources))

	if categories["LED"] == 0 {
		t.Error("Expected at least one LED source")
	}
	if categories["Incandescent"] != 3 {
		t.Errorf("Expected 3 incandescent sources, got %d", categories["Incandescent"])
	}
	if categories["Fluorescent"] != 12 {
		t.Errorf("Expected 12 fluorescent sources, got %d", categories["Fluorescent"])
	}
}
