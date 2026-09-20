package oiio

/*
#cgo CXXFLAGS: -std=c++17
#cgo pkg-config: OpenImageIO fmt
#include <stdlib.h>
#include "./texturesys_wrapper.h"
*/
import "C"

import (
	"fmt"
	"runtime"
	"unsafe"
)

// WrapMode selects how texture lookups behave outside the [0,1] st range.
type WrapMode C.TextureWrapMode

const (
	WrapDefault  WrapMode = C.WRAP_DEFAULT
	WrapBlack    WrapMode = C.WRAP_BLACK
	WrapClamp    WrapMode = C.WRAP_CLAMP
	WrapPeriodic WrapMode = C.WRAP_PERIODIC
	WrapMirror   WrapMode = C.WRAP_MIRROR
)

// TextureLookupOptions mirrors OIIO's TextureOpt fields relevant to a single
// filtered lookup. The R fields only apply to Texture3D (volumetric) lookups.
// Use NewTextureLookupOptions to get OIIO's defaults (notably the width
// fields at 1) rather than a zero-valued struct.
type TextureLookupOptions struct {
	FirstChannel           int
	Subimage               int
	SWrap, TWrap, RWrap    WrapMode
	SBlur, TBlur, RBlur    float32
	SWidth, TWidth, RWidth float32
	Fill                   float32
}

// NewTextureLookupOptions returns TextureLookupOptions set to OIIO's own
// defaults. Starting from a zero-valued struct instead would set the width
// fields to 0, disabling derivative-based filtering rather than using OIIO's
// default multiplier of 1.
func NewTextureLookupOptions() TextureLookupOptions {
	return TextureLookupOptions{SWidth: 1, TWidth: 1, RWidth: 1}
}

func (o TextureLookupOptions) toC() C.TextureLookupOptions {
	return C.TextureLookupOptions{
		firstchannel: C.int(o.FirstChannel),
		subimage:     C.int(o.Subimage),
		swrap:        C.TextureWrapMode(o.SWrap),
		twrap:        C.TextureWrapMode(o.TWrap),
		rwrap:        C.TextureWrapMode(o.RWrap),
		sblur:        C.float(o.SBlur),
		tblur:        C.float(o.TBlur),
		rblur:        C.float(o.RBlur),
		swidth:       C.float(o.SWidth),
		twidth:       C.float(o.TWidth),
		rwidth:       C.float(o.RWidth),
		fill:         C.float(o.Fill),
	}
}

// Vec3 is a 3-component vector used for Texture3D/Environment lookup
// coordinates and their screen-space derivatives.
type Vec3 struct {
	X, Y, Z float32
}

func (v Vec3) toC() [3]C.float {
	return [3]C.float{C.float(v.X), C.float(v.Y), C.float(v.Z)}
}

// TextureSystem wraps an OIIO::TextureSystem: a filtered, mipmapped,
// tile-cached texture lookup engine, as opposed to the whole-image I/O in
// ReadImage/WriteImage. Callers must call Close when done; a finalizer is
// registered as a safety net but should not be relied upon since the
// underlying tile cache can be large.
type TextureSystem struct {
	ptr *C.TextureSystemHandle
}

// NewTextureSystem creates a TextureSystem backed by OIIO's shared tile
// cache. It is safe to create more than one; OIIO shares the underlying
// cache across instances created with shared=true.
func NewTextureSystem() (*TextureSystem, error) {
	var errorMsg *C.char
	ptr := C.texturesystem_create(&errorMsg)
	if ptr == nil {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("failed to create texture system: %s", err)
		}
		return nil, fmt.Errorf("failed to create texture system")
	}

	ts := &TextureSystem{ptr: ptr}
	runtime.SetFinalizer(ts, (*TextureSystem).Close)
	return ts, nil
}

// Close releases the underlying OIIO::TextureSystem and its tile cache. It
// is safe to call more than once.
func (ts *TextureSystem) Close() {
	if ts.ptr != nil {
		C.texturesystem_destroy(ts.ptr)
		ts.ptr = nil
		runtime.SetFinalizer(ts, nil)
	}
}

// Texture performs a filtered lookup at (s,t), using the screen-space
// derivatives dsdx/dtdx (d/dx) and dsdy/dtdy (d/dy) to pick a mip level and
// filter width. This mirrors OSL's texture() shadeop, which always supplies
// derivatives computed from the shading point's surface differentials.
// Returns nchannels float32 values.
func (ts *TextureSystem) Texture(filename string, opts TextureLookupOptions, s, t, dsdx, dtdx, dsdy, dtdy float32, nchannels int) ([]float32, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	cOpts := opts.toC()
	result := make([]float32, nchannels)

	var errorMsg *C.char
	ret := C.texturesystem_texture(ts.ptr, cFilename, &cOpts, C.float(s), C.float(t),
		C.float(dsdx), C.float(dtdx), C.float(dsdy), C.float(dtdy),
		C.int(nchannels), (*C.float)(unsafe.Pointer(&result[0])), &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("texture lookup failed: %s", err)
		}
		return nil, fmt.Errorf("texture lookup failed")
	}

	return result, nil
}

// Texture3D performs a filtered volumetric lookup at point p, using the
// screen-space derivatives dpdx/dpdy/dpdz to pick a filter width. This
// mirrors OSL's texture3d() shadeop. Returns nchannels float32 values.
func (ts *TextureSystem) Texture3D(filename string, opts TextureLookupOptions, p, dpdx, dpdy, dpdz Vec3, nchannels int) ([]float32, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	cOpts := opts.toC()
	cP, cDPdx, cDPdy, cDPdz := p.toC(), dpdx.toC(), dpdy.toC(), dpdz.toC()
	result := make([]float32, nchannels)

	var errorMsg *C.char
	ret := C.texturesystem_texture3d(ts.ptr, cFilename, &cOpts,
		&cP[0], &cDPdx[0], &cDPdy[0], &cDPdz[0],
		C.int(nchannels), (*C.float)(unsafe.Pointer(&result[0])), &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("texture3d lookup failed: %s", err)
		}
		return nil, fmt.Errorf("texture3d lookup failed")
	}

	return result, nil
}

// Environment performs a filtered lookup along direction r (a lat-long or
// cube environment map), using the screen-space derivatives drdx/drdy to pick
// a filter width. This mirrors OSL's environment() shadeop. Returns
// nchannels float32 values.
func (ts *TextureSystem) Environment(filename string, opts TextureLookupOptions, r, drdx, drdy Vec3, nchannels int) ([]float32, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	cOpts := opts.toC()
	cR, cDRdx, cDRdy := r.toC(), drdx.toC(), drdy.toC()
	result := make([]float32, nchannels)

	var errorMsg *C.char
	ret := C.texturesystem_environment(ts.ptr, cFilename, &cOpts,
		&cR[0], &cDRdx[0], &cDRdy[0],
		C.int(nchannels), (*C.float)(unsafe.Pointer(&result[0])), &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("environment lookup failed: %s", err)
		}
		return nil, fmt.Errorf("environment lookup failed")
	}

	return result, nil
}

// GetTextureInfoInt retrieves an integer-valued texture attribute, e.g.
// "channels" or "exists", mirroring OSL's gettextureinfo() shadeop.
func (ts *TextureSystem) GetTextureInfoInt(filename, dataname string) (int, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	cDataname := C.CString(dataname)
	defer C.free(unsafe.Pointer(cDataname))

	var out C.int
	var errorMsg *C.char
	ret := C.texturesystem_get_texture_info_int(ts.ptr, cFilename, cDataname, &out, &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return 0, fmt.Errorf("get_texture_info failed: %s", err)
		}
		return 0, fmt.Errorf("get_texture_info failed")
	}

	return int(out), nil
}

// GetTextureInfoFloat retrieves a float-valued texture attribute, mirroring
// OSL's gettextureinfo() shadeop.
func (ts *TextureSystem) GetTextureInfoFloat(filename, dataname string) (float32, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))
	cDataname := C.CString(dataname)
	defer C.free(unsafe.Pointer(cDataname))

	var out C.float
	var errorMsg *C.char
	ret := C.texturesystem_get_texture_info_float(ts.ptr, cFilename, cDataname, &out, &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return 0, fmt.Errorf("get_texture_info failed: %s", err)
		}
		return 0, fmt.Errorf("get_texture_info failed")
	}

	return float32(out), nil
}

// SetAttributeInt sets a TextureSystem-wide integer configuration option,
// e.g. "max_open_files", "autotile", or "automip".
func (ts *TextureSystem) SetAttributeInt(name string, value int) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var errorMsg *C.char
	if C.texturesystem_attribute_int(ts.ptr, cName, C.int(value), &errorMsg) != 0 {
		return attributeError("set", name, errorMsg)
	}
	return nil
}

// SetAttributeFloat sets a TextureSystem-wide float configuration option,
// e.g. "max_memory_MB".
func (ts *TextureSystem) SetAttributeFloat(name string, value float32) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var errorMsg *C.char
	if C.texturesystem_attribute_float(ts.ptr, cName, C.float(value), &errorMsg) != 0 {
		return attributeError("set", name, errorMsg)
	}
	return nil
}

// SetAttributeString sets a TextureSystem-wide string configuration option,
// e.g. "searchpath" (a colon/semicolon-separated list of directories to
// search for texture files referenced by relative path).
func (ts *TextureSystem) SetAttributeString(name, value string) error {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))
	cValue := C.CString(value)
	defer C.free(unsafe.Pointer(cValue))

	var errorMsg *C.char
	if C.texturesystem_attribute_string(ts.ptr, cName, cValue, &errorMsg) != 0 {
		return attributeError("set", name, errorMsg)
	}
	return nil
}

// GetAttributeInt retrieves a TextureSystem-wide integer configuration
// option previously set via SetAttributeInt (or an OIIO built-in default).
func (ts *TextureSystem) GetAttributeInt(name string) (int, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var out C.int
	var errorMsg *C.char
	if C.texturesystem_getattribute_int(ts.ptr, cName, &out, &errorMsg) != 0 {
		return 0, attributeError("get", name, errorMsg)
	}
	return int(out), nil
}

// GetAttributeFloat retrieves a TextureSystem-wide float configuration
// option previously set via SetAttributeFloat (or an OIIO built-in default).
func (ts *TextureSystem) GetAttributeFloat(name string) (float32, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var out C.float
	var errorMsg *C.char
	if C.texturesystem_getattribute_float(ts.ptr, cName, &out, &errorMsg) != 0 {
		return 0, attributeError("get", name, errorMsg)
	}
	return float32(out), nil
}

// GetAttributeString retrieves a TextureSystem-wide string configuration
// option previously set via SetAttributeString (or an OIIO built-in default).
func (ts *TextureSystem) GetAttributeString(name string) (string, error) {
	cName := C.CString(name)
	defer C.free(unsafe.Pointer(cName))

	var out *C.char
	var errorMsg *C.char
	if C.texturesystem_getattribute_string(ts.ptr, cName, &out, &errorMsg) != 0 {
		return "", attributeError("get", name, errorMsg)
	}
	defer C.free(unsafe.Pointer(out))
	return C.GoString(out), nil
}

func attributeError(op, name string, errorMsg *C.char) error {
	if errorMsg != nil {
		err := C.GoString(errorMsg)
		C.free(unsafe.Pointer(errorMsg))
		return fmt.Errorf("failed to %s attribute %q: %s", op, name, err)
	}
	return fmt.Errorf("failed to %s attribute %q", op, name)
}

// PerThreadInfo holds OIIO thread-local scratch state that speeds up
// repeated texture lookups by avoiding lock contention in the tile cache.
//
// Go's goroutines are not OS threads and can migrate between them, so
// relying on OIIO's own thread-local Perthread (as passing thread_info=nil
// throughout would do) is not meaningful here. Instead, create one
// PerThreadInfo per long-lived rendering worker (e.g. one per worker
// goroutine that owns a fixed slice of work for its lifetime), reuse it for
// every lookup that worker makes, and Close it when the worker exits. Never
// share a PerThreadInfo between concurrently running goroutines.
type PerThreadInfo struct {
	ts  *TextureSystem
	ptr *C.PerThreadInfo
}

// NewPerThreadInfo creates per-thread scratch state for repeated texture
// lookups. See PerThreadInfo for the concurrency contract.
func (ts *TextureSystem) NewPerThreadInfo() *PerThreadInfo {
	ptr := C.texturesystem_create_thread_info(ts.ptr)
	info := &PerThreadInfo{ts: ts, ptr: ptr}
	runtime.SetFinalizer(info, (*PerThreadInfo).Close)
	return info
}

// Close releases the per-thread scratch state. Safe to call more than once.
func (info *PerThreadInfo) Close() {
	if info.ptr != nil {
		C.texturesystem_destroy_thread_info(info.ts.ptr, info.ptr)
		info.ptr = nil
		runtime.SetFinalizer(info, nil)
	}
}

func threadInfoPtr(info *PerThreadInfo) *C.PerThreadInfo {
	if info == nil {
		return nil
	}
	return info.ptr
}

// TextureHandle is a pre-resolved reference to a texture file, obtained via
// TextureSystem.GetTextureHandle. Reusing a handle (together with a
// PerThreadInfo) across many lookups skips the per-call filename hash/lookup
// that Texture/Texture3D/Environment/GetTextureInfo* otherwise repeat every
// time. Useful in a renderer's inner shading loop where the same texture is
// looked up repeatedly. The handle is owned by the TextureSystem: it remains
// valid for the TextureSystem's lifetime and must not be closed.
type TextureHandle struct {
	ts  *TextureSystem
	ptr *C.TextureHandle
}

// GetTextureHandle resolves filename once for reuse across many lookups.
// thread may be nil, at a throughput cost (see PerThreadInfo). Resolution is
// lazy: this rarely fails, even for a nonexistent file.
// A missing/unopenable file only surfaces as an error from the first actual
// lookup (Texture, Texture3D, Environment, or GetTextureInfo*) through the
// returned handle.
func (ts *TextureSystem) GetTextureHandle(filename string, thread *PerThreadInfo) (*TextureHandle, error) {
	cFilename := C.CString(filename)
	defer C.free(unsafe.Pointer(cFilename))

	var errorMsg *C.char
	ptr := C.texturesystem_get_texture_handle(ts.ptr, cFilename, threadInfoPtr(thread), &errorMsg)
	if ptr == nil {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("failed to resolve texture handle for %q: %s", filename, err)
		}
		return nil, fmt.Errorf("failed to resolve texture handle for %q", filename)
	}
	return &TextureHandle{ts: ts, ptr: ptr}, nil
}

// Texture is the TextureHandle fast-path equivalent of TextureSystem.Texture.
func (h *TextureHandle) Texture(thread *PerThreadInfo, opts TextureLookupOptions, s, t, dsdx, dtdx, dsdy, dtdy float32, nchannels int) ([]float32, error) {
	cOpts := opts.toC()
	result := make([]float32, nchannels)

	var errorMsg *C.char
	ret := C.texturesystem_texture_handle(h.ts.ptr, h.ptr, threadInfoPtr(thread), &cOpts,
		C.float(s), C.float(t), C.float(dsdx), C.float(dtdx), C.float(dsdy), C.float(dtdy),
		C.int(nchannels), (*C.float)(unsafe.Pointer(&result[0])), &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("texture lookup failed: %s", err)
		}
		return nil, fmt.Errorf("texture lookup failed")
	}

	return result, nil
}

// Texture3D is the TextureHandle fast-path equivalent of TextureSystem.Texture3D.
func (h *TextureHandle) Texture3D(thread *PerThreadInfo, opts TextureLookupOptions, p, dpdx, dpdy, dpdz Vec3, nchannels int) ([]float32, error) {
	cOpts := opts.toC()
	cP, cDPdx, cDPdy, cDPdz := p.toC(), dpdx.toC(), dpdy.toC(), dpdz.toC()
	result := make([]float32, nchannels)

	var errorMsg *C.char
	ret := C.texturesystem_texture3d_handle(h.ts.ptr, h.ptr, threadInfoPtr(thread), &cOpts,
		&cP[0], &cDPdx[0], &cDPdy[0], &cDPdz[0],
		C.int(nchannels), (*C.float)(unsafe.Pointer(&result[0])), &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("texture3d lookup failed: %s", err)
		}
		return nil, fmt.Errorf("texture3d lookup failed")
	}

	return result, nil
}

// Environment is the TextureHandle fast-path equivalent of TextureSystem.Environment.
func (h *TextureHandle) Environment(thread *PerThreadInfo, opts TextureLookupOptions, r, drdx, drdy Vec3, nchannels int) ([]float32, error) {
	cOpts := opts.toC()
	cR, cDRdx, cDRdy := r.toC(), drdx.toC(), drdy.toC()
	result := make([]float32, nchannels)

	var errorMsg *C.char
	ret := C.texturesystem_environment_handle(h.ts.ptr, h.ptr, threadInfoPtr(thread), &cOpts,
		&cR[0], &cDRdx[0], &cDRdy[0],
		C.int(nchannels), (*C.float)(unsafe.Pointer(&result[0])), &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return nil, fmt.Errorf("environment lookup failed: %s", err)
		}
		return nil, fmt.Errorf("environment lookup failed")
	}

	return result, nil
}

// GetTextureInfoInt is the TextureHandle fast-path equivalent of
// TextureSystem.GetTextureInfoInt.
func (h *TextureHandle) GetTextureInfoInt(thread *PerThreadInfo, dataname string) (int, error) {
	cDataname := C.CString(dataname)
	defer C.free(unsafe.Pointer(cDataname))

	var out C.int
	var errorMsg *C.char
	ret := C.texturesystem_get_texture_info_handle_int(h.ts.ptr, h.ptr, threadInfoPtr(thread), cDataname, &out, &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return 0, fmt.Errorf("get_texture_info failed: %s", err)
		}
		return 0, fmt.Errorf("get_texture_info failed")
	}

	return int(out), nil
}

// GetTextureInfoFloat is the TextureHandle fast-path equivalent of
// TextureSystem.GetTextureInfoFloat.
func (h *TextureHandle) GetTextureInfoFloat(thread *PerThreadInfo, dataname string) (float32, error) {
	cDataname := C.CString(dataname)
	defer C.free(unsafe.Pointer(cDataname))

	var out C.float
	var errorMsg *C.char
	ret := C.texturesystem_get_texture_info_handle_float(h.ts.ptr, h.ptr, threadInfoPtr(thread), cDataname, &out, &errorMsg)
	if ret != 0 {
		if errorMsg != nil {
			err := C.GoString(errorMsg)
			C.free(unsafe.Pointer(errorMsg))
			return 0, fmt.Errorf("get_texture_info failed: %s", err)
		}
		return 0, fmt.Errorf("get_texture_info failed")
	}

	return float32(out), nil
}
