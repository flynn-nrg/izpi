package spectral

import (
	"image"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/floatimage/floatimage"
)

// XYZ to ACEScg (AP1) transformation matrix with D60 white point.
// This matrix converts CIE XYZ values to linear ACEScg color space.
// Source: ACES specifications (Academy Color Encoding System)
var xyzToACEScgMatrix = XYZToRGBMatrix{
	{1.6410234, -0.3248033, -0.2364247},
	{-0.6636629, 1.6153316, 0.0167563},
	{0.0117219, -0.0082845, 0.9883949},
}

// XYZToACEScg converts individual XYZ values to ACEScg RGB.
// This is a convenience function for single color conversions.
func XYZToACEScg(x, y, z float64) (r, g, b float64) {
	return xyzToACEScgMatrix.Apply(x, y, z)
}

// XYZToRGB converts a CIE XYZ image to linear ACEScg RGB.
// The input image contains XYZ values in the R, G, B channels respectively.
// Returns a new image with linear ACEScg RGB values (not gamma corrected).
func XYZToRGB(in *floatimage.Float64NRGBA, exposure float64) *floatimage.Float64NRGBA {
	if in == nil {
		return nil
	}

	bounds := in.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Allocate pixel data (4 float64 values per pixel: R, G, B, A)
	pixelData := make([]float64, width*height*4)
	rect := image.Rect(0, 0, width, height)
	out := floatimage.NewFloat64NRGBA(rect, pixelData)

	// Process each pixel
	for y := range height {
		for x := range width {
			// Get XYZ values from the input image
			// Note: R, G, B channels contain X, Y, Z respectively
			xyzColor := in.At(x, y).(colour.Float64NRGBA)
			xVal := xyzColor.R
			yVal := xyzColor.G
			zVal := xyzColor.B
			alpha := xyzColor.A

			// Apply exposure
			xVal *= exposure
			yVal *= exposure
			zVal *= exposure

			// Apply ACEScg matrix to convert XYZ to ACEScg RGB
			r, g, b := xyzToACEScgMatrix.Apply(xVal, yVal, zVal)

			// Set the ACEScg RGB values in the output image
			out.Set(x, y, colour.Float64NRGBA{R: r, G: g, B: b, A: alpha})
		}
	}

	return out
}
