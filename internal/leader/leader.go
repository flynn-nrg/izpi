package leader

import (
	"context"
	"image"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/flynn-nrg/izpi/internal/assetprovider"
	"github.com/flynn-nrg/izpi/internal/colours"
	"github.com/flynn-nrg/izpi/internal/config"
	"github.com/flynn-nrg/izpi/internal/discovery"
	"github.com/flynn-nrg/izpi/internal/display"
	"github.com/flynn-nrg/izpi/internal/output"
	"github.com/flynn-nrg/izpi/internal/postprocess"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/render"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/scene"
	"github.com/flynn-nrg/izpi/internal/transport"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	log "github.com/sirupsen/logrus"
)

const (
	displayWindowTitle = "Izpi Render Output"
)

func RunAsLeader(ctx context.Context, cfg *config.Config, standalone bool) {
	var disp display.Display
	var err error
	var canvas image.Image
	var sceneData *scene.Scene
	var protoScene *pb_transport.Scene

	sceneFile, err := os.Open(cfg.Scene)
	if err != nil {
		log.Fatal(err)
	}

	defer sceneFile.Close()

	aspectRatio := float64(cfg.XSize) / float64(cfg.YSize)

	switch filepath.Ext(cfg.Scene) {
	case ".izpi":
		payload, err := io.ReadAll(sceneFile)
		if err != nil {
			log.Fatalf("Error reading scene file: %v", err)
		}
		protoScene = &pb_transport.Scene{}
		err = proto.Unmarshal(payload, protoScene)
		if err != nil {
			log.Fatalf("Error unmarshalling scene: %v", err)
		}
		t := transport.NewTransport(aspectRatio, protoScene, nil, nil)
		sceneData, err = t.ToScene()
		if err != nil {
			log.Fatalf("Error loading scene: %v", err)
		}
	case ".yaml":
		log.Fatalf("YAML scenes are not supported in leader mode")
	default:
		log.Fatalf("Unknown scene file extension: %s", filepath.Ext(cfg.Scene))
	}

	/*
		protoScene := &pb_transport.Scene{
			Name:    "Cornell Box",
			Version: "1.0.0",
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
				Aspect:    1,
				Aperture:  0,
				Focusdist: 10,
				Time0:     0,
				Time1:     1,
			},
			Objects: &pb_transport.SceneObjects{
				Triangles: []*pb_transport.Triangle{
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 100,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 100,
						},
						Vertex2: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 0,
						},
						Vertex1: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 0,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 0,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 100,
						},
						MaterialName: "White",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 0,
						},
						Vertex2: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 0,
						},
						MaterialName: "Green",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 0,
							Y: 100,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 100,
						},
						Vertex2: &pb_transport.Vec3{
							X: 0,
							Y: 0,
							Z: 0,
						},
						MaterialName: "Green",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 0,
						},
						Vertex1: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 100,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 0,
						},
						MaterialName: "Red",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 100,
						},
						Vertex1: &pb_transport.Vec3{
							X: 100,
							Y: 100,
							Z: 100,
						},
						Vertex2: &pb_transport.Vec3{
							X: 100,
							Y: 0,
							Z: 0,
						},
						MaterialName: "Red",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 33,
							Y: 99,
							Z: 33,
						},
						Vertex1: &pb_transport.Vec3{
							X: 66,
							Y: 99,
							Z: 33,
						},
						Vertex2: &pb_transport.Vec3{
							X: 66,
							Y: 99,
							Z: 66,
						},
						MaterialName: "white_light",
					},
					{
						Vertex0: &pb_transport.Vec3{
							X: 33,
							Y: 99,
							Z: 33,
						},
						Vertex1: &pb_transport.Vec3{
							X: 66,
							Y: 99,
							Z: 66,
						},
						Vertex2: &pb_transport.Vec3{
							X: 33,
							Y: 99,
							Z: 66,
						},
						MaterialName: "white_light",
					},
				},
				Spheres: []*pb_transport.Sphere{
					{
						Center: &pb_transport.Vec3{
							X: 30,
							Y: 15,
							Z: 30,
						},
						Radius:       15,
						MaterialName: "Glass",
					},
					{
						Center: &pb_transport.Vec3{
							X: 70,
							Y: 20,
							Z: 60,
						},
						Radius:       20,
						MaterialName: "Marine Blue",
					},
				},
			},
			Materials: map[string]*pb_transport.Material{
				"White": {
					Name: "White",
					Type: pb_transport.MaterialType_LAMBERT,
					MaterialProperties: &pb_transport.Material_Lambert{
						Lambert: &pb_transport.LambertMaterial{
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
				"Green": {
					Name: "Green",
					Type: pb_transport.MaterialType_LAMBERT,
					MaterialProperties: &pb_transport.Material_Lambert{
						Lambert: &pb_transport.LambertMaterial{
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
				"Red": {
					Name: "Red",
					Type: pb_transport.MaterialType_LAMBERT,
					MaterialProperties: &pb_transport.Material_Lambert{
						Lambert: &pb_transport.LambertMaterial{
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
				"white_light": {
					Name: "white_light",
					Type: pb_transport.MaterialType_DIFFUSE_LIGHT,
					MaterialProperties: &pb_transport.Material_Diffuselight{
						Diffuselight: &pb_transport.DiffuseLightMaterial{
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
				"Glass": {
					Name: "Glass",
					Type: pb_transport.MaterialType_DIELECTRIC,
					MaterialProperties: &pb_transport.Material_Dielectric{
						Dielectric: &pb_transport.DielectricMaterial{
							Refidx: 1.5,
						},
					},
				},
				"Marine Blue": {
					Name: "Marine Blue",
					Type: pb_transport.MaterialType_LAMBERT,
					MaterialProperties: &pb_transport.Material_Lambert{
						Lambert: &pb_transport.LambertMaterial{
							Albedo: &pb_transport.Texture{
								TextureProperties: &pb_transport.Texture_Constant{
									Constant: &pb_transport.ConstantTexture{
										Value: &pb_transport.Vec3{
											X: 0,
											Y: 0.26666666666666666,
											Z: 0.5058823529411764,
										},
									},
								},
							},
						},
					},
				},
			},
		}
	*/

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

	remoteWorkers := make([]*render.RemoteWorkerConfig, 0)

	if !standalone {
		jobID := uuid.New().String()

		discovery, err := discovery.New(time.Second * time.Duration(cfg.DiscoveryTimeout))
		if err != nil {
			log.Fatalln("Failed to initialize discovery:", err.Error())
		}

		workerHosts, err := discovery.FindWorkers()
		if err != nil {
			log.Fatalln("Failed to find workers:", err.Error())
		}

		log.Infof("Found %d worker(s)", len(workerHosts))

		assetProvider, assetProviderAddress, err := assetprovider.New(protoScene, nil, nil)
		if err != nil {
			log.Fatalln("Failed to create asset provider:", err.Error())
		}

		for target, workerHost := range workerHosts {
			log.Infof("Setting up worker: %s", workerHost.GetNodeName())
			conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
			controlClient := pb_control.NewRenderControlServiceClient(conn)
			if err != nil {
				log.Errorf("failed to create control client for worker %s: %v", target, err)
				continue
			}

			stream, err := controlClient.RenderSetup(ctx, &pb_control.RenderSetupRequest{
				SceneName:       protoScene.GetName(),
				JobId:           jobID,
				NumCores:        uint32(workerHost.GetAvailableCores()),
				SamplesPerPixel: uint32(cfg.Samples),
				Sampler:         stringToSamplerType(cfg.Sampler),
				ImageResolution: &pb_control.ImageResolution{
					Width:  uint32(cfg.XSize),
					Height: uint32(cfg.YSize),
				},
				MaxDepth: uint32(cfg.Depth),
				BackgroundColor: &pb_control.Vec3{
					X: 0,
					Y: 0,
					Z: 0,
				},
				InkColor: &pb_control.Vec3{
					X: 1,
					Y: 1,
					Z: 1,
				},
				AssetProvider: assetProviderAddress,
			})
			if err != nil {
				log.Errorf("failed to create render setup stream for worker %s: %v", target, err)
				continue
			}

			for {
				msg, err := stream.Recv()
				if err != nil {
					if err == io.EOF {
						log.Infof("Worker %s finished render setup", target)
						break
					}

					log.Errorf("failed to receive message from worker %s: %v", target, err)
					break
				}

				log.Infof("Worker %s status: %s", target, msg.GetStatus().String())
			}

			remoteWorkers = append(remoteWorkers, &render.RemoteWorkerConfig{
				Client:   controlClient,
				NumCores: int(workerHost.GetAvailableCores()),
			})
		}

		// Free up resources
		assetProvider.Stop()
		protoScene = nil

		log.Info("Finished setting up remote workers")
	}

	r := render.New(sceneData, int(cfg.XSize), int(cfg.YSize), int(cfg.Samples), int(cfg.Depth),
		colours.Black, colours.White, int(cfg.NumWorkers), remoteWorkers, cfg.Verbose, previewChan, cfg.Preview, sampler.StringToType(cfg.Sampler))

	wg := sync.WaitGroup{}
	wg.Add(1)
	// Detach the renderer as SDL needs to use the main thread for everything.
	go func() {
		canvas = r.Render(ctx)
		wg.Done()
	}()

	if cfg.Preview {
		switch cfg.DisplayMode {
		case "fyne":
			disp = display.NewFyneDisplay(displayWindowTitle, int(cfg.XSize), int(cfg.YSize), previewChan)
			disp.Start()
		case "sdl":
			disp = display.NewSDLDisplay(displayWindowTitle, int(cfg.XSize), int(cfg.YSize), previewChan)
			disp.Start()
		default:
			log.Fatalf("unknown display mode %q", cfg.DisplayMode)
		}
	}

	wg.Wait()

	switch cfg.OutputMode {
	case "png":
		pp := postprocess.NewPipeline([]postprocess.Filter{
			postprocess.NewGamma(),
			postprocess.NewClamp(1.0),
		})
		err = pp.Apply(canvas, sceneData)
		if err != nil {
			log.Fatal(err)
		}

		// Output
		out, err := output.NewPNG(cfg.OutputFile)
		if err != nil {
			log.Fatal(err)
		}

		err = out.Write(canvas)
		if err != nil {
			log.Fatal(err)
		}

	case "exr":
		outFileName := strings.Replace(cfg.OutputFile, "png", "exr", 1)
		out, err := output.NewOIIO(outFileName)
		if err != nil {
			log.Fatal(err)
		}

		err = out.Write(canvas)
		if err != nil {
			log.Fatal(err)
		}
	}

	if cfg.Preview {
		disp.Wait()
	}
}

func stringToSamplerType(s string) pb_control.SamplerType {
	switch s {
	case "colour":
		return pb_control.SamplerType_COLOUR
	case "albedo":
		return pb_control.SamplerType_ALBEDO
	case "normal":
		return pb_control.SamplerType_NORMAL
	case "wireframe":
		return pb_control.SamplerType_WIRE_FRAME
	default:
		log.Fatalf("unknown sampler type %q", s)
	}

	return pb_control.SamplerType_SAMPLER_TYPE_UNSPECIFIED
}
