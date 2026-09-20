package texturemanager

import "errors"

var (
	ErrTypeMismatch = errors.New("pixel data type mismatch")
	ErrInvalidSize  = errors.New("data buffer size does not match expected element size")
	ErrMisaligned   = errors.New("data buffer memory address is misaligned for type")
)
