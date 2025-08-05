package leader

import (
	"context"
	"fmt"
	"image"
	"strings"
	"sync"

	"github.com/flynn-nrg/izpi/internal/colours"
	"github.com/flynn-nrg/izpi/internal/config"
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
	/*
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
		case ".yaml":
			log.Fatalf("YAML scenes are not supported in leader mode")
		default:
			log.Fatalf("Unknown scene file extension: %s", filepath.Ext(cfg.Scene))
		}
	*/

	protoScene = scenes.CornellBoxPBRSpectral(aspectRatio)

	// Override the colour sampler if the scene is spectral.
	if protoScene.GetColourRepresentation() == pb_transport.ColourRepresentation_SPECTRAL && cfg.Sampler == "colour" {
		cfg.Sampler = "spectral"
	}

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

	if len(protoScene.Objects.Triangles) > triangleStreamThreshold {
		protoScene.StreamTriangles = true
	}

	var remoteWorkers []*render.RemoteWorkerConfig

	if !standalone {
		remoteWorkers, err = setupWorkers(ctx, cfg, protoScene, textures)
		if err != nil {
			log.Fatalf("failed to setup workers: %v", err)
		}
	}

	// Free up resources
	protoScene = nil

	fmt.Println("Rendering scene, sampler type: ", cfg.Sampler)
	r := render.New(
		sceneData,
		int(cfg.XSize), int(cfg.YSize),
		int(cfg.Samples), int(cfg.Depth),
		colours.Black,
		colours.White,
		colours.SpectralBlack,
		int(cfg.NumWorkers), remoteWorkers, cfg.Verbose, previewChan, cfg.Preview, sampler.StringToType(cfg.Sampler))

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
	case "spectral":
		return pb_control.SamplerType_SPECTRAL
	default:
		log.Fatalf("unknown sampler type %q", s)
	}

	return pb_control.SamplerType_SAMPLER_TYPE_UNSPECIFIED
}
