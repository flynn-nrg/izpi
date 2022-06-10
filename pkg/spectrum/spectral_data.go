package spectrum

const (
	nRGB2SpectSamples = 32
	nCIESamples       = 471
)

var (
	rgb2SpectLambda      = []float64{}
	rgbRefl2SpectWhite   = []float64{}
	rgbRefl2SpectCyan    = []float64{}
	rgbRefl2SpectMagenta = []float64{}
	rgbRefl2SpectYellow  = []float64{}
	rgbRefl2SpectRed     = []float64{}
	rgbRefl2SpectGreen   = []float64{}
	rgbRefl2SpectBlue    = []float64{}

	rgbIllum2SpectWhite   = []float64{}
	rgbIllum2SpectCyan    = []float64{}
	rgbIllum2SpectMagenta = []float64{}
	rgbIllum2SpectYellow  = []float64{}
	rgbIllum2SpectRed     = []float64{}
	rgbIllum2SpectGreen   = []float64{}
	rgbIllum2SpectBlue    = []float64{}

	cieX      = []float64{}
	cieY      = []float64{}
	cieZ      = []float64{}
	cieLambda = []float64{}
)
