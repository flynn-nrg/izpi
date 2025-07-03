package worker

import (
	"context"

	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
)

func (s *workerServer) QueryWorkerStatus(ctx context.Context, req *pb_discovery.QueryWorkerStatusRequest) (*pb_discovery.QueryWorkerStatusResponse, error) {
	return &pb_discovery.QueryWorkerStatusResponse{
		NodeName:         s.workerID,
		AvailableCores:   s.availableCores,
		TotalMemoryBytes: s.totalMemoryBytes,
		FreeMemoryBytes:  s.freeMemoryBytes,
		Status:           s.currentStatus,
	}, nil
}
