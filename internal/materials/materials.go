// Package materials implements a library of built-in spectral materials.
package materials

import (
	"github.com/flynn-nrg/izpi/internal/material"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/spectral"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/flynn-nrg/izpi/internal/vec3"
)

// MaterialDefinition represents a material configuration with spectral properties
type MaterialDefinition struct {
	Name           string
	Description    string
	CreateMaterial func() material.Material
}

// Porcelain spectral reflectance data
// Based on typical porcelain/ceramic spectral characteristics:
// - High reflectance across visible spectrum (0.75-0.90)
// - Slight warm tone from increased red reflectance
// - Minimal absorption in blue-violet region
// - Semi-glossy finish with some subsurface scattering
var porcelainSpectralReflectance = []float64{
	// 380-500nm (blue-violet to blue-green) - slightly lower reflectance
	0.78, 0.79, 0.80, 0.81, 0.82, 0.82, 0.83, 0.83, 0.84, 0.84,
	0.85, 0.85, 0.86, 0.86, 0.87, 0.87, 0.88, 0.88, 0.88, 0.88,
	0.88, 0.89, 0.89, 0.89,
	// 500-600nm (green to yellow-orange) - high reflectance
	0.89, 0.89, 0.90, 0.90, 0.90, 0.90, 0.90, 0.90, 0.91, 0.91,
	0.91, 0.91, 0.91, 0.91, 0.91, 0.92, 0.92, 0.92, 0.92, 0.92,
	// 600-750nm (orange to red) - slightly higher for warm tone
	0.92, 0.92, 0.92, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93,
	0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93,
	0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93, 0.93,
	0.93,
}

// Porcelain wavelengths (380-750nm in 5nm steps, matching CIE standard)
var porcelainWavelengths = []float64{
	380, 385, 390, 395, 400, 405, 410, 415, 420, 425,
	430, 435, 440, 445, 450, 455, 460, 465, 470, 475,
	480, 485, 490, 495, 500, 505, 510, 515, 520, 525,
	530, 535, 540, 545, 550, 555, 560, 565, 570, 575,
	580, 585, 590, 595, 600, 605, 610, 615, 620, 625,
	630, 635, 640, 645, 650, 655, 660, 665, 670, 675,
	680, 685, 690, 695, 700, 705, 710, 715, 720, 725,
	730, 735, 740, 745, 750,
}

// CreatePorcelain creates a porcelain material with default properties
// - Base: High reflectance white with warm tone
// - Roughness: Semi-glossy (0.15 - smooth but not mirror-like)
// - Metalness: 0 (dielectric)
// - SSS: Moderate subsurface scattering (0.05 strength, 0.1 radius)
func CreatePorcelain() material.Material {
	return CreatePorcelainCustom(0.15, 0.05, 0.1)
}

// CreatePorcelainCustom creates a porcelain material with custom parameters
// roughness: 0.0 (glossy) to 1.0 (matte)
// sssStrength: subsurface scattering strength (0.0 to 0.2 typical)
// sssRadius: subsurface scattering radius (0.0 to 0.3 typical)
func CreatePorcelainCustom(roughness, sssStrength, sssRadius float64) material.Material {
	// Create spectral albedo texture using tabulated data
	spd := spectral.NewSPD(porcelainWavelengths, porcelainSpectralReflectance)
	spectralAlbedo := texture.NewSpectralConstantFromSPD(spd)

	// For RGB rendering, use a white color (fallback, not used in spectral mode)
	rgbAlbedo := texture.NewConstant(vec3.Vec3Impl{X: 0.90, Y: 0.90, Z: 0.90})

	// Create roughness texture
	roughnessTexture := texture.NewConstant(vec3.Vec3Impl{X: roughness, Y: roughness, Z: roughness})

	// Metalness is 0 for porcelain (dielectric)
	metalnessTexture := texture.NewConstant(vec3.Vec3Impl{X: 0.0, Y: 0.0, Z: 0.0})

	// Subsurface scattering (porcelain is slightly translucent)
	sssTexture := texture.NewConstant(vec3.Vec3Impl{X: sssStrength, Y: sssStrength, Z: sssStrength})

	return material.NewPBRWithSpectralAlbedo(
		rgbAlbedo,
		spectralAlbedo,
		nil, // no normal map
		roughnessTexture,
		metalnessTexture,
		sssTexture,
		sssRadius,
	)
}

// CreatePorcelainMatte creates a matte porcelain material (higher roughness)
func CreatePorcelainMatte() material.Material {
	return CreatePorcelainCustom(0.4, 0.05, 0.1)
}

// CreatePorcelainGlossy creates a glossy porcelain material (lower roughness)
func CreatePorcelainGlossy() material.Material {
	return CreatePorcelainCustom(0.05, 0.05, 0.1)
}

// MaterialLibrary is the collection of all available built-in materials
var MaterialLibrary = map[string]*MaterialDefinition{
	"porcelain": {
		Name:           "porcelain",
		Description:    "High-quality porcelain with spectral reflectance (semi-glossy white with warm tone)",
		CreateMaterial: CreatePorcelain,
	},
	"porcelain_matte": {
		Name:           "porcelain_matte",
		Description:    "Matte porcelain with higher roughness",
		CreateMaterial: CreatePorcelainMatte,
	},
	"porcelain_glossy": {
		Name:           "porcelain_glossy",
		Description:    "Glossy porcelain with very low roughness",
		CreateMaterial: CreatePorcelainGlossy,
	},
}

// GetMaterial retrieves a material by name from the library
func GetMaterial(name string) (material.Material, bool) {
	def, ok := MaterialLibrary[name]
	if !ok {
		return nil, false
	}
	return def.CreateMaterial(), true
}

// ListMaterials returns a slice of all available material names
func ListMaterials() []string {
	keys := make([]string, 0, len(MaterialLibrary))
	for key := range MaterialLibrary {
		keys = append(keys, key)
	}
	return keys
}

// CreatePorcelainProtobufMaterial creates a protobuf material definition for porcelain
// This is useful for programmatic scene creation
func CreatePorcelainProtobufMaterial() *pb_transport.Material {
	return CreatePorcelainProtobufMaterialCustom("Porcelain")
}

// CreatePorcelainProtobufMaterialCustom creates a protobuf material definition for porcelain with custom name
func CreatePorcelainProtobufMaterialCustom(name string) *pb_transport.Material {
	// Convert float64 slices to float32 for protobuf
	wavelengths32 := make([]float32, len(porcelainWavelengths))
	values32 := make([]float32, len(porcelainSpectralReflectance))
	for i := range porcelainWavelengths {
		wavelengths32[i] = float32(porcelainWavelengths[i])
		values32[i] = float32(porcelainSpectralReflectance[i])
	}

	return &pb_transport.Material{
		Name: name,
		Type: pb_transport.MaterialType_LAMBERT,
		MaterialProperties: &pb_transport.Material_Lambert{
			Lambert: &pb_transport.LambertMaterial{
				AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
					SpectralAlbedo: &pb_transport.SpectralConstantTexture{
						SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
							Tabulated: &pb_transport.TabulatedSpectralConstant{
								Wavelengths: wavelengths32,
								Values:      values32,
							},
						},
					},
				},
			},
		},
	}
}
