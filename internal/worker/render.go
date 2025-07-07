package worker

import (
	"context"
	"fmt"
	"runtime"

	"github.com/flynn-nrg/izpi/internal/fastrandom"
	pb_control "github.com/flynn-nrg/izpi/internal/proto/control"
	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
	"github.com/flynn-nrg/izpi/internal/vec3"
	"github.com/pbnjay/memory"
	log "github.com/sirupsen/logrus"
)

func (s *workerServer) RenderTile(req *pb_control.RenderTileRequest, stream pb_control.RenderControlService_RenderTileServer) error {
	x0 := req.GetX0()
	y0 := req.GetY0()
	x1 := req.GetX1()
	y1 := req.GetY1()

	rand := s.randPool.Get().(*fastrandom.LCG)
	defer s.randPool.Put(rand)

	stripSize := req.GetStripHeight() * 4 * (x1 - x0 + 1)

	responseWidth := x1 - x0 + 1

	nx := float64(s.imageResolutionX)
	ny := float64(s.imageResolutionY)

	for y := y0; y <= y1; y++ {
		pixels := make([]float64, stripSize)

		i := 0
		for x := x0; x <= x1; x++ {
			select {
			case <-stream.Context().Done():
				log.Warnf("RenderTile stream cancelled for tile [%d,%d]: %v", req.GetX0(), req.GetY0(), stream.Context().Err())
				return stream.Context().Err()
			default:
				col := &vec3.Vec3Impl{}
				for sample := 0; sample < s.samplesPerPixel; sample++ {
					u := (float64(x) + rand.Float64()) / nx
					v := (float64(y) + rand.Float64()) / ny
					r := s.scene.Camera.GetRay(u, v)
					col = vec3.Add(col, vec3.DeNAN(s.sampler.Sample(r, s.scene.World, s.scene.Lights, 0, rand)))
				}

				col = vec3.ScalarDiv(col, float64(s.samplesPerPixel))
				pixels[i] = col.Z
				pixels[i+1] = col.Y
				pixels[i+2] = col.X
				pixels[i+3] = 1.0
				i += 4
			}

		}

		resp := &pb_control.RenderTileResponse{
			Width:  responseWidth,
			Height: 1,
			PosX:   x0,
			PosY:   y,
			Pixels: pixels,
		}

		if err := stream.Send(resp); err != nil {
			log.Errorf("Failed to send RenderTileResponse chunk for tile [%d,%d]: %v", req.GetX0(), req.GetY0(), err)
			return fmt.Errorf("failed to send stream chunk: %w", err)
		}
	}

	return nil
}

func (s *workerServer) RenderEnd(ctx context.Context, req *pb_control.RenderEndRequest) (*pb_control.RenderEndResponse, error) {
	s.currentStatus = pb_discovery.WorkerStatus_FREE

	log.Infof("RenderControlService: RenderEnd called by %s", s.workerID)
	log.Infof("Total rays traced: %d", s.numRays)

	stats := &pb_control.RenderEndResponse{
		TotalRaysTraced: s.numRays,
	}

	memBefore := memory.FreeMemory()

	// Free up resources
	s.scene = nil
	s.sampler = nil
	s.currentStatus = pb_discovery.WorkerStatus_FREE
	s.numRays = 0
	s.imageResolutionX = 0
	s.imageResolutionY = 0
	s.background = nil
	s.ink = nil
	s.samplesPerPixel = 0
	s.maxDepth = 0

	// Hint GC to collect any remaining resources
	runtime.GC()

	memAfter := memory.FreeMemory()

	log.Infof("Freed %d bytes of memory", memAfter-memBefore)
	log.Infof("Server is now in free state")

	// Refresh memory stats
	totalMem := memory.TotalMemory()
	freeMem := memory.FreeMemory()
	s.totalMemoryBytes = totalMem
	s.freeMemoryBytes = freeMem

	return stats, nil
}
