package scenes

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/vec3"
	"github.com/flynn-nrg/izpi/internal/wavefront"
	log "github.com/sirupsen/logrus"
)

func CornellBoxPBRColouredGlassSpectral(aspect float64) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box PBR Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
			WhiteBalance: &pb_transport.WhiteBalance{
				WhiteBalanceProperties: &pb_transport.WhiteBalance_WhiteBalanceFromLightSource{
					WhiteBalanceFromLightSource: &pb_transport.WhiteBalanceFromLightSource{
						LightSourceName: "cie_illuminant_a_2856k",
					},
				},
			},
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White) - Triangle 1: covers left half
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Floor (White) - Triangle 2: covers right half
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					MaterialName: "White",
				},

				// Ceiling triangles
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "Red",
				},
				// Light source
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Red glass sphere
				{
					Center: &pb_transport.Vec3{
						X: 30,
						Y: 15,
						Z: 30,
					},
					Radius:       15,
					MaterialName: "RedGlass",
				},
				// PBR spheres from the YAML file
				{
					Center: &pb_transport.Vec3{
						X: 65,
						Y: 20,
						Z: 60,
					},
					Radius:       20,
					MaterialName: "rusty-metal",
				},
				{
					Center: &pb_transport.Vec3{
						X: 85,
						Y: 12,
						Z: 20,
					},
					Radius:       12,
					MaterialName: "grainy-concrete",
				},
				{
					Center: &pb_transport.Vec3{
						X: 58,
						Y: 10,
						Z: 27,
					},
					Radius:       10,
					MaterialName: "fleshy_granite1",
				},
				{
					Center: &pb_transport.Vec3{
						X: 65,
						Y: 6,
						Z: 9,
					},
					Radius:       6,
					MaterialName: "bamboo-wood-semigloss",
				},
				{
					Center: &pb_transport.Vec3{
						X: 45,
						Y: 5,
						Z: 10,
					},
					Radius:       5,
					MaterialName: "lightgold",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Changed from 0.73 to 0.9
										CenterWavelength: 550.0, // Green wavelength
										Width:            50.0,  // Narrow bandwidth for green
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Changed from 0.73 to 0.9 for more vivid red
										CenterWavelength: 650.0, // Red wavelength
										Width:            50.0,  // Narrow bandwidth for red
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_FromLightSourceLibrary{
									FromLightSourceLibrary: &pb_transport.FromLightSourceLibrary{
										LightSourceName: "cie_illuminant_a_2856k",
									},
								},
							},
						},
					},
				}},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for typical glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
					},
				},
			},
			"RedGlass": {
				Name: "RedGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.15,  // Reduced absorption for more transparency while maintaining red tint
										CenterWavelength: 530.0, // Green wavelength (absorb green, transmit red)
										Width:            80.0,  // Broad absorption in green region
									},
								},
							},
						},
					},
				},
			},
			"rusty-metal": {
				Name: "rusty-metal",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/rusty-metal_albedo.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/rusty-metal_roughness.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/rusty-metal_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/rusty-metal_normal-ogl.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
			"grainy-concrete": {
				Name: "grainy-concrete",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/grainy-concrete_albedo.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/grainy-concrete_roughness.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/grainy-concrete_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/grainy-concrete_normal-ogl.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
			"fleshy_granite1": {
				Name: "fleshy_granite1",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/fleshy_granite1_albedo.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/fleshy_granite1_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/fleshy_granite1_normal-ogl.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.1, Y: 0.1, Z: 0.1},
								},
							},
						},
						SssRadius: 0.1,
					},
				},
			},
			"bamboo-wood-semigloss": {
				Name: "bamboo-wood-semigloss",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/bamboo-wood-semigloss-albedo.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/bamboo-wood-semigloss-roughness.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/bamboo-wood-semigloss-metal.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/bamboo-wood-semigloss-normal.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
			"lightgold": {
				Name: "lightgold",
				Type: pb_transport.MaterialType_PBR,
				MaterialProperties: &pb_transport.Material_Pbr{
					Pbr: &pb_transport.PBRMaterial{
						Albedo: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/lightgold_albedo.png",
								},
							},
						},
						Roughness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/lightgold_roughness.png",
								},
							},
						},
						Metalness: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/lightgold_metallic.png",
								},
							},
						},
						NormalMap: &pb_transport.Texture{
							Type: pb_transport.TextureType_IMAGE,
							TextureProperties: &pb_transport.Texture_Image{
								Image: &pb_transport.ImageTexture{
									Filename: "textures/lightgold_normal-ogl.png",
								},
							},
						},
						Sss: &pb_transport.Texture{
							Type: pb_transport.TextureType_CONSTANT,
							TextureProperties: &pb_transport.Texture_Constant{
								Constant: &pb_transport.ConstantTexture{
									Value: &pb_transport.Vec3{X: 0.0, Y: 0.0, Z: 0.0},
								},
							},
						},
						SssRadius: 0.0,
					},
				},
			},
		},
		ImageTextures: map[string]*pb_transport.ImageTextureMetadata{
			"textures/rusty-metal_albedo.png": {
				Filename: "textures/rusty-metal_albedo.png",
			},
			"textures/rusty-metal_roughness.png": {
				Filename: "textures/rusty-metal_roughness.png",
			},
			"textures/rusty-metal_metallic.png": {
				Filename: "textures/rusty-metal_metallic.png",
			},
			"textures/rusty-metal_normal-ogl.png": {
				Filename: "textures/rusty-metal_normal-ogl.png",
			},
			"textures/grainy-concrete_albedo.png": {
				Filename: "textures/grainy-concrete_albedo.png",
			},
			"textures/grainy-concrete_roughness.png": {
				Filename: "textures/grainy-concrete_roughness.png",
			},
			"textures/grainy-concrete_metallic.png": {
				Filename: "textures/grainy-concrete_metallic.png",
			},
			"textures/grainy-concrete_normal-ogl.png": {
				Filename: "textures/grainy-concrete_normal-ogl.png",
			},
			"textures/fleshy_granite1_albedo.png": {
				Filename: "textures/fleshy_granite1_albedo.png",
			},
			"textures/fleshy_granite1_metallic.png": {
				Filename: "textures/fleshy_granite1_metallic.png",
			},
			"textures/fleshy_granite1_normal-ogl.png": {
				Filename: "textures/fleshy_granite1_normal-ogl.png",
			},
			"textures/bamboo-wood-semigloss-albedo.png": {
				Filename: "textures/bamboo-wood-semigloss-albedo.png",
			},
			"textures/bamboo-wood-semigloss-roughness.png": {
				Filename: "textures/bamboo-wood-semigloss-roughness.png",
			},
			"textures/bamboo-wood-semigloss-metal.png": {
				Filename: "textures/bamboo-wood-semigloss-metal.png",
			},
			"textures/bamboo-wood-semigloss-normal.png": {
				Filename: "textures/bamboo-wood-semigloss-normal.png",
			},
			"textures/lightgold_albedo.png": {
				Filename: "textures/lightgold_albedo.png",
			},
			"textures/lightgold_roughness.png": {
				Filename: "textures/lightgold_roughness.png",
			},
			"textures/lightgold_metallic.png": {
				Filename: "textures/lightgold_metallic.png",
			},
			"textures/lightgold_normal-ogl.png": {
				Filename: "textures/lightgold_normal-ogl.png",
			},
		},
		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	return protoScene
}

func CornellBoxPBRStanfordDragonSpectral(aspect float64) *pb_transport.Scene {

	dragonReader, err := os.Open("meshes/dragon_tri.obj")
	if err != nil {
		log.Fatalf("Error opening dragon mesh: %v", err)
	}
	defer dragonReader.Close()

	dragonMesh, err := wavefront.NewObjFromReader(dragonReader, filepath.Dir("meshes/dragon.obj"))
	if err != nil {
		log.Fatalf("Error parsing dragon mesh: %v", err)
	}

	dragonMesh.Scale(vec3.Vec3Impl{X: 90.0, Y: 90.0, Z: 90.0})
	dragonMesh.Rotate(0.0, -(60.0 * math.Pi / 180.0), 0.0)
	dragonMesh.Translate(vec3.Vec3Impl{X: 50.0, Y: 50.0, Z: 60.0})

	var dragonTriangles []*pb_transport.Triangle

	for i := 0; i < dragonMesh.NumGroups(); i++ {
		triangles, err := dragonMesh.GroupToTransportTrianglesWithMaterial(i, "BlueGlass")
		if err != nil {
			log.Fatalf("Error converting dragon mesh to transport triangles: %v", err)
		}
		dragonTriangles = append(dragonTriangles, triangles...)
	}

	log.Infof("Converted %v dragon triangles", len(dragonTriangles))

	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box PBR Stanford Dragon Spectral",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_SPECTRAL,
		Camera: &pb_transport.Camera{
			Lookfrom: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: -140,
			},
			Lookat: &pb_transport.Vec3{
				X: 50,
				Y: 50,
				Z: 0,
			},
			Vup: &pb_transport.Vec3{
				X: 0,
				Y: 1,
				Z: 0,
			},
			Vfov:      40,
			Aspect:    float32(aspect),
			Aperture:  0,
			Focusdist: 10,
			Time0:     0,
			Time1:     1,
			WhiteBalance: &pb_transport.WhiteBalance{
				WhiteBalanceProperties: &pb_transport.WhiteBalance_WhiteBalanceFromLightSource{
					WhiteBalanceFromLightSource: &pb_transport.WhiteBalanceFromLightSource{
						LightSourceName: "cie_illuminant_a_2856k",
					},
				},
			},
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White) - Triangle 1: covers left half
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Floor (White) - Triangle 2: covers right half
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					MaterialName: "White",
				},

				// Ceiling triangles
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Left wall (Green)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "Red",
				},
				// Light source
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 33},
					Vertex2:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 33, Y: 99, Z: 33},
					Vertex1:      &pb_transport.Vec3{X: 66, Y: 99, Z: 66},
					Vertex2:      &pb_transport.Vec3{X: 33, Y: 99, Z: 66},
					MaterialName: "white_light",
				},
			},
			Spheres: []*pb_transport.Sphere{
				// Red glass sphere
				{
					Center: &pb_transport.Vec3{
						X: 30,
						Y: 15,
						Z: 30,
					},
					Radius:       15,
					MaterialName: "Glass",
				},
			},
		},
		Materials: map[string]*pb_transport.Material{
			"White": {
				Name: "White",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Neutral{
									Neutral: &pb_transport.NeutralSpectralConstant{
										Reflectance: 0.73,
									},
								},
							},
						},
					},
				},
			},
			"Green": {
				Name: "Green",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Changed from 0.73 to 0.9
										CenterWavelength: 550.0, // Green wavelength
										Width:            50.0,  // Narrow bandwidth for green
									},
								},
							},
						},
					},
				},
			},
			"Red": {
				Name: "Red",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Changed from 0.73 to 0.9 for more vivid red
										CenterWavelength: 650.0, // Red wavelength
										Width:            50.0,  // Narrow bandwidth for red
									},
								},
							},
						},
					},
				},
			},
			"Blue": {
				Name: "Blue",
				Type: pb_transport.MaterialType_LAMBERT,
				MaterialProperties: &pb_transport.Material_Lambert{
					Lambert: &pb_transport.LambertMaterial{
						AlbedoProperties: &pb_transport.LambertMaterial_SpectralAlbedo{
							SpectralAlbedo: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.9,   // Vivid blue
										CenterWavelength: 470.0, // Blue wavelength
										Width:            50.0,  // Narrow bandwidth for blue
									},
								},
							},
						},
					},
				},
			},
			"white_light": {
				Name: "white_light",
				Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
				MaterialProperties: &pb_transport.Material_Diffuselight{
					Diffuselight: &pb_transport.DiffuseLightMaterial{
						EmissionProperties: &pb_transport.DiffuseLightMaterial_SpectralEmit{
							SpectralEmit: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_FromLightSourceLibrary{
									FromLightSourceLibrary: &pb_transport.FromLightSourceLibrary{
										LightSourceName: "cie_illuminant_a_2856k",
									},
								},
							},
						},
					},
				}},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for typical glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
					},
				},
			},
			"BlueGlass": {
				Name: "BlueGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.2,   // Absorb yellow/orange wavelengths to transmit blue
										CenterWavelength: 590.0, // Yellow/orange wavelength (absorb to transmit blue)
										Width:            90.0,  // Broad absorption in yellow-red region
									},
								},
							},
						},
					},
				},
			},
			"RedGlass": {
				Name: "RedGlass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 400, 420, 440, 460, 480, 500, 520, 540, 560, 580, 600, 620, 640, 660, 680, 700, 720, 740, 750},
										// Refractive indices for glass (slight dispersion)
										Values: []float32{1.52, 1.51, 1.51, 1.50, 1.50, 1.49, 1.49, 1.48, 1.48, 1.47, 1.47, 1.46, 1.46, 1.45, 1.45, 1.44, 1.44, 1.43, 1.43, 1.42},
									},
								},
							},
						},
						AbsorptionProperties: &pb_transport.DielectricMaterial_SpectralAbsorptionCoeff{
							SpectralAbsorptionCoeff: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Gaussian{
									Gaussian: &pb_transport.GaussianSpectralConstant{
										PeakValue:        0.15,  // Reduced absorption for more transparency while maintaining red tint
										CenterWavelength: 530.0, // Green wavelength (absorb green, transmit red)
										Width:            80.0,  // Broad absorption in green region
									},
								},
							},
						},
					},
				},
			},
		},

		// Spectral background - black (no emission)
		SpectralBackground: &pb_transport.TabulatedSpectralConstant{
			Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
			Values:      []float32{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0},
		},
	}

	protoScene.Objects.Triangles = append(protoScene.Objects.Triangles, dragonTriangles...)

	fmt.Printf("Dragon triangles: %v\n", len(dragonTriangles))
	fmt.Printf("Total triangles: %v\n", len(protoScene.Objects.Triangles))

	return protoScene
}
