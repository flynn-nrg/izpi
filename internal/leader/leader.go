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
	"github.com/flynn-nrg/izpi/internal/scenes"
	"github.com/flynn-nrg/izpi/internal/texture"
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

	triangleStreamThreshold = 10000
)

func RunAsLeader(ctx context.Context, cfg *config.Config, standalone bool) {
	var disp display.Display
	var err error
	var canvas image.Image
	var sceneData *scene.Scene
	var protoScene *pb_transport.Scene

	aspectRatio := float64(cfg.XSize) / float64(cfg.YSize)

	sceneFile, err := os.Open(cfg.Scene)
	if err != nil {
		log.Fatal(err)
	}

	defer sceneFile.Close()

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

	protoScene = scenes.CornellBoxPB(aspectRatio)

	// Load textures
	textures := make(map[string]*texture.ImageTxt)
	for _, t := range protoScene.GetImageTextures() {
		log.Infof("Loading texture %s", t.GetFilename())
		imageText, err := texture.NewFromFile(t.GetFilename())
		if err != nil {
			log.Fatalf("Error loading texture %s: %v", t.GetFilename(), err)
		}
		textures[t.GetFilename()] = imageText

		// Update metadata. The pixel format is always float64 with 4 channels.
		t.Width = uint32(imageText.GetData().Bounds().Dx())
		t.Height = uint32(imageText.GetData().Bounds().Dy())
		t.PixelFormat = pb_transport.TexturePixelFormat_FLOAT64
		t.Channels = 4
	}

	t := transport.NewTransport(aspectRatio, protoScene, nil, textures)
	sceneData, err = t.ToScene()
	if err != nil {
		log.Fatalf("Error loading scene: %v", err)
	}

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

	remoteWorkers := make([]*render.RemoteWorkerConfig, 0)

	if len(protoScene.Objects.Triangles) > triangleStreamThreshold {
		protoScene.StreamTriangles = true
	}

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

		var trianglesToStream []*pb_transport.Triangle

		if len(protoScene.Objects.Triangles) > triangleStreamThreshold {
			protoScene.StreamTriangles = true
			trianglesToStream = protoScene.Objects.Triangles
			protoScene.TotalTriangles = uint64(len(trianglesToStream))
			protoScene.Objects.Triangles = nil
		}

		assetProvider, assetProviderAddress, err := assetprovider.New(protoScene, textures, trianglesToStream)
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
