package texturemanager

import (
	"fmt"
	"unsafe"
)

// Float64Slice returns a zero-copy []float64 view over Data.
func (t *SerialisedTexture) Float64Slice() ([]float64, error) {
	if t.PixelDataType != PixelDataTypeRGBAFloat64 {
		return nil, fmt.Errorf("%w: expected RGBAFloat64, got %v", ErrTypeMismatch, t.PixelDataType)
	}

	if len(t.Data) == 0 {
		return nil, nil
	}

	if len(t.Data)%8 != 0 {
		return nil, fmt.Errorf("%w: %d bytes is not a multiple of 8", ErrInvalidSize, len(t.Data))
	}

	ptr := unsafe.SliceData(t.Data)
	if uintptr(unsafe.Pointer(ptr))%unsafe.Alignof(float64(0)) != 0 {
		return nil, ErrMisaligned
	}

	return unsafe.Slice((*float64)(unsafe.Pointer(ptr)), len(t.Data)/8), nil
}

// Float32Slice returns a zero-copy []float32 view over Data.
func (t *SerialisedTexture) Float32Slice() ([]float32, error) {
	if t.PixelDataType != PixelDataTypeRGBAFloat32 {
		return nil, fmt.Errorf("%w: expected RGBAFloat32, got %v", ErrTypeMismatch, t.PixelDataType)
	}
	if len(t.Data) == 0 {
		return nil, nil
	}
	if len(t.Data)%4 != 0 {
		return nil, fmt.Errorf("%w: %d bytes is not a multiple of 4", ErrInvalidSize, len(t.Data))
	}

	ptr := unsafe.SliceData(t.Data)
	if uintptr(unsafe.Pointer(ptr))%unsafe.Alignof(float32(0)) != 0 {
		return nil, ErrMisaligned
	}

	return unsafe.Slice((*float32)(unsafe.Pointer(ptr)), len(t.Data)/4), nil
}
