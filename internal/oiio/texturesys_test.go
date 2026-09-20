package oiio

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

// writeFlatTestTexture writes a small flat-color PNG so filtered lookups are
// deterministic regardless of the mip level or filter width OIIO picks. The
// alpha must be non-opaque: Go's png encoder silently drops the alpha channel
// (encoding RGB instead of RGBA) whenever the image is fully opaque.
func writeFlatTestTexture(t *testing.T, path string, c color.NRGBA) {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, 8, 8))
	for y := range 8 {
		for x := range 8 {
			img.SetNRGBA(x, y, c)
		}
	}
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to create test texture: %v", err)
	}
	defer f.Close()
	if err := png.Encode(f, img); err != nil {
		t.Fatalf("failed to encode test texture: %v", err)
	}
}

func TestTextureSystemLookup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "flat.png")
	writeFlatTestTexture(t, path, color.NRGBA{R: 255, G: 128, B: 0, A: 254})

	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	opts := NewTextureLookupOptions()
	result, err := ts.Texture(path, opts, 0.5, 0.5, 0, 0, 0, 0, 4)
	if err != nil {
		t.Fatalf("Texture: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("expected 4 channels, got %d", len(result))
	}

	// PNG is 8-bit sRGB; OIIO decodes it to linear-ish float unless told
	// otherwise, so just sanity-check ordering (R > G > B) and full alpha
	// rather than exact values.
	if !(result[0] > result[1] && result[1] > result[2]) {
		t.Fatalf("expected R > G > B, got %v", result)
	}
	if result[3] < 0.9 {
		t.Fatalf("expected alpha close to 1.0, got %v", result[3])
	}
}

func TestTextureSystemGetTextureInfo(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "flat.png")
	writeFlatTestTexture(t, path, color.NRGBA{R: 10, G: 20, B: 30, A: 254})

	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	channels, err := ts.GetTextureInfoInt(path, "channels")
	if err != nil {
		t.Fatalf("GetTextureInfoInt(channels): %v", err)
	}
	if channels != 4 {
		t.Fatalf("expected 4 channels, got %d", channels)
	}

	exists, err := ts.GetTextureInfoInt(path, "exists")
	if err != nil {
		t.Fatalf("GetTextureInfoInt(exists): %v", err)
	}
	if exists != 1 {
		t.Fatalf("expected exists=1, got %d", exists)
	}
}

func TestTextureSystemMissingFile(t *testing.T) {
	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	opts := NewTextureLookupOptions()
	_, err = ts.Texture("/nonexistent/path/does-not-exist.png", opts, 0.5, 0.5, 0, 0, 0, 0, 4)
	if err == nil {
		t.Fatalf("expected an error looking up a missing texture")
	}
}

func TestTextureSystemEnvironmentLookup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "flat.png")
	writeFlatTestTexture(t, path, color.NRGBA{R: 200, G: 100, B: 50, A: 254})

	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	opts := NewTextureLookupOptions()
	r := Vec3{X: 0, Y: 0, Z: 1}
	result, err := ts.Environment(path, opts, r, Vec3{}, Vec3{}, 4)
	if err != nil {
		t.Fatalf("Environment: %v", err)
	}
	if len(result) != 4 {
		t.Fatalf("expected 4 channels, got %d", len(result))
	}
	if !(result[0] > result[1] && result[1] > result[2]) {
		t.Fatalf("expected R > G > B, got %v", result)
	}
}

func TestTextureSystemTexture3DMissingFile(t *testing.T) {
	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	opts := NewTextureLookupOptions()
	p := Vec3{X: 0.5, Y: 0.5, Z: 0.5}
	_, err = ts.Texture3D("/nonexistent/path/does-not-exist.vdb", opts, p, Vec3{}, Vec3{}, Vec3{}, 4)
	if err == nil {
		t.Fatalf("expected an error looking up a missing 3D texture")
	}
}

func TestTextureSystemAttributes(t *testing.T) {
	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	if err := ts.SetAttributeInt("max_open_files", 42); err != nil {
		t.Fatalf("SetAttributeInt: %v", err)
	}
	got, err := ts.GetAttributeInt("max_open_files")
	if err != nil {
		t.Fatalf("GetAttributeInt: %v", err)
	}
	if got != 42 {
		t.Fatalf("expected max_open_files=42, got %d", got)
	}

	if err := ts.SetAttributeFloat("max_memory_MB", 128.5); err != nil {
		t.Fatalf("SetAttributeFloat: %v", err)
	}
	gotF, err := ts.GetAttributeFloat("max_memory_MB")
	if err != nil {
		t.Fatalf("GetAttributeFloat: %v", err)
	}
	if gotF != 128.5 {
		t.Fatalf("expected max_memory_MB=128.5, got %v", gotF)
	}

	dir := t.TempDir()
	if err := ts.SetAttributeString("searchpath", dir); err != nil {
		t.Fatalf("SetAttributeString: %v", err)
	}
	gotS, err := ts.GetAttributeString("searchpath")
	if err != nil {
		t.Fatalf("GetAttributeString: %v", err)
	}
	if gotS != dir {
		t.Fatalf("expected searchpath=%q, got %q", dir, gotS)
	}
}

func TestTextureSystemAttributeUnknown(t *testing.T) {
	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	if err := ts.SetAttributeInt("not_a_real_attribute", 1); err == nil {
		t.Fatalf("expected an error setting an unrecognized attribute")
	}
	if _, err := ts.GetAttributeInt("not_a_real_attribute"); err == nil {
		t.Fatalf("expected an error getting an unrecognized attribute")
	}
}

func TestTextureSystemHandleLookup(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "flat.png")
	writeFlatTestTexture(t, path, color.NRGBA{R: 255, G: 128, B: 0, A: 254})

	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	thread := ts.NewPerThreadInfo()
	defer thread.Close()

	handle, err := ts.GetTextureHandle(path, thread)
	if err != nil {
		t.Fatalf("GetTextureHandle: %v", err)
	}

	opts := NewTextureLookupOptions()
	viaHandle, err := handle.Texture(thread, opts, 0.5, 0.5, 0, 0, 0, 0, 4)
	if err != nil {
		t.Fatalf("handle.Texture: %v", err)
	}
	viaFilename, err := ts.Texture(path, opts, 0.5, 0.5, 0, 0, 0, 0, 4)
	if err != nil {
		t.Fatalf("ts.Texture: %v", err)
	}
	for i := range viaHandle {
		if viaHandle[i] != viaFilename[i] {
			t.Fatalf("handle-based and filename-based lookups disagree: %v vs %v", viaHandle, viaFilename)
		}
	}

	channels, err := handle.GetTextureInfoInt(thread, "channels")
	if err != nil {
		t.Fatalf("handle.GetTextureInfoInt: %v", err)
	}
	if channels != 4 {
		t.Fatalf("expected 4 channels, got %d", channels)
	}
}

// GetTextureHandle resolves lazily: it succeeds even for a nonexistent file,
// and the failure only surfaces on the first actual lookup through the
// handle (see the doc comment on GetTextureHandle).
func TestTextureSystemHandleMissingFile(t *testing.T) {
	ts, err := NewTextureSystem()
	if err != nil {
		t.Fatalf("NewTextureSystem: %v", err)
	}
	defer ts.Close()

	handle, err := ts.GetTextureHandle("/nonexistent/path/does-not-exist.png", nil)
	if err != nil {
		t.Fatalf("GetTextureHandle: %v", err)
	}

	opts := NewTextureLookupOptions()
	if _, err := handle.Texture(nil, opts, 0.5, 0.5, 0, 0, 0, 0, 4); err == nil {
		t.Fatalf("expected an error looking up a missing texture via handle")
	}
}
