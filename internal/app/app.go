package app

import (
	"os"

	"mc2lua/internal/minecraft"
	"mc2lua/internal/model"
	"mc2lua/internal/pipeline"
	"mc2lua/internal/runtime"
)

type App struct{}

func New() *App {
	return &App{}
}

func (svc *App) Run(cfg AppConfig) error {
	runCfg := buildConfig(cfg)
	runner := buildDeps()
	return runner.Run(runCfg)
}

func buildConfig(cfg AppConfig) pipeline.RunConfig {
	return pipeline.RunConfig{
		Input:     cfg.Input,
		AssetsDir: cfg.AssetsDir,
		Output:    cfg.Output,
		Scale:     cfg.Scale,
		NoOffset:  cfg.NoOffset,
		Bounds: model.Bounds{
			XMin: cfg.XMin, XMax: cfg.XMax,
			YMin: cfg.YMin, YMax: cfg.YMax,
			ZMin: cfg.ZMin, ZMax: cfg.ZMax,
		},
	}
}

func buildDeps() *pipeline.Runner {
	fs := runtime.NewFS()

	chunkDecoder := minecraft.NewChunkDecoder()
	assetScanner := minecraft.NewAssetScanner(fs)
	propsKeyBuilder := minecraft.NewPropsKeyBuilder()
	blockstateParser := minecraft.NewBlockstateParser(fs, propsKeyBuilder)
	modelAnalyzer := minecraft.NewModelAnalyzer()
	modelParser := minecraft.NewModelParser(fs)
	textureResolver := minecraft.NewTextureResolver()
	blockResolver := minecraft.NewBlockResolver(blockstateParser, modelParser, textureResolver)
	elementRotator := minecraft.NewElementRotator()

	regionReader := pipeline.NewRegionReader(fs, chunkDecoder)
	coordNormalizer := pipeline.NewCoordNormalizer()
	blockCollector := pipeline.NewBlockCollector(blockResolver, propsKeyBuilder)
	styleIndexer := pipeline.NewStyleIndexer(modelAnalyzer, elementRotator)
	voxelIndexer := pipeline.NewVoxelIndexer(propsKeyBuilder)
	regionMerger := pipeline.NewRegionMerger()
	luaGenerator := pipeline.NewLuaGenerator(propsKeyBuilder)

	return pipeline.NewRunner(
		regionReader, coordNormalizer, assetScanner, blockCollector, styleIndexer, voxelIndexer, regionMerger, luaGenerator, os.Stdout)
}
