package scenes

import (
	"testing"

	"github.com/flynn-nrg/izpi/internal/proto/transport"
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
