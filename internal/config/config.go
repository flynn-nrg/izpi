package config

type Config struct {
	Scene            string
	NumWorkers       int64
	XSize            int64
	YSize            int64
	Samples          int64
	Sampler          string
	Depth            int64
	OutputMode       string
	OutputFile       string
	Verbose          bool
	Preview          bool
	DisplayMode      string
	DiscoveryTimeout int64
}
