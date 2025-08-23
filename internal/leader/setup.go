package leader

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/flynn-nrg/izpi/internal/assetprovider"
	"github.com/flynn-nrg/izpi/internal/config"
	"github.com/flynn-nrg/izpi/internal/discovery"
	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/render"
	"github.com/flynn-nrg/izpi/internal/texture"
	"github.com/google/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func setupWorkers(ctx context.Context, cfg *config.Config, protoScene *pb_transport.Scene, textures map[string]*texture.ImageTxt) ([]*render.RemoteWorkerConfig, error) {
	jobID := uuid.New().String()

	remoteWorkers := make([]*render.RemoteWorkerConfig, 0)

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
		return nil, fmt.Errorf("failed to create asset provider: %w", err)
	}

	for target, workerHost := range workerHosts {
		log.Infof("Setting up worker: %s", workerHost.GetNodeName())
		conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
		controlClient := pb_control.NewRenderControlServiceClient(conn)
		if err != nil {
			log.Errorf("failed to create control client for worker %s: %v", target, err)
			continue
		}

		// Convert scene's spectral background to control protobuf format
		var spectralBackground *pb_control.SpectralBackground
		if protoScene.GetSpectralBackground() != nil {
			spectralBackground = &pb_control.SpectralBackground{
				SpectralProperties: &pb_control.SpectralBackground_Tabulated{
					Tabulated: &pb_control.TabulatedSpectralConstant{
						Wavelengths: make([]float32, len(protoScene.GetSpectralBackground().GetWavelengths())),
						Values:      make([]float32, len(protoScene.GetSpectralBackground().GetValues())),
					},
				},
			}
			// Convert float32 to float64
			for i, w := range protoScene.GetSpectralBackground().GetWavelengths() {
				spectralBackground.GetTabulated().Wavelengths[i] = w
			}
			for i, v := range protoScene.GetSpectralBackground().GetValues() {
				spectralBackground.GetTabulated().Values[i] = v
			}
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
			AssetProvider:      assetProviderAddress,
			SpectralBackground: spectralBackground,
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

	assetProvider.Stop()

	log.Info("Finished setting up remote workers")

	return remoteWorkers, nil
}
