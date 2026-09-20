package oiio

/*
#cgo CXXFLAGS: -std=c++17
#cgo pkg-config: OpenImageIO fmt
#include <stdlib.h>
#include "./oiio_wrapper.h"
*/
import "C"

import (
	"errors"
	"fmt"
	"image"
	"math"
	"path/filepath"
	"strings"
	"unsafe"

	"github.com/flynn-nrg/izpi/internal/floatimage"
)

type ReadImageOptions C.ReadImageOptions

const (
	Raw             ReadImageOptions = C.RAW
	LineariseSRGB   ReadImageOptions = C.LINEARISE_SRGB
	ConvertToACEScg ReadImageOptions = C.CONVERT_TO_ACESCG
)

func ReadImage32(filename string) (*floatimage.Float32NRGBA, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var errorMsg *C.char
	cImage := C.read_image(cFilename, &errorMsg, C.RAW)
	if cImage == nil {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("failed to read image: %s", err)
		}
		return nil, fmt.Errorf("failed to read image")
	}
	defer C.free_image(cImage)

	width := int(cImage.width)
	height := int(cImage.height)
	numChannels := int(cImage.channels)
	cData := (*[1 << 30]C.float)(unsafe.Pointer(cImage.data))[: width*height*numChannels : width*height*numChannels]

	var data []float32

	switch cImage.channels {
	case 3:
		data = toRGB32Slice(cData, width, height)
	case 4:
		data = toRGBA32Slice(cData, width, height, numChannels)
	default:
		return nil, fmt.Errorf("unsupported number of channels: %d", cImage.channels)
	}

	return floatimage.NewFloat32NRGBA(image.Rect(0, 0, width, height), data), nil
}

// ReadImage reads an image and converts it to the specified color space using OpenColorIO.
func ReadImage(filename string, options ReadImageOptions) (*floatimage.Float32NRGBA, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var errorMsg *C.char
	cImage := C.read_image(cFilename, &errorMsg, C.ReadImageOptions(options))
	if cImage == nil {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("failed to read image: %s", err)
		}
		return nil, fmt.Errorf("failed to read image")
	}
	defer C.free_image(cImage)

	width := int(cImage.width)
	height := int(cImage.height)
	numChannels := int(cImage.channels)
	cData := (*[1 << 30]C.float)(unsafe.Pointer(cImage.data))[: width*height*numChannels : width*height*numChannels]

	var data []float32

	switch cImage.channels {
	case 3:
		data = toRGB32Slice(cData, width, height)
	case 4:
		data = toRGBA32Slice(cData, width, height, numChannels)
	default:
		return nil, fmt.Errorf("unsupported number of channels: %d", cImage.channels)
	}

	return floatimage.NewFloat32NRGBA(image.Rect(0, 0, width, height), data), nil
}

func ReadImage64(filename string) (*floatimage.Float64NRGBA, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var errorMsg *C.char
	cImage := C.read_image(cFilename, &errorMsg, C.RAW)
	if cImage == nil {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("failed to read image: %s", err)
		}
		return nil, fmt.Errorf("failed to read image")
	}
	defer C.free_image(cImage)

	width := int(cImage.width)
	height := int(cImage.height)
	numChannels := int(cImage.channels)
	cData := (*[1 << 30]C.float)(unsafe.Pointer(cImage.data))[: width*height*numChannels : width*height*numChannels]

	var data []float64

	switch cImage.channels {
	case 3:
		data = toRGB64Slice(cData, width, height)
	case 4:
		data = toRGBA64Slice(cData, width, height, numChannels)
	default:
		return nil, fmt.Errorf("unsupported number of channels: %d", cImage.channels)
	}

	return floatimage.NewFloat64NRGBA(image.Rect(0, 0, width, height), data), nil
}

// isHDR returns true if the file extension indicates an HDR format
func isHDR(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	switch ext {
	case ".hdr", ".exr", ".pfm", ".dpx":
		return true
	default:
		return false
	}
}

func WriteImage(filename string, image image.Image) error {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var cError *C.char
	defer func() {
		if cError != nil {
			C.free(unsafe.Pointer(cError))
		}
	}()

	isHDRFormat := isHDR(filename)
	cImage := toCImage(image, isHDRFormat)
	defer C.free_image(cImage)

	var cHdr C.int
	if isHDRFormat {
		cHdr = 1
	}

	if ret := C.write_image(cFilename, cImage, &cError, cHdr); ret != 0 {
		if cError != nil {
			return errors.New(C.GoString(cError))
		}
		return errors.New("failed to write image")
	}

	return nil
}

// ACESMetadata contains metadata for ACES/EXR output
type ACESMetadata struct {
	// DisplayWindow defines the full canvas (displayX, displayY, displayWidth, displayHeight)
	DisplayWindow image.Rectangle

	// DataWindow defines the actual pixel data region (can be subset of display window)
	DataWindow image.Rectangle

	// PixelAspectRatio defines the pixel aspect ratio (1.0 for square pixels, 2.0 for 2x anamorphic, etc.)
	// Common values: 1.0 (square), 2.0 (2x anamorphic), 1.5 (1.5x anamorphic)
	PixelAspectRatio float32

	// Timecode is optional SMPTE timecode (e.g., "01:23:45:12")
	Timecode string

	// ACESVersion specifies the ACES version (e.g., "ACES 1.3")
	ACESVersion string
}

// WriteImageACES writes an image in ACEScg color space with full ACES metadata.
// The image data should already be in ACEScg color space (e.g., from ReadImage32Aces).
// Output is always in EXR format with embedded metadata including:
// - oiio:ColorSpace = "ACEScg"
// - Display window and data window
// - Optional timecode
// - ACES version
func WriteImageACES(filename string, img image.Image, metadata *ACESMetadata) error {
	// Ensure .exr extension
	if !strings.HasSuffix(strings.ToLower(filename), ".exr") {
		filename += ".exr"
	}

	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var cError *C.char
	defer func() {
		if cError != nil {
			C.free(unsafe.Pointer(cError))
		}
	}()

	// Convert image to C format (always HDR for ACES)
	cImage := toCImage(img, true)
	defer C.free_image(cImage)

	// Prepare metadata
	var cMetadata C.ACESMetadata

	// Set display window
	cMetadata.display_x = C.int(metadata.DisplayWindow.Min.X)
	cMetadata.display_y = C.int(metadata.DisplayWindow.Min.Y)
	cMetadata.display_width = C.int(metadata.DisplayWindow.Dx())
	cMetadata.display_height = C.int(metadata.DisplayWindow.Dy())

	// Set data window
	cMetadata.data_x = C.int(metadata.DataWindow.Min.X)
	cMetadata.data_y = C.int(metadata.DataWindow.Min.Y)
	cMetadata.data_width = C.int(metadata.DataWindow.Dx())
	cMetadata.data_height = C.int(metadata.DataWindow.Dy())

	// Set pixel aspect ratio (default to 1.0 if not set)
	pixelAspectRatio := metadata.PixelAspectRatio
	if pixelAspectRatio == 0 {
		pixelAspectRatio = 1.0
	}
	cMetadata.pixel_aspect_ratio = C.float(pixelAspectRatio)

	// Set optional timecode
	var cTimecode *C.char
	if metadata.Timecode != "" {
		cTimecode = C.CString(metadata.Timecode)
		defer C.free(unsafe.Pointer(cTimecode))
	}
	cMetadata.timecode = cTimecode

	// Set ACES version
	var cAcesVersion *C.char
	if metadata.ACESVersion != "" {
		cAcesVersion = C.CString(metadata.ACESVersion)
		defer C.free(unsafe.Pointer(cAcesVersion))
	}
	cMetadata.aces_version = cAcesVersion

	// Write image with ACES metadata
	if ret := C.write_image_aces(cFilename, cImage, &cMetadata, &cError); ret != 0 {
		if cError != nil {
			return errors.New(C.GoString(cError))
		}
		return errors.New("failed to write ACES image")
	}

	return nil
}

func toCImage(image image.Image, isHDRFormat bool) *C.Image {
	bounds := image.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	numChannels := 4

	// Allocate Image struct on the heap
	cImage := (*C.Image)(C.malloc(C.size_t(unsafe.Sizeof(C.Image{}))))
	cImage.width = C.int(width)
	cImage.height = C.int(height)
	cImage.channels = C.int(numChannels)
	cImage.data = (*C.float)(C.malloc(C.size_t(width) * C.size_t(height) * C.size_t(numChannels) * C.size_t(unsafe.Sizeof(float32(0)))))

	// Convert to a slice for easier access
	data := (*[1 << 30]C.float)(unsafe.Pointer(cImage.data))[: width*height*numChannels : width*height*numChannels]

	switch image := image.(type) {
	case *floatimage.Float32NRGBA:
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := image.Float32NRGBAAt(x, y)
				idx := ((y-bounds.Min.Y)*width + (x - bounds.Min.X)) * numChannels
				if isHDRFormat {
					// For HDR formats, pass the values directly
					data[idx] = C.float(c.R)
					data[idx+1] = C.float(c.G)
					data[idx+2] = C.float(c.B)
					data[idx+3] = C.float(c.A)
				} else {
					// For LDR formats, apply gamma correction and scale to [0,1]
					gamma := 2.2
					data[idx] = C.float(math.Pow(math.Min(1.0, math.Max(0.0, float64(c.R))), 1.0/gamma))
					data[idx+1] = C.float(math.Pow(math.Min(1.0, math.Max(0.0, float64(c.G))), 1.0/gamma))
					data[idx+2] = C.float(math.Pow(math.Min(1.0, math.Max(0.0, float64(c.B))), 1.0/gamma))
					data[idx+3] = C.float(math.Min(1.0, math.Max(0.0, float64(c.A))))
				}
			}
		}
	case *floatimage.Float64NRGBA:
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				c := image.Float64NRGBAAt(x, y)
				idx := ((y-bounds.Min.Y)*width + (x - bounds.Min.X)) * numChannels
				if isHDRFormat {
					// For HDR formats, pass the values directly
					data[idx] = C.float(c.R)
					data[idx+1] = C.float(c.G)
					data[idx+2] = C.float(c.B)
					data[idx+3] = C.float(c.A)
				} else {
					// For LDR formats, apply gamma correction and scale to [0,1]
					gamma := 2.2
					data[idx] = C.float(math.Pow(math.Min(1.0, math.Max(0.0, c.R)), 1.0/gamma))
					data[idx+1] = C.float(math.Pow(math.Min(1.0, math.Max(0.0, c.G)), 1.0/gamma))
					data[idx+2] = C.float(math.Pow(math.Min(1.0, math.Max(0.0, c.B)), 1.0/gamma))
					data[idx+3] = C.float(math.Min(1.0, math.Max(0.0, c.A)))
				}
			}
		}
	default:
		for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
			for x := bounds.Min.X; x < bounds.Max.X; x++ {
				r, g, b, a := image.At(x, y).RGBA()
				idx := ((y-bounds.Min.Y)*width + (x - bounds.Min.X)) * numChannels
				// Convert from 16-bit RGBA to float in range [0,1]
				data[idx] = C.float(float32(r) / 65535.0)
				data[idx+1] = C.float(float32(g) / 65535.0)
				data[idx+2] = C.float(float32(b) / 65535.0)
				data[idx+3] = C.float(float32(a) / 65535.0)
			}
		}
	}

	return cImage
}

func toRGBA32Slice(cData []C.float, width int, height int, numChannels int) []float32 {

	data := make([]float32, width*height*numChannels)
	for i := 0; i < len(cData); i++ {
		data[i] = float32(cData[i])
	}
	return data
}

func toRGB32Slice(cData []C.float, width int, height int) []float32 {
	data := make([]float32, width*height*4)

	j := 0

	for i := 0; i < len(data); i += 4 {
		data[i] = float32(cData[j])
		j++
		data[i+1] = float32(cData[j])
		j++
		data[i+2] = float32(cData[j])
		j++
		data[i+3] = 1.0 // alpha channel is always 1.0
	}

	return data
}

func toRGBA64Slice(cData []C.float, width int, height int, numChannels int) []float64 {

	data := make([]float64, width*height*numChannels)
	for i := 0; i < len(cData); i++ {
		data[i] = float64(cData[i])
	}
	return data
}

func toRGB64Slice(cData []C.float, width int, height int) []float64 {
	data := make([]float64, width*height*4)

	j := 0

	for i := 0; i < len(data); i += 4 {
		data[i] = float64(cData[j])
		j++
		data[i+1] = float64(cData[j])
		j++
		data[i+2] = float64(cData[j])
		j++
		data[i+3] = 1.0 // alpha channel is always 1.0
	}

	return data
}
