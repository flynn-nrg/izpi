package scenes

import pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"

func CornellBoxEmptyDisplacementSpectral(aspect float32) *pb_transport.Scene {
	protoScene := &pb_transport.Scene{
		Name:                 "Cornell Box Empty Displacement Spectral",
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
		},
		Objects: &pb_transport.SceneObjects{
			Triangles: []*pb_transport.Triangle{
				// Back wall (White)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					MaterialName: "White",
				},
				// Floor (White) - Positioned slightly below water (Y=-0.1) to allow caustics to form
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 1},
					Uv1:          &pb_transport.Vec2{U: 0, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 0},
					MaterialName: "White",
				},
				// Additional floor triangle
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 1},
					Uv1:          &pb_transport.Vec2{U: 1, V: 0},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "White",
				},
				// Water surface triangles (40% of box height, aligned with water box front)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 0, V: 1},
					Uv2:          &pb_transport.Vec2{U: 1, V: 1},
					MaterialName: "Water",

					Operator: pb_transport.GeometryOperator_DISPLACE,
					OperatorProperties: &pb_transport.Triangle_Displace{
						Displace: &pb_transport.DisplaceOperator{
							DisplacementMap: "water_128b.png",
							Min:             -1,
							Max:             0,
						},
					},
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					Uv0:          &pb_transport.Vec2{U: 0, V: 0},
					Uv1:          &pb_transport.Vec2{U: 1, V: 1},
					Uv2:          &pb_transport.Vec2{U: 1, V: 0},
					MaterialName: "Water",

					Operator: pb_transport.GeometryOperator_DISPLACE,
					OperatorProperties: &pb_transport.Triangle_Displace{
						Displace: &pb_transport.DisplaceOperator{
							DisplacementMap: "water_128b.png",
							Min:             -1,
							Max:             0,
						},
					},
				},
				// Water box front face (up to 40% height, aligned with water surface)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					MaterialName: "Water",
				},
				// Water box back face (at Z=100)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "Water",
				},
				// Water box left face (at X=0)
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 40, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					MaterialName: "Water",
				},
				// Water box right face (at X=100)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 0},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 40, Z: 100},
					MaterialName: "Water",
				},
				// Water box bottom face (at Y=0) - CRITICAL FOR CAUSTICS
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 0},
					MaterialName: "Water",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 0, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: 0, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 0, Z: 100},
					MaterialName: "Water",
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
					Vertex1:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 0},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: 100, Z: 0},
					MaterialName: "Green",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 0, Y: 100, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 0, Y: -0.1, Z: 0},
					MaterialName: "Green",
				},
				// Right wall (Red)
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 0},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: 100, Z: 0},
					MaterialName: "Red",
				},
				{
					Vertex0:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 100},
					Vertex1:      &pb_transport.Vec3{X: 100, Y: 100, Z: 100},
					Vertex2:      &pb_transport.Vec3{X: 100, Y: -0.1, Z: 0},
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
			"Water": {
				Name: "Water",
				Type: pb_transport.MaterialType_DIELECTRIC,
				MaterialProperties: &pb_transport.Material_Dielectric{
					Dielectric: &pb_transport.DielectricMaterial{
						RefractiveIndexProperties: &pb_transport.DielectricMaterial_SpectralRefidx{
							SpectralRefidx: &pb_transport.SpectralConstantTexture{
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
										// Water refractive indices (approximate values for visible spectrum)
										Values: []float32{1.344, 1.343, 1.342, 1.341, 1.340, 1.339, 1.338, 1.337, 1.336, 1.335, 1.334, 1.333, 1.332, 1.331, 1.330, 1.329, 1.328, 1.327, 1.326, 1.325, 1.324, 1.323, 1.322, 1.321, 1.320, 1.319, 1.318, 1.317, 1.316, 1.315, 1.314, 1.313, 1.312, 1.311, 1.310, 1.309, 1.308, 1.307},
									},
								},
							},
						},
						//ComputeBeerLambertAttenuation: true,
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
								SpectralProperties: &pb_transport.SpectralConstantTexture_Tabulated{
									Tabulated: &pb_transport.TabulatedSpectralConstant{
										// Wavelengths in nanometers (visible spectrum)
										Wavelengths: []float32{380, 390, 400, 410, 420, 430, 440, 450, 460, 470, 480, 490, 500, 510, 520, 530, 540, 550, 560, 570, 580, 590, 600, 610, 620, 630, 640, 650, 660, 670, 680, 690, 700, 710, 720, 730, 740, 750},
										// White light spectrum (daylight-like, balanced across all wavelengths)
										Values: []float32{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0},
									},
								},
							},
						},
					},
				},
			},
		},
		DisplacementMaps: map[string]*pb_transport.ImageTextureMetadata{
			"water_128b.png": {
				Filename:    "water_128b.png",
				Width:       130,
				Height:      130,
				Channels:    4,
				PixelFormat: pb_transport.TexturePixelFormat_float32,
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
