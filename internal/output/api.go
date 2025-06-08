// Package output implements the file output functionality.
package output

import "image"

// Output handles the output functionality once the render is complete.
type Output interface {
	// Write persists the render data in a specific format.
	Write(i image.Image) error
}
