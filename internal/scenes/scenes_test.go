package scenes

import (
	"testing"

	"github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/stretchr/testify/assert"
)

func TestCornellBoxPBRRGB(t *testing.T) {
	// Test that the scene can be created
	scene := CornellBoxPBRRGB(1.0)

	// Verify basic scene properties
	if scene.Name != "Cornell Box PBR RGB" {
		t.Errorf("Expected scene name 'Cornell Box PBR RGB', got '%s'", scene.Name)
	}

	if scene.ColourRepresentation != transport.ColourRepresentation_RGB {
		t.Errorf("Expected RGB colour representation, got %v", scene.ColourRepresentation)
	}

	// Verify camera properties
	if scene.Camera == nil {
		t.Fatal("Scene camera is nil")
	}

	if scene.Camera.Vfov != 40 {
		t.Errorf("Expected camera VFOV 40, got %f", scene.Camera.Vfov)
	}

	// Verify materials exist
	if len(scene.Materials) == 0 {
		t.Fatal("Scene has no materials")
	}

	// Check for specific materials
	expectedMaterials := []string{"White", "Green", "Red", "white_light", "Glass", "rusty-metal", "grainy-concrete", "fleshy_granite1", "bamboo-wood-semigloss", "lightgold"}
	for _, materialName := range expectedMaterials {
		if _, exists := scene.Materials[materialName]; !exists {
			t.Errorf("Expected material '%s' not found in scene", materialName)
		}
	}

	// Verify objects exist
	if scene.Objects == nil {
		t.Fatal("Scene objects is nil")
	}

	if len(scene.Objects.Triangles) == 0 {
		t.Fatal("Scene has no triangles")
	}

	if len(scene.Objects.Spheres) == 0 {
		t.Fatal("Scene has no spheres")
	}

	// Verify image textures exist
	if len(scene.ImageTextures) == 0 {
		t.Fatal("Scene has no image textures")
	}

	// Check for specific texture files
	expectedTextures := []string{
		"textures/rusty-metal_albedo.png",
		"textures/grainy-concrete_albedo.png",
		"textures/fleshy_granite1_albedo.png",
		"textures/bamboo-wood-semigloss-albedo.png",
		"textures/lightgold_albedo.png",
	}

	for _, textureName := range expectedTextures {
		if _, exists := scene.ImageTextures[textureName]; !exists {
			t.Errorf("Expected texture '%s' not found in scene", textureName)
		}
	}

	t.Logf("Scene created successfully with %d triangles, %d spheres, %d materials, and %d textures",
		len(scene.Objects.Triangles),
		len(scene.Objects.Spheres),
		len(scene.Materials),
		len(scene.ImageTextures))
}

func TestCornellBoxPBRSpectral(t *testing.T) {
	scene := CornellBoxPBRSpectral(1.0)

	// Test basic scene properties
	assert.Equal(t, "Cornell Box PBR Spectral", scene.Name)
	assert.Equal(t, "1.0.0", scene.Version)
	assert.Equal(t, transport.ColourRepresentation_SPECTRAL, scene.ColourRepresentation)

	// Test camera properties
	assert.NotNil(t, scene.Camera)
	assert.Equal(t, float32(50), scene.Camera.Lookfrom.X)
	assert.Equal(t, float32(50), scene.Camera.Lookfrom.Y)
	assert.Equal(t, float32(-140), scene.Camera.Lookfrom.Z)

	// Test materials
	assert.NotNil(t, scene.Materials)
	assert.Len(t, scene.Materials, 10) // White, Green, Red, white_light, Glass, rusty-metal, grainy-concrete, fleshy_granite1, bamboo-wood-semigloss, lightgold

	// Test spectral materials
	whiteMaterial := scene.Materials["White"]
	assert.NotNil(t, whiteMaterial)
	assert.Equal(t, transport.MaterialType_LAMBERT, whiteMaterial.Type)
	lambert := whiteMaterial.GetLambert()
	assert.NotNil(t, lambert)
	assert.NotNil(t, lambert.GetSpectralAlbedo())
	assert.NotNil(t, lambert.GetSpectralAlbedo().GetNeutral())
	assert.Equal(t, float32(0.73), lambert.GetSpectralAlbedo().GetNeutral().Reflectance)

	greenMaterial := scene.Materials["Green"]
	assert.NotNil(t, greenMaterial)
	greenLambert := greenMaterial.GetLambert()
	assert.NotNil(t, greenLambert.GetSpectralAlbedo())
	assert.NotNil(t, greenLambert.GetSpectralAlbedo().GetGaussian())
	assert.Equal(t, float32(0.9), greenLambert.GetSpectralAlbedo().GetGaussian().PeakValue)
	assert.Equal(t, float32(550.0), greenLambert.GetSpectralAlbedo().GetGaussian().CenterWavelength)

	redMaterial := scene.Materials["Red"]
	assert.NotNil(t, redMaterial)
	redLambert := redMaterial.GetLambert()
	assert.NotNil(t, redLambert.GetSpectralAlbedo())
	assert.NotNil(t, redLambert.GetSpectralAlbedo().GetGaussian())
	assert.Equal(t, float32(0.9), redLambert.GetSpectralAlbedo().GetGaussian().PeakValue)
	assert.Equal(t, float32(650.0), redLambert.GetSpectralAlbedo().GetGaussian().CenterWavelength)

	// Test spectral light
	lightMaterial := scene.Materials["white_light"]
	assert.NotNil(t, lightMaterial)
	assert.Equal(t, transport.MaterialType_DIFFUSE_LIGHT, lightMaterial.Type)
	diffuseLight := lightMaterial.GetDiffuselight()
	assert.NotNil(t, diffuseLight)
	assert.NotNil(t, diffuseLight.GetSpectralEmit())
	assert.NotNil(t, diffuseLight.GetSpectralEmit().GetNeutral())
	assert.Equal(t, float32(15.0), diffuseLight.GetSpectralEmit().GetNeutral().Reflectance)

	// Test spectral glass material
	glassMaterial := scene.Materials["Glass"]
	assert.NotNil(t, glassMaterial)
	assert.Equal(t, transport.MaterialType_DIELECTRIC, glassMaterial.Type)
	dielectric := glassMaterial.GetDielectric()
	assert.NotNil(t, dielectric)
	assert.NotNil(t, dielectric.GetSpectralRefidx())
	assert.NotNil(t, dielectric.GetSpectralRefidx().GetTabulated())
	assert.Len(t, dielectric.GetSpectralRefidx().GetTabulated().Wavelengths, 20)
	assert.Len(t, dielectric.GetSpectralRefidx().GetTabulated().Values, 20)
	assert.Equal(t, float32(380), dielectric.GetSpectralRefidx().GetTabulated().Wavelengths[0])
	assert.Equal(t, float32(750), dielectric.GetSpectralRefidx().GetTabulated().Wavelengths[19])
	assert.Equal(t, float32(1.52), dielectric.GetSpectralRefidx().GetTabulated().Values[0])
	assert.Equal(t, float32(1.42), dielectric.GetSpectralRefidx().GetTabulated().Values[19])

	// Test PBR materials (should remain the same as RGB version)
	rustyMetal := scene.Materials["rusty-metal"]
	assert.NotNil(t, rustyMetal)
	assert.Equal(t, transport.MaterialType_PBR, rustyMetal.Type)
	pbr := rustyMetal.GetPbr()
	assert.NotNil(t, pbr)
	assert.NotNil(t, pbr.Albedo)
	assert.Equal(t, "textures/rusty-metal_albedo.png", pbr.Albedo.GetImage().Filename)

	// Test objects
	assert.NotNil(t, scene.Objects)
	assert.Len(t, scene.Objects.Triangles, 13) // Same as RGB version
	assert.Len(t, scene.Objects.Spheres, 6)    // Same as RGB version

	// Test image textures
	assert.NotNil(t, scene.ImageTextures)
	assert.Len(t, scene.ImageTextures, 19) // Same texture count as RGB version
}
