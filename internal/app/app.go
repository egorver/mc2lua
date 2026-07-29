package app

import (
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
		Input:    cfg.Input,
		Output:   cfg.Output,
		Scale:    cfg.Scale,
		NoOffset: cfg.NoOffset,
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
	blockResolver := minecraft.NewBlockResolver()

	regionReader := pipeline.NewRegionReader(fs, chunkDecoder)
	coordNormalizer := pipeline.NewCoordNormalizer()
	indexBuilder := pipeline.NewIndexBuilder(blockResolver)

	return pipeline.NewRunner(regionReader, coordNormalizer, indexBuilder)
}
