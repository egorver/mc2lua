package app

import (
	"fmt"
	"os"
	"path/filepath"

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

	runner, err := buildDeps(cfg.ConfigDir)
	if err != nil {
		return fmt.Errorf("build runner: %w", err)
	}

	if err := runner.Run(runCfg); err != nil {
		return fmt.Errorf("run pipeline: %w", err)
	}

	return nil
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

func buildDeps(configDir string) (*pipeline.Runner, error) {
	fs := runtime.NewFS()

	chunkDecoder := minecraft.NewChunkDecoder()
	assetScanner := minecraft.NewAssetScanner(fs)
	propsKeyBuilder := minecraft.NewPropsKeyBuilder()
	blockstateParser := minecraft.NewBlockstateParser(fs, propsKeyBuilder)
	gridAnalyzer := minecraft.NewGridAnalyzer()
	modelParser := minecraft.NewModelParser(fs)
	textureResolver := minecraft.NewTextureResolver()
	elementRotator := minecraft.NewElementRotator()
	blockResolver := minecraft.NewBlockResolver(blockstateParser, modelParser, textureResolver, elementRotator)
	colorExtractor := minecraft.NewColorExtractor(fs)

	materialMatcher, err := pipeline.NewMaterialMatcher(fs, filepath.Join(configDir, "materials.yaml"))
	if err != nil {
		return nil, fmt.Errorf("create material matcher: %w", err)
	}

	brightnessMatcher, err := pipeline.NewBrightnessMatcher(fs, filepath.Join(configDir, "materials.yaml"))
	if err != nil {
		return nil, fmt.Errorf("create brightness matcher: %w", err)
	}

	colorMatcher, err := pipeline.NewColorMatcher(fs, filepath.Join(configDir, "colors.yaml"))
	if err != nil {
		return nil, fmt.Errorf("create color matcher: %w", err)
	}

	regionReader := pipeline.NewRegionReader(fs, chunkDecoder)
	coordNormalizer := pipeline.NewCoordNormalizer()
	blockCollector := pipeline.NewBlockCollector(blockResolver, propsKeyBuilder)
	styleIndexer := pipeline.NewStyleIndexer(
		gridAnalyzer, elementRotator, materialMatcher, brightnessMatcher, colorExtractor, colorMatcher)
	blockVoxelIndexer := pipeline.NewBlockVoxelIndexer(propsKeyBuilder)
	microVoxelIndexer := pipeline.NewMicroVoxelIndexer(propsKeyBuilder)
	regionMerger := pipeline.NewRegionMerger()
	luaGenerator := pipeline.NewLuaGenerator(fs, propsKeyBuilder)

	return pipeline.NewRunner(
		regionReader,
		coordNormalizer,
		assetScanner,
		blockCollector,
		styleIndexer,
		blockVoxelIndexer,
		microVoxelIndexer,
		regionMerger,
		luaGenerator,
		os.Stdout), nil
}
