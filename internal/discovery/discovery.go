package discovery

import (
	"context"
	"fmt" // Import net for parsing IP addresses from Avahi
	"net"
	"runtime"
	"time"

	"github.com/godbus/dbus/v5" // D-Bus for Avahi
	"github.com/grandcat/zeroconf"
	"github.com/holoplot/go-avahi" // Avahi D-Bus client
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb_discovery "github.com/flynn-nrg/izpi/internal/proto/discovery"
)

const (
	serviceName = "_izpi-worker._tcp"
	domain      = "local."
)

// Discovery encapsulates the mDNS resolver logic, now OS-dependent.
type Discovery struct {
	// Either a zeroconf.Resolver or Avahi D-Bus connection/server, depending on OS
	resolver  interface{}
	timeout   time.Duration
	currentOS string // Store the OS for conditional logic
}

func New(timeout time.Duration) (*Discovery, error) {
	currentOS := runtime.GOOS

	log.Infof("Discovery: Initializing discovery for OS: %s", currentOS)

	d := &Discovery{
		timeout:   timeout,
		currentOS: currentOS,
	}

	switch currentOS {
	case "linux", "freebsd":
		// For Linux/FreeBSD, connect to D-Bus for Avahi
		conn, err := dbus.SystemBus()
		if err != nil {
			return nil, fmt.Errorf("failed to connect to D-Bus system bus for Avahi: %w", err)
		}
		server, err := avahi.ServerNew(conn)
		if err != nil {
			conn.Close()
			return nil, fmt.Errorf("failed to create Avahi server client: %w", err)
		}
		d.resolver = struct {
			conn   *dbus.Conn
			server *avahi.Server
		}{conn, server}
	case "darwin":
		// For MacOS, use grandcat/zeroconf's direct mDNS implementation
		resolver, err := zeroconf.NewResolver(nil)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize zeroconf resolver: %w", err)
		}
		d.resolver = resolver
	default:
		return nil, fmt.Errorf("unsupported operating system for mDNS discovery: %s", currentOS)
	}

	return d, nil
}

// Close ensures proper shutdown of the OS-specific resolver.
func (d *Discovery) Close() {
	switch d.currentOS {
	case "linux", "freebsd":
		if res, ok := d.resolver.(struct {
			conn   *dbus.Conn
			server *avahi.Server
		}); ok {
			log.Info("Discovery: Closing Avahi D-Bus connection and server...")
			res.server.Close()

			if err := res.conn.Close(); err != nil {
				log.Errorf("Error closing D-Bus connection: %v", err)
			}
		}
	case "darwin":
		// grandcat/zeroconf resolver doesn't have a specific Close method for Browse.
	}
}

func (d *Discovery) FindWorkers() (map[string]*pb_discovery.QueryWorkerStatusResponse, error) {
	workerHosts := make(map[string]*pb_discovery.QueryWorkerStatusResponse)

	switch d.currentOS {
	case "linux", "freebsd":
		res, ok := d.resolver.(struct {
			conn   *dbus.Conn
			server *avahi.Server
		})
		if !ok {
			return nil, fmt.Errorf("avahi resolver not initialized correctly")
		}
		server := res.server

		ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
		defer cancel()

		sb, err := server.ServiceBrowserNew(
			avahi.InterfaceUnspec,
			avahi.ProtoUnspec,
			serviceName,
			domain,
			0,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create Avahi service browser: %w", err)
		}

		defer func() {
			if sb != nil {
				server.ServiceBrowserFree(sb)
			}
		}()

		log.Infof("Leader: Browse for Avahi services type '%s' in domain '%s'", serviceName, domain)

		for {
			select {
			case addEvent := <-sb.AddChannel:
				log.Infof("Leader: Discovered Avahi service ADD: Name='%s', Type='%s', Domain='%s'", addEvent.Name, addEvent.Type, addEvent.Domain)

				resolvedService, err := server.ResolveService(
					addEvent.Interface,
					addEvent.Protocol,
					addEvent.Name,
					addEvent.Type,
					addEvent.Domain,
					avahi.ProtoUnspec,
					0,
				)
				if err != nil {
					log.Errorf("failed to resolve Avahi service %s: %v", addEvent.Name, err)
					continue
				}

				var ipToDial string
				if resolvedService.Address != "" {
					ip := net.ParseIP(resolvedService.Address)
					if ip != nil {
						if ip.To4() != nil {
							ipToDial = ip.String()
						} else {
							ipToDial = fmt.Sprintf("[%s]", ip.String())
						}
					}
				}

				if ipToDial == "" {
					log.Warnf("No valid IP address found after resolving service: %s", resolvedService.Name)
					continue
				}

				target := fmt.Sprintf("%s:%d", ipToDial, resolvedService.Port)

				// Convert Txt records to string
				txtRecords := make([]string, len(resolvedService.Txt))
				for i, txt := range resolvedService.Txt {
					txtRecords[i] = string(txt)
				}

				log.Infof("Leader: Resolved service: Host='%s', IP='%s', Port=%d, TXT=%v", resolvedService.Host, ipToDial, resolvedService.Port, txtRecords)

				statusResp, err := d.discoverWorker(resolvedService.Host, target)
				if err != nil {
					log.Errorf("failed to discover worker %s: %v", target, err)
					continue
				}

				if statusResp.GetStatus() == pb_discovery.WorkerStatus_FREE {
					workerHosts[target] = statusResp
				}
			case <-ctx.Done():
				log.Infof("Leader: Avahi Browse context done: %v", ctx.Err())
				return workerHosts, nil
			}
		}

	case "darwin":
		resolver, ok := d.resolver.(*zeroconf.Resolver)
		if !ok {
			return nil, fmt.Errorf("zeroconf resolver not initialized correctly")
		}

		entries := make(chan *zeroconf.ServiceEntry)
		ctx, cancel := context.WithTimeout(context.Background(), d.timeout)
		defer cancel()

		err := resolver.Browse(ctx, serviceName, domain, entries)
		if err != nil {
			log.Fatalf("failed to browse with zeroconf: %v", err)
		}

		for entry := range entries {
			log.Infof("Leader: Discovered zeroconf service: HostName='%s', Port=%d, Addrs=%v", entry.HostName, entry.Port, entry.AddrIPv4)

			var ipToDial string
			if len(entry.AddrIPv4) > 0 {
				ipToDial = entry.AddrIPv4[0].String()
			} else if len(entry.AddrIPv6) > 0 {
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

			if statusResp.GetStatus() == pb_discovery.WorkerStatus_FREE {
				workerHosts[target] = statusResp
			}
		}
	default:
		return nil, fmt.Errorf("unsupported operating system for mDNS discovery: %s", d.currentOS)
	}

	return workerHosts, nil
}

// discoverWorker remains the same
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
