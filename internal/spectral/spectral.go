package spectral

import (
	"math"
)

const (
	WavelengthMin = 380
	WavelengthMax = 750
	// CIE 1931 color matching functions are defined from 360-830nm
	// but we focus on visible range 380-750nm
)

// CIE 1931 color matching functions (simplified, tabulated data)
// These are the actual CIE x̄(λ), ȳ(λ), z̄(λ) functions
var cieX = []float64{
	0.0014, 0.0022, 0.0042, 0.0076, 0.0143, 0.0232, 0.0435, 0.0776, 0.1344, 0.2148,
	0.2839, 0.3285, 0.3483, 0.3481, 0.3362, 0.3187, 0.2908, 0.2511, 0.1954, 0.1421,
	0.0956, 0.0580, 0.0320, 0.0147, 0.0049, 0.0024, 0.0093, 0.0291, 0.0633, 0.1096,
	0.1655, 0.2257, 0.2904, 0.3597, 0.4334, 0.5121, 0.5945, 0.6784, 0.7621, 0.8425,
	0.9163, 0.9786, 1.0263, 1.0567, 1.0622, 1.0456, 1.0026, 0.9384, 0.8544, 0.7514,
	0.6424, 0.5419, 0.4479, 0.3608, 0.2835, 0.2187, 0.1649, 0.1212, 0.0874, 0.0636,
	0.0468, 0.0329, 0.0227, 0.0158, 0.0114, 0.0081, 0.0058, 0.0041, 0.0029, 0.0021,
	0.0015, 0.0011, 0.0008, 0.0006, 0.0004, 0.0003, 0.0002, 0.0002, 0.0001, 0.0001,
}

var cieY = []float64{
	0.0000, 0.0001, 0.0001, 0.0002, 0.0004, 0.0006, 0.0012, 0.0022, 0.0040, 0.0073,
	0.0116, 0.0168, 0.0230, 0.0298, 0.0380, 0.0480, 0.0600, 0.0739, 0.0910, 0.1126,
	0.1390, 0.1693, 0.2080, 0.2586, 0.3230, 0.4073, 0.5030, 0.6082, 0.7100, 0.7932,
	0.8620, 0.9149, 0.9540, 0.9803, 0.9950, 1.0000, 0.9950, 0.9786, 0.9520, 0.9154,
	0.8700, 0.8163, 0.7570, 0.6949, 0.6310, 0.5668, 0.5030, 0.4412, 0.3810, 0.3210,
	0.2650, 0.2170, 0.1750, 0.1382, 0.1070, 0.0816, 0.0610, 0.0446, 0.0320, 0.0232,
	0.0170, 0.0119, 0.0082, 0.0057, 0.0041, 0.0029, 0.0021, 0.0015, 0.0010, 0.0007,
	0.0005, 0.0004, 0.0003, 0.0002, 0.0001, 0.0001, 0.0001, 0.0000, 0.0000, 0.0000,
}

var cieZ = []float64{
	0.0065, 0.0105, 0.0201, 0.0362, 0.0679, 0.1102, 0.2074, 0.3713, 0.6456, 1.0391,
	1.3856, 1.6230, 1.7471, 1.7826, 1.7721, 1.7441, 1.6692, 1.5281, 1.2876, 1.0419,
	0.8130, 0.6162, 0.4652, 0.3533, 0.2720, 0.2123, 0.1582, 0.1117, 0.0782, 0.0573,
	0.0422, 0.0298, 0.0203, 0.0134, 0.0087, 0.0057, 0.0039, 0.0027, 0.0021, 0.0018,
	0.0017, 0.0014, 0.0011, 0.0010, 0.0009, 0.0008, 0.0006, 0.0003, 0.0002, 0.0000,
	0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
	0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
	0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000, 0.0000,
}

// Wavelengths corresponding to the CIE data (380-750nm in 5nm steps)
var cieWavelengths = []float64{
	380, 385, 390, 395, 400, 405, 410, 415, 420, 425,
	430, 435, 440, 445, 450, 455, 460, 465, 470, 475,
	480, 485, 490, 495, 500, 505, 510, 515, 520, 525,
	530, 535, 540, 545, 550, 555, 560, 565, 570, 575,
	580, 585, 590, 595, 600, 605, 610, 615, 620, 625,
	630, 635, 640, 645, 650, 655, 660, 665, 670, 675,
	680, 685, 690, 695, 700, 705, 710, 715, 720, 725,
	730, 735, 740, 745, 750,
}

// SpectralPowerDistribution represents spectral data
type SpectralPowerDistribution struct {
	wavelengths []float64
	values      []float64
}

// NewSPD creates a new spectral power distribution
func NewSPD(wavelengths, values []float64) *SpectralPowerDistribution {
	return &SpectralPowerDistribution{
		wavelengths: wavelengths,
		values:      values,
	}
}

func NewEmptyCIESPD() *SpectralPowerDistribution {
	return &SpectralPowerDistribution{
		wavelengths: cieWavelengths,
		values:      make([]float64, len(cieWavelengths)),
	}
}

func (spd *SpectralPowerDistribution) SetValue(index int, value float64) {
	spd.values[index] = value
}

func (spd *SpectralPowerDistribution) AddValue(index int, value float64) {
	spd.values[index] += value
}

func (spd *SpectralPowerDistribution) NumWavelengths() int {
	return len(spd.wavelengths)
}

func (spd *SpectralPowerDistribution) Wavelength(index int) float64 {
	return spd.wavelengths[index]
}

func (spd *SpectralPowerDistribution) Normalise(numSamples int) {
	for i, value := range spd.values {
		spd.values[i] = value / float64(numSamples)
	}
}

// NormaliseAndScale normalises the spectral power distribution by dividing by numSamples
// and then applies a scale factor in a single pass for efficiency
func (spd *SpectralPowerDistribution) NormaliseAndScale(numSamples int, scale float64) {
	for i, value := range spd.values {
		spd.values[i] = (value / float64(numSamples)) * scale
	}
}

// Wavelengths returns the wavelengths array
func (spd *SpectralPowerDistribution) Wavelengths() []float64 {
	return spd.wavelengths
}

// Values returns the values array
func (spd *SpectralPowerDistribution) Values() []float64 {
	return spd.values
}

// Value returns the spectral value at the given wavelength, interpolating if necessary
func (spd *SpectralPowerDistribution) Value(wavelength float64) float64 {
	if len(spd.wavelengths) == 0 {
		return 0.0
	}

	// Clamp wavelength to valid range
	if wavelength < spd.wavelengths[0] {
		return spd.values[0]
	}
	if wavelength > spd.wavelengths[len(spd.wavelengths)-1] {
		return spd.values[len(spd.values)-1]
	}

	// Find the two wavelengths that bracket the input wavelength
	for i := 0; i < len(spd.wavelengths)-1; i++ {
		w1 := spd.wavelengths[i]
		w2 := spd.wavelengths[i+1]

		if wavelength >= w1 && wavelength <= w2 {
			// Linear interpolation between the two points
			t := (wavelength - w1) / (w2 - w1)
			v1 := spd.values[i]
			v2 := spd.values[i+1]
			return v1 + t*(v2-v1)
		}
	}

	// If we get here, wavelength is outside the range
	return 0.0
}

// SampleWavelength samples a wavelength according to the CIE Y function (luminance)
// This is useful for importance sampling in spectral rendering
func SampleWavelength(random float64) float64 {
	// Use CIE Y function as importance sampling distribution
	// This samples wavelengths according to human eye sensitivity

	// Find the wavelength that corresponds to the random value
	// by integrating the CIE Y function
	total := 0.0
	for _, y := range cieY {
		total += y
	}

	target := random * total
	current := 0.0

	for i, y := range cieY {
		current += y
		if current >= target {
			// Interpolate between wavelengths
			if i > 0 {
				prev := current - y
				t := (target - prev) / (current - prev)
				return cieWavelengths[i-1] + t*(cieWavelengths[i]-cieWavelengths[i-1])
			}
			return cieWavelengths[i]
		}
	}

	return WavelengthMax
}

// WavelengthToRGB converts a single wavelength to RGB using CIE color matching functions
// This is for visualization/debugging, not for the main spectral rendering loop
func WavelengthToRGB(wavelength float64) (r, g, b float64) {
	// Clamp wavelength to visible range
	if wavelength < WavelengthMin {
		wavelength = WavelengthMin
	}
	if wavelength > WavelengthMax {
		wavelength = WavelengthMax
	}

	// Find the index in the CIE data
	index := 0
	for i, w := range cieWavelengths {
		if w >= wavelength {
			index = i
			break
		}
	}

	// Interpolate CIE values
	var x, y, z float64
	if index == 0 {
		x, y, z = cieX[0], cieY[0], cieZ[0]
	} else if index >= len(cieWavelengths)-1 {
		x, y, z = cieX[len(cieX)-1], cieY[len(cieY)-1], cieZ[len(cieZ)-1]
	} else {
		// Linear interpolation
		w1, w2 := cieWavelengths[index-1], cieWavelengths[index]
		t := (wavelength - w1) / (w2 - w1)
		x = cieX[index-1] + t*(cieX[index]-cieX[index-1])
		y = cieY[index-1] + t*(cieY[index]-cieY[index-1])
		z = cieZ[index-1] + t*(cieZ[index]-cieZ[index-1])
	}

	// Convert XYZ to RGB using sRGB transformation matrix
	// Using a different matrix that might be more accurate for neutral materials
	r = 3.2404542*x - 1.5371385*y - 0.4985314*z
	g = -0.9692660*x + 1.8760108*y + 0.0415560*z
	b = 0.0556434*x - 0.2040259*y + 1.0572252*z

	// Clamp to [0,1] - gamma correction should be applied at the end of the pipeline
	r = math.Max(0, math.Min(1, r))
	g = math.Max(0, math.Min(1, g))
	b = math.Max(0, math.Min(1, b))

	return r, g, b
}

// SPDToRGB converts a spectral power distribution to RGB
// This is what you'd use at the end of spectral rendering
func SPDToRGB(spd *SpectralPowerDistribution) (r, g, b float64) {
	var x, y, z float64
	var numValidWavelengths int

	// Integrate SPD with CIE color matching functions
	for i, wavelength := range spd.wavelengths {
		if wavelength < WavelengthMin || wavelength > WavelengthMax {
			continue
		}

		// Get CIE values for this wavelength
		cieX, cieY, cieZ := GetCIEValues(wavelength)

		// Multiply SPD value by CIE values and accumulate
		value := spd.values[i]
		x += value * cieX
		y += value * cieY
		z += value * cieZ
		numValidWavelengths++
	}

	// Normalize by the number of wavelengths to ensure proper integration
	if numValidWavelengths > 0 {
		x /= float64(numValidWavelengths)
		y /= float64(numValidWavelengths)
		z /= float64(numValidWavelengths)
	}

	// Convert XYZ to RGB using sRGB transformation matrix
	r = 3.2404542*x - 1.5371385*y - 0.4985314*z
	g = -0.9692660*x + 1.8760108*y + 0.0415560*z
	b = 0.0556434*x - 0.2040259*y + 1.0572252*z

	return r, g, b
}

// GetCIEValues returns the CIE color matching function values for a given wavelength
func GetCIEValues(wavelength float64) (x, y, z float64) {
	// Find the index in the CIE data
	index := 0
	for i, w := range cieWavelengths {
		if w >= wavelength {
			index = i
			break
		}
	}

	// Interpolate CIE values
	if index == 0 {
		return cieX[0], cieY[0], cieZ[0]
	} else if index >= len(cieWavelengths)-1 {
		return cieX[len(cieX)-1], cieY[len(cieY)-1], cieZ[len(cieZ)-1]
	} else {
		// Linear interpolation
		w1, w2 := cieWavelengths[index-1], cieWavelengths[index]
		t := (wavelength - w1) / (w2 - w1)
		x = cieX[index-1] + t*(cieX[index]-cieX[index-1])
		y = cieY[index-1] + t*(cieY[index]-cieY[index-1])
		z = cieZ[index-1] + t*(cieZ[index]-cieZ[index-1])
		return x, y, z
	}
}
