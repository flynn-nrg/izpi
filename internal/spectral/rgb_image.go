package spectral

import (
	"image"

	"github.com/flynn-nrg/floatimage/colour"
	"github.com/flynn-nrg/floatimage/floatimage"
)

// XYZToRGB converts a CIE XYZ image to linear RGB, applying exposure and white balance.
// The input image contains XYZ values in the R, G, B channels respectively.
// Returns a new image with linear RGB values (not gamma corrected).
func XYZToRGB(in *floatimage.Float64NRGBA, exposure float64, whiteBalance *WhiteBalanceConfig) *floatimage.Float64NRGBA {
	if in == nil {
		return nil
	}

	// Use default white balance if none provided
	if whiteBalance == nil {
		whiteBalance = NewWhiteBalanceDefault()
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

			// Apply white balance matrix to convert XYZ to RGB
			r, g, b := whiteBalance.Matrix.Apply(xVal, yVal, zVal)

			// Set the RGB values in the output image
			out.Set(x, y, colour.Float64NRGBA{R: r, G: g, B: b, A: alpha})
		}
	}

	return out
}
