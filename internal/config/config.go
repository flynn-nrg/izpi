package config

type Config struct {
	LogLevel         string `name:"log-level" help:"The log level: error, warn, info, debug, trace." default:"info"`
	Scene            string `type:"existingfile" name:"scene" help:"Scene file to render" default:"${defaultSceneFile}"`
	NumWorkers       int64  `name:"num-workers" help:"Number of worker threads" default:"${defaultNumWorkers}"`
	XSize            int64  `name:"x" help:"Output image x size" default:"${defaultXSize}"`
	YSize            int64  `name:"y" help:"Output image y size" default:"${defaultYSize}"`
	Samples          int64  `name:"samples" help:"Number of samples per ray" default:"${defaultSamples}"`
	Sampler          string `name:"sampler-type" help:"Sampler function to use: colour, albedo, normal, wireframe" default:"colour"`
	Depth            int64  `name:"max-depth" help:"Maximum depth" default:"${defaultMaxDepth}"`
	OutputMode       string `name:"output-mode" help:"Output mode: png, exr, hdr or pfm" default:"png"`
	OutputFile       string `type:"file" name:"output-file" help:"Output file." default:"${defaultOutputFile}"`
	Verbose          bool   `name:"v" help:"Print rendering progress bar" default:"true"`
	Preview          bool   `name:"p" help:"Display rendering progress in a window" default:"true"`
	DisplayMode      string `name:"display-mode" help:"Display mode: fyne or sdl" default:"fyne"`
	CpuProfile       string `name:"cpu-profile" help:"Enable cpu profiling"`
	Instrument       bool   `name:"instrument" help:"Enable instrumentation" default:"false"`
	Role             string `name:"role" help:"Role: worker, leader or standalone" default:"standalone"`
	DiscoveryTimeout int64  `name:"discovery-timeout" help:"Discovery timeout in seconds" default:"${defaultDiscoveryTimeout}"`
}
