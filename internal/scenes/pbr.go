package scenes

import (
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
)

func CornellBoxPBRRGB(aspect float64) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box PBR RGB",
		Version:              "1.0.0",
		ColourRepresentation: pb_transport.ColourRepresentation_RGB,
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
				// Floor (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				// Additional floor triangles
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
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
				// Glass sphere
				{
					Center: &pb_transport.Vec3{
						X: 30,
						Y: 15,
						Z: 30,
					},
					Radius:       15,
					MaterialName: "Glass",
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
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.73,
											Y: 0.73,
											Z: 0.73,
										},
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
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.0,
											Y: 0.73,
											Z: 0.0,
										},
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
						AlbedoProperties: &pb_transport.LambertMaterial_Albedo{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0.73,
											Y: 0.0,
											Z: 0.0,
										},
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
						EmissionProperties: &pb_transport.DiffuseLightMaterial_Emit{
							Emit: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 15,
											Y: 15,
											Z: 15,
										},
									},
								},
							},
						},
					},
				},
			},
			"Glass": {
				Name: "Glass",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_Refidx{
							Refidx: 1.40,
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
	}

	return protoScene
}
