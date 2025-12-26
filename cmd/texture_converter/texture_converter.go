package main

import (
	"github.com/flynn-nrg/go-vfx/go-oiio/oiio"

	"github.com/alecthomas/kong"
	log "github.com/sirupsen/logrus"
)

const (
	programName       = "texture_converter"
	dataProfileAlbedo = "albedo"
	dataProfileLinear = "linear"
	dataPrfileRaw     = "raw"
	defaultOutputFile = "output.exr"
)

var flags struct {
	DataProfile string `name:"data-profile" help:"Data profile: albedo, linear or raw" default:"${dataProfileAlbedo}"`
	InputFile   string `name:"input-file" help:"Input file" required:"true"`
	OutputFile  string `name:"output-file" help:"Output file" default:"${defaultOutputFile}"`
}

func main() {
	kong.Parse(&flags,
		kong.Name(programName),
		kong.Description("A tool to convert textures between different data profiles"),
		kong.Vars{
			"dataProfileAlbedo": dataProfileAlbedo,
			"defaultOutputFile": defaultOutputFile,
		},
	)

	log.Infof("Converting texture from %s to %s with data profile %s", flags.InputFile, flags.OutputFile, flags.DataProfile)

	switch flags.DataProfile {
	case dataProfileAlbedo:
		img, err := oiio.ReadImage(flags.InputFile, oiio.ConvertToACEScg)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}

		err = oiio.WriteImage(flags.OutputFile, img)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
	case dataProfileLinear:
		img, err := oiio.ReadImage(flags.InputFile, oiio.LineariseSRGB)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}

		err = oiio.WriteImage(flags.OutputFile, img)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
	case dataPrfileRaw:
		img, err := oiio.ReadImage(flags.InputFile, oiio.Raw)
		if err != nil {
			log.Fatalf("Failed to read input file: %v", err)
		}

		err = oiio.WriteImage(flags.OutputFile, img)
		if err != nil {
			log.Fatalf("Failed to write output file: %v", err)
		}
	}
}
