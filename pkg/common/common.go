// Package common implements utility functions used by several packages.
package common

// Lerp performs a linear interpolation between v0 and v1 at t.
func Lerp(v0 float64, v1 float64, t float64) float64 {
	return (1-t)*v0 + t*v1
}

func Max(a, b int) int {
	if a > b {
		return a
	}

	return b
}

func Min(a, b int) int {
	if a < b {
		return a
	}

	return b
}

func MaxF(a, b float64) float64 {
	if a > b {
		return a
	}

	return b
}

func MinF(a, b float64) float64 {
	if a < b {
		return a
	}

	return b
}

func ClampF(v, low, high float64) float64 {
	if v < low {
		return low
	}

	if v > high {
		return high
	}

	return v
}

func Clamp(v, low, high int) int {
	if v < low {
		return low
	}

	if v > high {
		return high
	}

	return v
}
