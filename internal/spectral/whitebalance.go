package spectral

import (
	"fmt"

	"github.com/flynn-nrg/go-vfx/math32"
)

// WhitePointXYZ represents a white point in CIE XYZ color space
type WhitePointXYZ struct {
	X, Y, Z float32
}

// D65 is the standard daylight illuminant white point
var D65 = WhitePointXYZ{X: 0.95047, Y: 1.00000, Z: 1.08883}

// ComputeWhitePointFromSPD computes the XYZ white point from a spectral power distribution
func ComputeWhitePointFromSPD(spd *SpectralPowerDistribution) WhitePointXYZ {
	var sumX, sumY, sumZ float32

	// Integrate the SPD against the CIE color matching functions
	wavelengths := spd.Wavelengths()
	values := spd.Values()

	for i, wavelength := range wavelengths {
		if i >= len(values) {
			break
		}

		// Get CIE color matching function values at this wavelength
		x, y, z := GetCIEValues(wavelength)

		// Multiply by the SPD value and accumulate
		spdValue := values[i]
		sumX += spdValue * x
		sumY += spdValue * y
		sumZ += spdValue * z
	}

	// Normalize so that Y = 1.0 (standard for white points)
	if sumY > 0 {
		sumX /= sumY
		sumZ /= sumY
		sumY = 1.0
	}

	return WhitePointXYZ{X: sumX, Y: sumY, Z: sumZ}
}

// ComputeWhitePointFromTemperature computes the XYZ white point from a color temperature
func ComputeWhitePointFromTemperature(temperature float32) WhitePointXYZ {
	// Create a blackbody SPD at the given temperature
	spd := NewBlackbodySPD(temperature)
	return ComputeWhitePointFromSPD(spd)
}

// XYZToRGBMatrix represents a 3x3 transformation matrix for XYZ to RGB conversion
type XYZToRGBMatrix [3][3]float32

// Apply applies the matrix transformation to XYZ values, returning RGB
func (m *XYZToRGBMatrix) Apply(x, y, z float32) (r, g, b float32) {
	r = m[0][0]*x + m[0][1]*y + m[0][2]*z
	g = m[1][0]*x + m[1][1]*y + m[1][2]*z
	b = m[2][0]*x + m[2][1]*y + m[2][2]*z
	return r, g, b
}

// sRGB D65 matrix (for reference - what we use when no adaptation is needed)
var sRGBD65Matrix = XYZToRGBMatrix{
	{3.2404542, -1.5371385, -0.4985314},
	{-0.9692660, 1.8760108, 0.0415560},
	{0.0556434, -0.2040259, 1.0572252},
}

// Bradford chromatic adaptation matrix (from XYZ to cone response space)
var bradfordMatrix = [3][3]float32{
	{0.8951000, 0.2664000, -0.1614000},
	{-0.7502000, 1.7135000, 0.0367000},
	{0.0389000, -0.0685000, 1.0296000},
}

// Bradford inverse matrix (from cone response space back to XYZ)
var bradfordMatrixInv = [3][3]float32{
	{0.9869929, -0.1470543, 0.1599627},
	{0.4323053, 0.5183603, 0.0492912},
	{-0.0085287, 0.0400428, 0.9684867},
}

// multiplyMatrix3x3 multiplies two 3x3 matrices
func multiplyMatrix3x3(a, b [3][3]float32) [3][3]float32 {
	var result [3][3]float32
	for i := range 3 {
		for j := range 3 {
			result[i][j] = 0
			for k := range 3 {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	return result
}

// ComputeChromaticAdaptationMatrix computes a Bradford chromatic adaptation matrix
// that adapts from sourceWhite to targetWhite
func ComputeChromaticAdaptationMatrix(sourceWhite, targetWhite WhitePointXYZ) [3][3]float32 {
	// Convert source white point to cone response space
	srcRho := bradfordMatrix[0][0]*sourceWhite.X + bradfordMatrix[0][1]*sourceWhite.Y + bradfordMatrix[0][2]*sourceWhite.Z
	srcGamma := bradfordMatrix[1][0]*sourceWhite.X + bradfordMatrix[1][1]*sourceWhite.Y + bradfordMatrix[1][2]*sourceWhite.Z
	srcBeta := bradfordMatrix[2][0]*sourceWhite.X + bradfordMatrix[2][1]*sourceWhite.Y + bradfordMatrix[2][2]*sourceWhite.Z

	// Convert target white point to cone response space
	dstRho := bradfordMatrix[0][0]*targetWhite.X + bradfordMatrix[0][1]*targetWhite.Y + bradfordMatrix[0][2]*targetWhite.Z
	dstGamma := bradfordMatrix[1][0]*targetWhite.X + bradfordMatrix[1][1]*targetWhite.Y + bradfordMatrix[1][2]*targetWhite.Z
	dstBeta := bradfordMatrix[2][0]*targetWhite.X + bradfordMatrix[2][1]*targetWhite.Y + bradfordMatrix[2][2]*targetWhite.Z

	// Compute scaling factors
	var scaleRho, scaleGamma, scaleBeta float32
	if srcRho != 0 {
		scaleRho = dstRho / srcRho
	} else {
		scaleRho = 1.0
	}
	if srcGamma != 0 {
		scaleGamma = dstGamma / srcGamma
	} else {
		scaleGamma = 1.0
	}
	if srcBeta != 0 {
		scaleBeta = dstBeta / srcBeta
	} else {
		scaleBeta = 1.0
	}

	// Create diagonal scaling matrix
	scaleMatrix := [3][3]float32{
		{scaleRho, 0, 0},
		{0, scaleGamma, 0},
		{0, 0, scaleBeta},
	}

	// Compute the chromatic adaptation matrix:
	// M_adapt = M_bradford_inv * S * M_bradford
	temp := multiplyMatrix3x3(scaleMatrix, bradfordMatrix)
	adaptMatrix := multiplyMatrix3x3(bradfordMatrixInv, temp)

	return adaptMatrix
}

// ComputeAdaptedXYZToRGBMatrix computes an XYZ-to-RGB conversion matrix adapted
// for the given white point. This combines chromatic adaptation from the source
// white point to D65 (sRGB's white point) with the standard XYZ-to-RGB matrix.
func ComputeAdaptedXYZToRGBMatrix(whitePoint WhitePointXYZ) XYZToRGBMatrix {
	// If the white point is very close to D65, use the standard matrix
	const epsilon = 0.0001
	if math32.Abs(whitePoint.X-D65.X) < epsilon &&
		math32.Abs(whitePoint.Y-D65.Y) < epsilon &&
		math32.Abs(whitePoint.Z-D65.Z) < epsilon {
		return sRGBD65Matrix
	}

	// Compute chromatic adaptation from source white point to D65
	adaptMatrix := ComputeChromaticAdaptationMatrix(whitePoint, D65)

	// Combine adaptation matrix with sRGB D65 matrix
	// result = sRGB * adaptation
	var result XYZToRGBMatrix
	for i := range 3 {
		for j := range 3 {
			result[i][j] = 0
			for k := range 3 {
				result[i][j] += sRGBD65Matrix[i][k] * adaptMatrix[k][j]
			}
		}
	}

	return result
}

// WhiteBalanceConfig holds the configuration for white balance
type WhiteBalanceConfig struct {
	Matrix      XYZToRGBMatrix
	WhitePoint  WhitePointXYZ
	Description string
}

// NewWhiteBalanceFromTemperature creates a white balance configuration from a color temperature
func NewWhiteBalanceFromTemperature(temperature float32) (*WhiteBalanceConfig, error) {
	if temperature < 1000 || temperature > 25000 {
		return nil, fmt.Errorf("temperature %v K is out of valid range (1000-25000 K)", temperature)
	}

	whitePoint := ComputeWhitePointFromTemperature(temperature)
	matrix := ComputeAdaptedXYZToRGBMatrix(whitePoint)

	return &WhiteBalanceConfig{
		Matrix:      matrix,
		WhitePoint:  whitePoint,
		Description: fmt.Sprintf("%.0fK blackbody", temperature),
	}, nil
}

// NewWhiteBalanceFromSPD creates a white balance configuration from a spectral power distribution
func NewWhiteBalanceFromSPD(spd *SpectralPowerDistribution, description string) (*WhiteBalanceConfig, error) {
	if spd == nil {
		return nil, fmt.Errorf("spectral power distribution is nil")
	}

	whitePoint := ComputeWhitePointFromSPD(spd)
	matrix := ComputeAdaptedXYZToRGBMatrix(whitePoint)

	return &WhiteBalanceConfig{
		Matrix:      matrix,
		WhitePoint:  whitePoint,
		Description: description,
	}, nil
}

// NewWhiteBalanceDefault creates a default white balance configuration using D65
func NewWhiteBalanceDefault() *WhiteBalanceConfig {
	return &WhiteBalanceConfig{
		Matrix:      sRGBD65Matrix,
		WhitePoint:  D65,
		Description: "D65 (default)",
	}
}
