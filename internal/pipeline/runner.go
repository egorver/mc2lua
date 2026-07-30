package pipeline

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type regionReader interface {
	Run(input string, bounds model.Bounds) ([]model.RawBlock, error)
}

type coordNormalizer interface {
	Run(blocks []model.RawBlock, noOffset bool) ([]model.RawBlock, error)
}

type assetScanner interface {
	Run(root string) (map[string][]string, error)
}

type blockCollector interface {
	Run(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string)
}

type styleIndexer interface {
	Run(blocks []model.ResolvedBlock) *index.StyleIndex
}

type blockVoxelIndexer interface {
	Run(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex
}

type microVoxelIndexer interface {
	Run(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex
}

type regionMerger interface {
	Run(grid *index.VoxelIndex) []model.Cuboid
}

type luaGenerator interface {
	Run(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, styleIndex index.StyleIndex, scale float64, outputPath string) (int, error)
}

type Runner struct {
	regionReader      regionReader
	coordNormalizer   coordNormalizer
	assetScanner      assetScanner
	blockCollector    blockCollector
	styleIndexer      styleIndexer
	blockVoxelIndexer blockVoxelIndexer
	microVoxelIndexer microVoxelIndexer
	regionMerger      regionMerger
	luaGenerator      luaGenerator
	logOutput         io.Writer
}

func NewRunner(
	rr regionReader,
	cn coordNormalizer,
	as assetScanner,
	bc blockCollector,
	si styleIndexer,
	bvi blockVoxelIndexer,
	mvi microVoxelIndexer,
	rm regionMerger,
	lg luaGenerator,
	logOutput io.Writer,
) *Runner {
	return &Runner{
		regionReader:      rr,
		coordNormalizer:   cn,
		assetScanner:      as,
		blockCollector:    bc,
		styleIndexer:      si,
		blockVoxelIndexer: bvi,
		microVoxelIndexer: mvi,
		regionMerger:      rm,
		luaGenerator:      lg,
		logOutput:         logOutput,
	}
}

type RunConfig struct {
	Input     string
	AssetsDir string
	Output    string
	Scale     int
	NoOffset  bool
	Bounds    model.Bounds
}

func (svc *Runner) Run(cfg RunConfig) error {
	blocks, err := svc.regionReader.Run(cfg.Input, cfg.Bounds)
	svc.logRegionErrors(err)
	svc.log("Read %d blocks\n", len(blocks))

	blocks, err = svc.coordNormalizer.Run(blocks, cfg.NoOffset)
	if err != nil {
		return fmt.Errorf("normalize coords: %w", err)
	}
	svc.log("Adjusted coordinates for %d blocks\n", len(blocks))

	namespaces, err := svc.assetScanner.Run(cfg.AssetsDir)
	if err != nil {
		return fmt.Errorf("scan assets: %w", err)
	}
	svc.logNamespaces(namespaces)

	resolved, unresolved := svc.blockCollector.Run(blocks, namespaces)
	svc.logUnresolved(unresolved)
	svc.log("Collected %d unique block variant(s)\n", len(resolved))

	styledIdx := svc.styleIndexer.Run(resolved)
	svc.log("Built style index: %d entries\n", styledIdx.Len())

	blockIdx := svc.blockVoxelIndexer.Run(blocks, *styledIdx)
	svc.log("Built block voxel index: %d blocks\n", len(blockIdx.Blocks()))

	blockRegions := svc.regionMerger.Run(blockIdx)
	svc.logMergeStats(len(blockIdx.Blocks()), len(blockRegions))

	microIdx := svc.microVoxelIndexer.Run(blocks, *styledIdx)
	svc.log("Built micro voxel index: %d blocks\n", len(microIdx.Blocks()))

	microRegions := svc.regionMerger.Run(microIdx)
	svc.logMergeStats(len(microIdx.Blocks()), len(microRegions))

	totalParts, err := svc.luaGenerator.Run(blocks, blockRegions, microRegions, *styledIdx, float64(cfg.Scale), cfg.Output)
	if err != nil {
		return fmt.Errorf("generate lua: %w", err)
	}
	svc.log("Generated %d part(s)\n", totalParts)

	return nil
}

func (svc *Runner) log(format string, args ...any) {
	fmt.Fprintf(svc.logOutput, format, args...)
}

func (svc *Runner) logRegionErrors(err error) {
	if err == nil {
		return
	}
	svc.log("Warning: errors reading regions:\n")
	for _, line := range strings.Split(err.Error(), "\n") {
		svc.log("  %s\n", line)
	}
}

func (svc *Runner) logNamespaces(namespaces map[string][]string) {
	nsNames := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		nsNames = append(nsNames, ns)
	}
	sort.Strings(nsNames)
	svc.log("Found %d namespace(s): %s\n", len(namespaces), strings.Join(nsNames, ", "))
}

func (svc *Runner) logUnresolved(unresolved map[string]string) {
	ids := make([]string, 0, len(unresolved))
	for id := range unresolved {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		svc.log("Warning: skipping %s: %s\n", id, unresolved[id])
	}
}

func (svc *Runner) logMergeStats(voxelCount, regionCount int) {
	svc.log("Merged into %d region(s)", regionCount)
	if voxelCount > 0 {
		svc.log(" (%.1f%% of original blocks)\n", float64(regionCount)/float64(voxelCount)*100)
	} else {
		svc.log("\n")
	}
}
