package main

import (
	"flag"
	"fmt"
	"io"
	"math"
	"os"

	"mc2lua/internal/app"
)

const (
	defaultInputDir      = "region"
	defaultAssetsDir     = "assets"
	defaultOutputFile    = "output.lua"
	defaultPartsTemplate = "output/template.yaml"
	defaultScale         = 4
	defaultConfigDir     = "config"
)

// appRunner is a minimal interface for running the app, introduced
// solely to make main.go testable.
type appRunner interface {
	Run(cfg app.AppConfig) error
}

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, func() appRunner { return app.New() }))
}

func run(args []string, stdout, stderr io.Writer, newRunner func() appRunner) int {
	fs := flag.NewFlagSet("mc2lua", flag.ContinueOnError)
	fs.SetOutput(stderr)

	inputPath := fs.String("input", defaultInputDir, "path to region files directory")
	assetsDir := fs.String("assets", defaultAssetsDir, "path to Minecraft assets directory")
	outputPath := fs.String("output", defaultOutputFile, "output Lua file")
	partsTemplate := fs.String("parts-template", defaultPartsTemplate, "path to write template parts.yaml for new block types")
	scale := fs.Int("scale", defaultScale, "block scale factor")
	configDir := fs.String("config", defaultConfigDir, "path to configs directory")
	noOffset := fs.Bool("no-offset", false, "disable auto-offset to y=0")
	xmin := fs.Int("xmin", math.MinInt32, "minimum X")
	xmax := fs.Int("xmax", math.MaxInt32, "maximum X")
	ymin := fs.Int("ymin", math.MinInt32, "minimum Y")
	ymax := fs.Int("ymax", math.MaxInt32, "maximum Y")
	zmin := fs.Int("zmin", math.MinInt32, "minimum Z")
	zmax := fs.Int("zmax", math.MaxInt32, "maximum Z")
	help := fs.Bool("help", false, "show help")
	fs.BoolVar(help, "h", false, "show help")

	if err := fs.Parse(args); err != nil {
		// flag.ContinueOnError: parse errors are already printed to stderr.
		// flag.ErrHelp is returned for -h/-help with ContinueOnError.
		if err == flag.ErrHelp {
			return 0
		}
		return 2
	}

	if *help {
		fs.Usage()
		return 0
	}

	cfg := app.AppConfig{
		Input:         *inputPath,
		AssetsDir:     *assetsDir,
		Output:        *outputPath,
		PartsTemplate: *partsTemplate,
		Scale:         *scale,
		NoOffset:      *noOffset,
		ConfigDir:     *configDir,
		XMin:          *xmin,
		XMax:          *xmax,
		YMin:          *ymin,
		YMax:          *ymax,
		ZMin:          *zmin,
		ZMax:          *zmax,
	}

	if err := newRunner().Run(cfg); err != nil {
		fmt.Fprintf(stderr, "Error: %v\n", err)
		return 1
	}
	return 0
}
