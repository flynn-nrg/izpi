package transport

import (
	"testing"

	"github.com/flynn-nrg/izpi/internal/material"
	"github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

func TestPBRMaterialTransformation(t *testing.T) {
	// Create a simple test scene with RGB representation
	protoScene := &transport.Scene{
		ColourRepresentation: transport.ColourRepresentation_RGB,
		Materials: map[string]*transport.Material{
			"test_pbr": {
				Name: "test_pbr",
				Type: transport.MaterialType_PBR,
				MaterialProperties: &transport.Material_Pbr{
					Pbr: &transport.PBRMaterial{
						Albedo: &transport.Texture{
							Name: "test_albedo",
							Type: transport.TextureType_CONSTANT,
							TextureProperties: &transport.Texture_Constant{
								Constant: &transport.ConstantTexture{
									Value: &transport.Vec3{X: 1.0, Y: 0.5, Z: 0.2},
								},
							},
						},
						Roughness: &transport.Texture{
							Name: "test_roughness",
							Type: transport.TextureType_CONSTANT,
							TextureProperties: &transport.Texture_Constant{
								Constant: &transport.ConstantTexture{
									Value: &transport.Vec3{X: 0.5, Y: 0.5, Z: 0.5},
								},
							},
						},
						Metalness: &transport.Texture{
							Name: "test_metalness",
							Type: transport.TextureType_CONSTANT,
							TextureProperties: &transport.Texture_Constant{
								Constant: &transport.ConstantTexture{
									Value: &transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						NormalMap: &transport.Texture{
							Name: "test_normal",
							Type: transport.TextureType_CONSTANT,
							TextureProperties: &transport.Texture_Constant{
								Constant: &transport.ConstantTexture{
									Value: &transport.Vec3{X: 0.5, Y: 0.5, Z: 1.0},
								},
							},
						},
						Sss: &transport.Texture{
							Name: "test_sss",
							Type: transport.TextureType_CONSTANT,
							TextureProperties: &transport.Texture_Constant{
								Constant: &transport.ConstantTexture{
									Value: &transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
		},
	}

	// Create transport instance
	trans := &Transport{
		colourRepresentation: protoScene.GetColourRepresentation(),
		protoScene:           protoScene,
		textures:             make(map[string]*texture.ImageTxt),
	}

	// Test RGB PBR material creation
	mat, err := trans.toScenePBRMaterial(protoScene.GetMaterials()["test_pbr"])
	if err != nil {
		t.Fatalf("Failed to create RGB PBR material: %v", err)
	}

	// Verify it's a regular PBR material (not spectral)
	if _, ok := mat.(*material.PBR); !ok {
		t.Errorf("Expected regular PBR material, got %T", mat)
	}

	// Test spectral PBR material creation
	protoScene.ColourRepresentation = transport.ColourRepresentation_SPECTRAL
	trans.colourRepresentation = transport.ColourRepresentation_SPECTRAL

	spectralMat, err := trans.toScenePBRMaterial(protoScene.GetMaterials()["test_pbr"])
	if err != nil {
		t.Fatalf("Failed to create spectral PBR material: %v", err)
	}

	// Verify it's a PBR material with spectral support
	if _, ok := spectralMat.(*material.PBR); !ok {
		t.Errorf("Expected PBR material, got %T", spectralMat)
	}

	// Test that the spectral material has spectral albedo support
	pbrMat := spectralMat.(*material.PBR)

	// Test spectral albedo values at different wavelengths
	testWavelengths := []float64{380, 550, 650}
	for _, lambda := range testWavelengths {
		value := pbrMat.SpectralAlbedo(0.5, 0.5, lambda, vec3.Vec3Impl{})
		if value < 0.0 || value > 1.0 {
			t.Errorf("Spectral albedo value at %.0fnm is out of range [0,1]: %f", lambda, value)
		}
	}
}

func TestTextureToSpectralTexture(t *testing.T) {
	trans := &Transport{}

	// Test constant texture conversion
	constantTex := texture.NewConstant(vec3.Vec3Impl{X: 1.0, Y: 0.5, Z: 0.2})
	spectralTex, err := trans.textureToSpectralTexture(constantTex)
	if err != nil {
		t.Fatalf("Failed to convert constant texture to spectral: %v", err)
	}

	if spectralTex == nil {
		t.Fatal("Expected non-nil spectral texture")
	}

	// Test that the spectral texture returns reasonable values
	value := spectralTex.Value(0.5, 0.5, 550.0, vec3.Vec3Impl{})
	if value < 0.0 || value > 1.0 {
		t.Errorf("Spectral texture value is out of range [0,1]: %f", value)
	}
}

func TestLightSourceLibraryIntegration(t *testing.T) {
	trans := &Transport{}

	tests := []struct {
		name           string
		lightSourceKey string
		expectSuccess  bool
	}{
		{"Valid LED light source", "hy_cree_llf_tm_30_90", true},
		{"Valid fluorescent light source", "cie_f4_warm_white_fluorescent", true},
		{"Valid incandescent light source", "incandescent_2800k", true},
		{"Invalid light source (fallback to CIE A)", "nonexistent_light_source", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spectralText := &transport.SpectralConstantTexture{
				SpectralProperties: &transport.SpectralConstantTexture_FromLightSourceLibrary{
					FromLightSourceLibrary: &transport.FromLightSourceLibrary{
						LightSourceName: tt.lightSourceKey,
					},
				},
			}

			spectralTex, err := trans.toSceneSpectralTexture(spectralText)
			if !tt.expectSuccess && err != nil {
				// Expected to fail
				return
			}

			if err != nil {
				t.Fatalf("Failed to create spectral texture from light source '%s': %v", tt.lightSourceKey, err)
			}

			if spectralTex == nil {
				t.Fatal("Expected non-nil spectral texture")
			}

			// Test that the spectral texture returns reasonable values
			testWavelengths := []float64{400, 500, 600, 700}
			for _, lambda := range testWavelengths {
				value := spectralTex.Value(0.5, 0.5, lambda, vec3.Vec3Impl{})
				if value < 0.0 || value > 1.0 {
					t.Errorf("Light source '%s': spectral value at %.0fnm is out of range [0,1]: %f",
						tt.lightSourceKey, lambda, value)
				}
			}

			t.Logf("Successfully created spectral texture from light source '%s'", tt.lightSourceKey)
		})
	}
}

func TestSpectralTextureMethods(t *testing.T) {
	trans := &Transport{}

	tests := []struct {
		name          string
		createTexture func() *transport.SpectralConstantTexture
	}{
		{
			"Gaussian spectral texture",
			func() *transport.SpectralConstantTexture {
				return &transport.SpectralConstantTexture{
					SpectralProperties: &transport.SpectralConstantTexture_Gaussian{
						Gaussian: &transport.GaussianSpectralConstant{
							PeakValue:        1.0,
							CenterWavelength: 550,
							Width:            40,
						},
					},
				}
			},
		},
		{
			"Neutral spectral texture",
			func() *transport.SpectralConstantTexture {
				return &transport.SpectralConstantTexture{
					SpectralProperties: &transport.SpectralConstantTexture_Neutral{
						Neutral: &transport.NeutralSpectralConstant{
							Reflectance: 0.73,
						},
					},
				}
			},
		},
		{
			"Tabulated spectral texture",
			func() *transport.SpectralConstantTexture {
				return &transport.SpectralConstantTexture{
					SpectralProperties: &transport.SpectralConstantTexture_Tabulated{
						Tabulated: &transport.TabulatedSpectralConstant{
							Wavelengths: []float32{380, 500, 600, 750},
							Values:      []float32{0.1, 0.5, 0.8, 0.3},
						},
					},
				}
			},
		},
		{
			"Light source from library",
			func() *transport.SpectralConstantTexture {
				return &transport.SpectralConstantTexture{
					SpectralProperties: &transport.SpectralConstantTexture_FromLightSourceLibrary{
						FromLightSourceLibrary: &transport.FromLightSourceLibrary{
							LightSourceName: "cie_f7_broadband_daylight",
						},
					},
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			spectralText := tt.createTexture()
			spectralTex, err := trans.toSceneSpectralTexture(spectralText)
			if err != nil {
				t.Fatalf("Failed to create spectral texture: %v", err)
			}

			if spectralTex == nil {
				t.Fatal("Expected non-nil spectral texture")
			}

			// Test that the spectral texture returns reasonable values
			value := spectralTex.Value(0.5, 0.5, 550.0, vec3.Vec3Impl{})
			if value < 0.0 {
				t.Errorf("Spectral texture value at 550nm is negative: %f", value)
			}

			t.Logf("%s: value at 550nm = %f", tt.name, value)
		})
	}
}
