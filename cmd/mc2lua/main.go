package main

import (
	"flag"
	"fmt"
	"math"
	"os"

	"mc2lua/internal/app"
)

const (
	defaultInputDir   = "region"
	defaultAssetsDir  = "assets"
	defaultOutputFile = "output.lua"
	defaultScale      = 4
	defaultConfigDir  = "config"
)

func main() {
	inputPath := flag.String("input", defaultInputDir, "path to region files directory")
	assetsDir := flag.String("assets", defaultAssetsDir, "path to Minecraft assets directory")
	outputPath := flag.String("output", defaultOutputFile, "output Lua file")
	scale := flag.Int("scale", defaultScale, "block scale factor")
	configDir := flag.String("config", defaultConfigDir, "path to configs directory")
	noOffset := flag.Bool("no-offset", false, "disable auto-offset to y=0")
	xmin := flag.Int("xmin", math.MinInt32, "minimum X")
	xmax := flag.Int("xmax", math.MaxInt32, "maximum X")
	ymin := flag.Int("ymin", math.MinInt32, "minimum Y")
	ymax := flag.Int("ymax", math.MaxInt32, "maximum Y")
	zmin := flag.Int("zmin", math.MinInt32, "minimum Z")
	zmax := flag.Int("zmax", math.MaxInt32, "maximum Z")
	help := flag.Bool("help", false, "show help")
	flag.BoolVar(help, "h", false, "show help")

	flag.Parse()

	if *help {
		flag.Usage()
		return
	}

	cfg := app.AppConfig{
		Input:     *inputPath,
		AssetsDir: *assetsDir,
		Output:    *outputPath,
		Scale:     *scale,
		NoOffset:  *noOffset,
		ConfigDir: *configDir,
		XMin:      *xmin,
		XMax:      *xmax,
		YMin:      *ymin,
		YMax:      *ymax,
		ZMin:      *zmin,
		ZMax:      *zmax,
	}

	if err := app.New().Run(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
