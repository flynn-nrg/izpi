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
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

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

	sceneFile, err := os.Open(cfg.Scene)
	if err != nil {
		log.Fatal(err)
	}

	scene, err := scene.FromYAML(sceneFile, filepath.Dir(cfg.Scene), 0)
	if err != nil {
		log.Fatalf("Error loading scene: %v", err)
	}

	previewChan := make(chan display.DisplayTile)
	defer close(previewChan)

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

		protoScene := &pb_transport.Scene{
			Name:    "Test Scene",
			Version: "1.0.0",
			Camera: &pb_transport.Camera{
				Lookfrom: &pb_transport.Vec3{
					X: 0,
					Y: 0,
					Z: 0,
				},
				Lookat: &pb_transport.Vec3{
					X: 0,
					Y: 0,
					Z: 0,
				},
				Vup: &pb_transport.Vec3{
					X: 0,
					Y: 0,
					Z: 0,
				},
				Vfov:   90,
				Aspect: 1,
			},
		}

		assetProvider, assetProviderAddress, err := assetprovider.New(protoScene, nil, nil)
		if err != nil {
			log.Fatalln("Failed to create asset provider:", err.Error())
		}

		defer assetProvider.Stop()

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
		}

	}

	log.Info("Finished querying all specified workers.")
	// In a real leader, you'd likely keep track of these workers and their status

	r := render.New(scene, int(cfg.XSize), int(cfg.YSize), int(cfg.Samples), int(cfg.Depth),
		colours.Black, colours.White, int(cfg.NumWorkers), cfg.Verbose, previewChan, cfg.Preview, sampler.StringToType(cfg.Sampler))

	wg := sync.WaitGroup{}
	wg.Add(1)
	// Detach the renderer as SDL needs to use the main thread for everything.
	go func() {
		canvas = r.Render()
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
		err = pp.Apply(canvas, scene)
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
