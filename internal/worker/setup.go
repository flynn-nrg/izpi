package worker

import (
	"context"
	"fmt"

	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
	pb_transport "github.com/flynn-nrg/izpi/internal/proto/transport"
	"github.com/flynn-nrg/izpi/internal/sampler"
	"github.com/flynn-nrg/izpi/internal/transport"
	"github.com/flynn-nrg/izpi/internal/vec3"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// sendStatus sends a RenderSetupResponse with the given status and an optional error message.
func (s *workerServer) sendStatus(stream pb_control.RenderControlService_RenderSetupServer, cfgStatus pb_control.RenderSetupStatus, errMsg string) error {
	resp := &pb_control.RenderSetupResponse{
		Status:       cfgStatus,
		ErrorMessage: errMsg,
	}
	if err := stream.Send(resp); err != nil {
		log.Errorf("Failed to send RenderSetupResponse status %s (error: %s): %v", cfgStatus.String(), errMsg, err)
		return err
	}
	log.Infof("RenderSetup: Sent status: %s %s", cfgStatus.String(), errMsg)
	return nil
}

// getScene fetches the Scene proto message from the asset provider.
func (s *workerServer) getScene(ctx context.Context, transportClient pb_transport.SceneTransportServiceClient, sceneName string) (*pb_transport.Scene, error) {
	req := &pb_transport.GetSceneRequest{SceneName: sceneName}
	scene, err := transportClient.GetScene(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to get scene '%s': %w", sceneName, err)
	}
	return scene, nil
}

// streamTextureFile streams a texture file from the asset provider.
// `expectedTotalSize` is used for pre-allocation, obtained from ImageTexture.size.
func (s *workerServer) streamTextureFile(ctx context.Context, transportClient pb_transport.SceneTransportServiceClient, filename string, expectedTotalSize uint64) ([]byte, error) {
	req := &pb_transport.StreamTextureFileRequest{
		Filename:  filename,
		Offset:    0,         // Start from beginning
		ChunkSize: 64 * 1024, // Consistent chunk size
	}
	stream, err := transportClient.StreamTextureFile(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to open texture stream for %s: %w", filename, err)
	}

	textureData := make([]byte, 0, expectedTotalSize) // Pre-allocate based on ImageTexture.size
	receivedBytes := uint64(0)

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == context.Canceled {
				return nil, fmt.Errorf("texture stream cancelled for %s", filename)
			}
			if s, ok := status.FromError(err); ok {
				if s.Code() == codes.NotFound {
					return nil, fmt.Errorf("texture '%s' not found on provider", filename)
				}
				if s.Code() == codes.Unavailable {
					log.Warnf("Texture stream for %s closed by server gracefully (Unavailable). Received %d of %d bytes.", filename, receivedBytes, expectedTotalSize)
					break // Server closed stream, assume end.
				}
			}
			if err.Error() == "EOF" { // gRPC stream end
				log.Infof("Finished streaming texture '%s'. Received %d of %d bytes.", filename, receivedBytes, expectedTotalSize)
				break
			}
			return nil, fmt.Errorf("failed to receive texture chunk for %s: %w", filename, err)
		}

		// resp.GetSize() here is the size of the *current chunk*, not the total size.
		// Use len(resp.GetData()) instead if chunk size is meant for individual chunk size
		textureData = append(textureData, resp.GetChunk()...)
		receivedBytes += uint64(len(resp.GetChunk()))
	}

	if expectedTotalSize > 0 && receivedBytes != expectedTotalSize {
		return nil, fmt.Errorf("texture '%s' stream ended prematurely. Expected %d bytes, got %d", filename, expectedTotalSize, receivedBytes)
	}

	return textureData, nil
}

// streamTriangles streams triangles from the asset provider.
func (s *workerServer) streamTriangles(ctx context.Context, transportClient pb_transport.SceneTransportServiceClient, sceneName string, totalTriangles uint64, batchSize uint32) ([]*pb_transport.Triangle, error) {
	allTriangles := make([]*pb_transport.Triangle, 0, totalTriangles) // Pre-allocate for all triangles
	var fetchedCount uint64 = 0

	// Loop to fetch all batches
	for fetchedCount < totalTriangles {
		req := &pb_transport.StreamTrianglesRequest{
			SceneName: sceneName,
			BatchSize: batchSize,
			Offset:    fetchedCount, // Request from the current offset
		}
		stream, err := transportClient.StreamTriangles(ctx, req)
		if err != nil {
			return nil, fmt.Errorf("failed to open triangles stream for scene '%s': %w", sceneName, err)
		}

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == context.Canceled {
					return nil, fmt.Errorf("triangles stream cancelled for scene '%s'", sceneName)
				}
				if s, ok := status.FromError(err); ok {
					if s.Code() == codes.NotFound {
						return nil, fmt.Errorf("triangles for scene '%s' not found on provider", sceneName)
					}
					if s.Code() == codes.Unavailable {
						log.Warnf("Triangles stream for scene '%s' closed by server gracefully (Unavailable). Received %d of %d triangles.", sceneName, fetchedCount, totalTriangles)
						// This might happen if the server decides to close the stream early
						break // Assume end of stream for this batch
					}
				}
				if err.Error() == "EOF" { // gRPC stream end
					log.Infof("Finished streaming triangles for scene '%s'. Fetched %d of %d total triangles.", sceneName, fetchedCount, totalTriangles)
					break
				}
				return nil, fmt.Errorf("failed to receive triangle batch for scene '%s': %w", sceneName, err)
			}

			allTriangles = append(allTriangles, resp.GetTriangles()...)
			fetchedCount += uint64(len(resp.GetTriangles()))

			// If the batch returned fewer than requested (and we haven't hit total), it means we're at the end
			if uint64(len(resp.GetTriangles())) < uint64(batchSize) {
				log.Infof("Received partial triangle batch. Assuming end of stream for scene '%s'. Fetched %d of %d total triangles.", sceneName, fetchedCount, totalTriangles)
				break
			}
		}

		// After receiving all responses for one batch request, if fetchedCount is still less than totalTriangles,
		// the outer loop will continue for the next batch.
		if fetchedCount >= totalTriangles {
			break // All triangles fetched
		}
	}

	if fetchedCount != totalTriangles {
		return nil, fmt.Errorf("triangles stream for scene '%s' ended prematurely. Expected %d triangles, got %d", sceneName, totalTriangles, fetchedCount)
	}

	return allTriangles, nil
}

// RenderSetup is a streaming RPC to configure a worker node.
// It fetches scene data, textures, and triangles from the asset provider.
func (s *workerServer) RenderSetup(req *pb_control.RenderSetupRequest, stream pb_control.RenderControlService_RenderSetupServer) error {
	s.currentStatus = pb_discovery.WorkerStatus_ALLOCATED

	log.Printf("RenderControlService: RenderSetup called by %s", s.workerID)
	log.Printf("RenderSetup Configuration: Scene='%s', Sampler='%s', AssetProvider='%s', JobID='%s'",
		req.GetSceneName(), req.GetSampler().String(), req.GetAssetProvider(), req.GetJobId())

	assetProviderAddr := req.GetAssetProvider()
	if assetProviderAddr == "" {
		errMsg := "Asset provider address is empty"
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.InvalidArgument, errMsg)
	}

	ctx := stream.Context() // Use the stream's context for asset fetching client

	// Establish connection to asset provider (SceneTransportService)
	// Using WithBlock to ensure connection is established before proceeding
	assetConn, err := grpc.NewClient(assetProviderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		errMsg := fmt.Sprintf("Failed to dial asset provider %s: %v", assetProviderAddr, err)
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.Unavailable, errMsg)
	}
	defer assetConn.Close()
	transportClient := pb_transport.NewSceneTransportServiceClient(assetConn)

	s.currentStatus = pb_discovery.WorkerStatus_BUSY_RENDER_SETUP

	// Step 1: Send LOADING_SCENE status and fetch scene file
	if err := s.sendStatus(stream, pb_control.RenderSetupStatus_LOADING_SCENE, ""); err != nil {
		return status.Errorf(codes.Internal, "failed to send LOADING_SCENE status: %v", err)
	}
	log.Infof("RenderSetup: Attempting to fetch scene '%s' from '%s'...", req.GetSceneName(), assetProviderAddr)

	protoScene, err := s.getScene(ctx, transportClient, req.GetSceneName())
	if err != nil {
		errMsg := fmt.Sprintf("Failed to load scene '%s': %v", req.GetSceneName(), err)
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.NotFound, errMsg) // Use NotFound for scene not found
	}

	log.Infof("RenderSetup: Successfully loaded scene '%s' (version: %s). Contains %d materials, %d spheres.",
		protoScene.GetName(), protoScene.GetVersion(), len(protoScene.GetMaterials()), len(protoScene.GetObjects().GetSpheres()))

	triangles := make([]*pb_transport.Triangle, 0)

	if protoScene.GetStreamTriangles() {
		log.Infof("RenderSetup: Scene indicates triangles need to be streamed. Total triangles: %d", protoScene.GetTotalTriangles())
		trianglesChunk, err := s.streamTriangles(ctx, transportClient, protoScene.GetName(), protoScene.GetTotalTriangles(), 1000) // Fetch in batches of 1000
		if err != nil {
			errMsg := fmt.Sprintf("Failed to stream triangles for scene '%s': %v", protoScene.GetName(), err)
			s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
			return status.Error(codes.Internal, errMsg)
		}

		triangles = append(triangles, trianglesChunk...)

		log.Infof("RenderSetup: Successfully streamed %d triangles for scene '%s'.", len(triangles), protoScene.GetName())
	}

	// Step 2: Send STREAMING_TEXTURES status and fetch textures
	if err := s.sendStatus(stream, pb_control.RenderSetupStatus_STREAMING_TEXTURES, ""); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to send STREAMING_TEXTURES status: %v", err))
	}
	log.Infof("RenderSetup: Streaming textures from '%s'...", assetProviderAddr)

	// Collect all unique ImageTexture filenames from materials
	texturesToFetch := make(map[string]*pb_transport.ImageTexture)
	for _, mat := range protoScene.GetMaterials() {
		if mat.GetLambert() != nil && mat.GetLambert().GetAlbedo() != nil {
			if imgTex := mat.GetLambert().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		// Add similar logic for other material types and their textures (e.g., DiffuseLight, Metal, PBR)
		if mat.GetDiffuselight() != nil && mat.GetDiffuselight().GetEmit() != nil {
			if imgTex := mat.GetDiffuselight().GetEmit().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		if mat.GetIsotropic() != nil && mat.GetIsotropic().GetAlbedo() != nil {
			if imgTex := mat.GetIsotropic().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		if mat.GetMetal() != nil && mat.GetMetal().GetAlbedo() != nil {
			if imgTex := mat.GetMetal().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
		if mat.GetPbr() != nil {
			if imgTex := mat.GetPbr().GetAlbedo().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetRoughness().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetMetalness().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetNormalMap().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
			if imgTex := mat.GetPbr().GetSss().GetImage(); imgTex != nil {
				texturesToFetch[imgTex.GetFilename()] = imgTex
			}
		}
	}

	textures := make(map[string][]byte)

	for filename, imgTex := range texturesToFetch {
		log.Infof("RenderSetup: Fetching texture '%s' (expected size: %d bytes)...", filename, imgTex.GetSize())
		texData, err := s.streamTextureFile(ctx, transportClient, filename, imgTex.GetSize())
		if err != nil {
			errMsg := fmt.Sprintf("Failed to load texture '%s': %v", filename, err)
			s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
			return status.Error(codes.Internal, errMsg)
		}

		textures[filename] = texData

		log.Infof("RenderSetup: Successfully loaded texture '%s'. Actual size: %d bytes.", filename, len(texData))
	}

	log.Infof("RenderSetup: Finished streaming %d unique textures.", len(textures))

	// Step 3: Transform the scene to its internal representation
	cameraAspectRatio := float64(req.GetImageResolution().GetWidth()) / float64(req.GetImageResolution().GetHeight())
	t := transport.NewTransport(cameraAspectRatio, protoScene, triangles, textures)

	scene, err := t.ToScene()
	if err != nil {
		errMsg := fmt.Sprintf("Failed to convert scene to internal representation: %v", err)
		s.sendStatus(stream, pb_control.RenderSetupStatus_FAILED, errMsg)
		return status.Error(codes.Internal, errMsg)
	}

	s.scene = scene

	// Step 4: Setup render parameters
	s.maxDepth = int(req.GetMaxDepth())
	s.background = &vec3.Vec3Impl{X: req.GetBackgroundColor().GetX(), Y: req.GetBackgroundColor().GetY(), Z: req.GetBackgroundColor().GetZ()}
	s.ink = &vec3.Vec3Impl{X: req.GetInkColor().GetX(), Y: req.GetInkColor().GetY(), Z: req.GetInkColor().GetZ()}
	s.samplesPerPixel = int(req.GetSamplesPerPixel())
	s.imageResolutionX = int(req.GetImageResolution().GetWidth())
	s.imageResolutionY = int(req.GetImageResolution().GetHeight())

	switch req.GetSampler() {
	case pb_control.SamplerType_ALBEDO:
		s.sampler = sampler.NewAlbedo(&s.numRays)
	case pb_control.SamplerType_COLOUR:
		s.sampler = sampler.NewColour(s.maxDepth, s.background, &s.numRays)
	case pb_control.SamplerType_NORMAL:
		s.sampler = sampler.NewNormal(&s.numRays)
	case pb_control.SamplerType_WIRE_FRAME:
		s.sampler = sampler.NewWireFrame(s.background, s.ink, &s.numRays)
	default:
		return status.Errorf(codes.InvalidArgument, "invalid sampler type: %s", req.GetSampler().String())
	}

	log.Debugf("Render parameters: Max depth: %d, Background: %v, Ink: %v, Sampler: %s", s.maxDepth, s.background, s.ink, req.GetSampler().String())

	// Step 5: Send READY status
	if err := s.sendStatus(stream, pb_control.RenderSetupStatus_READY, ""); err != nil {
		return status.Errorf(codes.Internal, "failed to send READY status: %v", err)
	}
	log.Infof("RenderSetup: Worker is READY for rendering with scene '%s', %d triangles, and %d textures.",
		req.GetSceneName(), len(protoScene.GetObjects().GetTriangles()), len(textures))

	s.currentStatus = pb_discovery.WorkerStatus_BUSY_RENDERING

	return nil // Successfully configured
}
