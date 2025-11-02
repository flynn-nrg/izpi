package spectral

import (
	"math"

	"github.com/flynn-nrg/floatimage/floatimage"
)

// FireflyRejection performs in-place firefly rejection on CIE XYZ image data.
// It uses a neighborhood-based outlier detection approach with mean + k×stddev threshold.
// Pixels with Y (luminance) values that exceed the threshold are clamped while preserving chromaticity.
func FireflyRejection(in *floatimage.Float64NRGBA) {
	if in == nil {
		return
	}

	bounds := in.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width == 0 || height == 0 {
		return
	}

	const (
		kernelRadius = 1   // 3×3 kernel
		kThreshold   = 2.5 // Threshold multiplier
		minNeighbors = 3   // Minimum neighbors needed for statistics
	)

	// Create a copy of Y values to avoid reading modified values during processing
	yValues := make([]float64, width*height)
	for y := range height {
		for x := range width {
			idx := (y*width + x) * 4
			yValues[y*width+x] = in.Pix[idx+1] // Y component
		}
	}

	// Process each pixel
	for y := range height {
		for x := range width {
			pixelIdx := (y*width + x) * 4

			// Get current pixel Y value
			currentY := yValues[y*width+x]

			// Skip if Y is zero or negative (no light)
			if currentY <= 0 {
				continue
			}

			// Collect neighborhood Y values (excluding center pixel)
			var neighborYs []float64

			for dy := -kernelRadius; dy <= kernelRadius; dy++ {
				for dx := -kernelRadius; dx <= kernelRadius; dx++ {
					// Skip center pixel
					if dx == 0 && dy == 0 {
						continue
					}

					nx := x + dx
					ny := y + dy

					// Check bounds
					if nx >= 0 && nx < width && ny >= 0 && ny < height {
						neighborY := yValues[ny*width+nx]
						// Only include positive values
						if neighborY > 0 {
							neighborYs = append(neighborYs, neighborY)
						}
					}
				}
			}

			// Need enough neighbors for meaningful statistics
			if len(neighborYs) < minNeighbors {
				continue
			}

			// Calculate mean
			sum := 0.0
			for _, val := range neighborYs {
				sum += val
			}
			mean := sum / float64(len(neighborYs))

			// Calculate standard deviation
			varianceSum := 0.0
			for _, val := range neighborYs {
				diff := val - mean
				varianceSum += diff * diff
			}
			stddev := math.Sqrt(varianceSum / float64(len(neighborYs)))

			// Calculate threshold
			threshold := mean + kThreshold*stddev

			// If current Y exceeds threshold, clamp it
			if currentY > threshold && threshold > 0 {
				// Calculate scaling ratio to bring Y down to threshold
				ratio := threshold / currentY

				// Scale the entire XYZ triplet to preserve chromaticity
				in.Pix[pixelIdx] *= ratio   // X
				in.Pix[pixelIdx+1] *= ratio // Y
				in.Pix[pixelIdx+2] *= ratio // Z
				// Alpha (pixelIdx+3) remains unchanged
			}
		}
	}
}
