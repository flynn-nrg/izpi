package discovery

import (
	"context"
	"fmt"
	"time"

	"github.com/grandcat/zeroconf"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
)

const (
	serviceName = "_izpi-worker._tcp"
	domain      = "local."
)

type Discovery struct {
	resolver *zeroconf.Resolver
	timeout  time.Duration
}

func New(timeout time.Duration) (*Discovery, error) {
	resolver, err := zeroconf.NewResolver(nil)
	if err != nil {
		return nil, err
	}

	return &Discovery{resolver: resolver, timeout: timeout}, nil
}

func (d *Discovery) FindWorkers() (map[string]*pb_discovery.QueryWorkerStatusResponse, error) {
	workerHosts := make(map[string]*pb_discovery.QueryWorkerStatusResponse)

	entries := make(chan *zeroconf.ServiceEntry)

	ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
	defer cancel()
	err := d.resolver.Browse(ctx, serviceName, domain, entries)
	if err != nil {
		log.Fatalln("failed to browse:", err.Error())
	}

	for entry := range entries {
		log.Infof("Discovered service: HostName='%s', Port=%d, Addrs=%v", entry.HostName, entry.Port, entry.AddrIPv4) // Add this line

		var ipToDial string
		// Prefer IPv4 from the discovered addresses
		if len(entry.AddrIPv4) > 0 {
			ipToDial = entry.AddrIPv4[0].String()
		} else if len(entry.AddrIPv6) > 0 {
			// For IPv6, ensure it's in brackets for net.Dial/grpc.NewClient
			ipToDial = fmt.Sprintf("[%s]", entry.AddrIPv6[0].String())
		} else {
			log.Warnf("No IP addresses found for service entry: %s", entry.HostName)
			continue
		}

		target := fmt.Sprintf("%s:%d", ipToDial, entry.Port)

		statusResp, err := d.discoverWorker(entry.HostName, target)
		if err != nil {
			log.Errorf("failed to discover worker %s: %v", target, err)
			continue
		}

		// Only add workers that are free
		if statusResp.GetStatus() == pb_discovery.WorkerStatus_FREE {
			workerHosts[target] = statusResp
		}
	}

	return workerHosts, nil
}

func (d *Discovery) discoverWorker(nodeName string, target string) (*pb_discovery.QueryWorkerStatusResponse, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Infof("Leader: Attempting gRPC dial to %s", nodeName)
	conn, err := grpc.NewClient(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Errorf("failed to connect to worker %s: %v", target, err)
		return nil, err
	}

	log.Infof("remote connection address: %s", conn.Target())
	defer conn.Close()

	discoveryClient := pb_discovery.NewWorkerDiscoveryServiceClient(conn)

	statusResp, err := discoveryClient.QueryWorkerStatus(ctx, &pb_discovery.QueryWorkerStatusRequest{})
	if err != nil {
		log.Errorf("failed to query status from worker %s: %v", nodeName, err)
		return nil, err
	}

	// Print the response from the worker
	log.Infof("--- Status from Worker %s ---", statusResp.GetNodeName())
	log.Infof("  Node Name: %s", statusResp.GetNodeName())
	log.Infof("  Endianness: %s", statusResp.GetEndianness().String())
	log.Infof("  Available Cores: %d", statusResp.GetAvailableCores())
	log.Infof("  Total Memory: %d MiB", statusResp.GetTotalMemoryBytes()/1024/1024)
	log.Infof("  Free Memory: %d MiB", statusResp.GetFreeMemoryBytes()/1024/1024)
	log.Infof("  Status: %s", statusResp.GetStatus().String())
	log.Info("--------------------------------------")

	return statusResp, nil
}
