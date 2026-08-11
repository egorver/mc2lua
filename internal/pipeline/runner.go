package pipeline

import (
	"fmt"
	"io"
	"math"
	"sort"
	"strings"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

const percentMultiplier = 100.0

type regionReader interface {
	Run(input string, bounds model.Bounds) ([]model.RawBlock, error)
}

type boundsResolver interface {
	Run(blocks []model.RawBlock) model.Bounds
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
	Run(blocks []model.ResolvedBlock, namespaces map[string][]string) *stateful.StyleIndex
}

type blockVoxelIndexer interface {
	Run(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex
}

type microVoxelIndexer interface {
	Run(blocks []model.RawBlock, styles stateful.StyleIndex) *stateful.VoxelIndex
}

type regionMerger interface {
	Run(grid *stateful.VoxelIndex) []model.Cuboid
}

type occupancyIndexer interface {
	Run(blocks []model.RawBlock, blockIdx, microIdx *stateful.VoxelIndex, styles stateful.StyleIndex) *stateful.OccupancyIndex
}

type faceCuller interface {
	Run(occ *stateful.OccupancyIndex, blockRegions, microRegions []model.Cuboid, blocks []model.RawBlock, styles stateful.StyleIndex) model.FaceVisibility
}

type partBuilder interface {
	Run(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, visibility model.FaceVisibility, styleIndex stateful.StyleIndex, scale float64) ([]model.Part, error)
}

type templateGenerator interface {
	Run(parts []model.Part, blocks []model.RawBlock, outputPath string) error
}

type partStylizer interface {
	Run(parts []model.Part, styleIndex stateful.StyleIndex) []model.Part
}

type luaGenerator interface {
	Run(parts []model.Part, scale float64, outputPath string) error
}

type Runner struct {
	regionReader      regionReader
	boundsResolver    boundsResolver
	coordNormalizer   coordNormalizer
	assetScanner      assetScanner
	blockCollector    blockCollector
	styleIndexer      styleIndexer
	blockVoxelIndexer blockVoxelIndexer
	microVoxelIndexer microVoxelIndexer
	regionMerger      regionMerger
	occupancyIndexer  occupancyIndexer
	faceCuller        faceCuller
	partBuilder       partBuilder
	templateGenerator templateGenerator
	partStylizer      partStylizer
	luaGenerator      luaGenerator
	logOutput         io.Writer
}

func NewRunner(
	rr regionReader,
	br boundsResolver,
	cn coordNormalizer,
	as assetScanner,
	bc blockCollector,
	si styleIndexer,
	bvi blockVoxelIndexer,
	mvi microVoxelIndexer,
	rm regionMerger,
	oi occupancyIndexer,
	fc faceCuller,
	pb partBuilder,
	tg templateGenerator,
	ps partStylizer,
	lg luaGenerator,
	logOutput io.Writer,
) *Runner {
	return &Runner{
		regionReader:      rr,
		boundsResolver:    br,
		coordNormalizer:   cn,
		assetScanner:      as,
		blockCollector:    bc,
		styleIndexer:      si,
		blockVoxelIndexer: bvi,
		microVoxelIndexer: mvi,
		regionMerger:      rm,
		occupancyIndexer:  oi,
		faceCuller:        fc,
		partBuilder:       pb,
		templateGenerator: tg,
		partStylizer:      ps,
		luaGenerator:      lg,
		logOutput:         logOutput,
	}
}

type RunConfig struct {
	Input         string
	AssetsDir     string
	Output        string
	PartsTemplate string
	Scale         int
	NoOffset      bool
	Bounds        model.Bounds
}

func (svc *Runner) Run(cfg RunConfig) error {
	blocks, err := svc.regionReader.Run(cfg.Input, cfg.Bounds)
	svc.logRegionErrors(err)
	svc.log("Read %d blocks\n", len(blocks))

	actualBounds := svc.boundsResolver.Run(blocks)
	svc.logArea(cfg.Bounds, actualBounds)

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

	styledIdx := svc.styleIndexer.Run(resolved, namespaces)
	svc.log("Built style index: %d entries\n", styledIdx.Len())

	blockIdx := svc.blockVoxelIndexer.Run(blocks, *styledIdx)
	svc.log("Built block voxel index: %d blocks\n", len(blockIdx.Blocks()))

	blockRegions := svc.regionMerger.Run(blockIdx)
	svc.logMergeStats(len(blockIdx.Blocks()), len(blockRegions))

	microIdx := svc.microVoxelIndexer.Run(blocks, *styledIdx)
	svc.log("Built micro voxel index: %d blocks\n", len(microIdx.Blocks()))

	microRegions := svc.regionMerger.Run(microIdx)
	svc.logMergeStats(len(microIdx.Blocks()), len(microRegions))

	occIdx := svc.occupancyIndexer.Run(blocks, blockIdx, microIdx, *styledIdx)
	svc.log("Built occupancy index: %d cell(s)\n", occIdx.Len())

	visibility := svc.faceCuller.Run(occIdx, blockRegions, microRegions, blocks, *styledIdx)
	svc.logFaceVisibility(visibility)

	parts, err := svc.partBuilder.Run(blocks, blockRegions, microRegions, visibility, *styledIdx, float64(cfg.Scale))
	if err != nil {
		return fmt.Errorf("build parts: %w", err)
	}
	svc.log("Built %d part(s)\n", len(parts))

	if cfg.PartsTemplate != "" {
		if err := svc.templateGenerator.Run(parts, blocks, cfg.PartsTemplate); err != nil {
			return fmt.Errorf("generate parts template: %w", err)
		}
		svc.log("Parts template generated at %s\n", cfg.PartsTemplate)
	}

	parts = svc.partStylizer.Run(parts, *styledIdx)
	svc.logStyling(parts)

	if err := svc.luaGenerator.Run(parts, float64(cfg.Scale), cfg.Output); err != nil {
		return fmt.Errorf("generate lua: %w", err)
	}
	svc.log("Lua script generated at %s\n", cfg.Output)

	return nil
}

func (svc *Runner) log(format string, args ...any) {
	fmt.Fprintf(svc.logOutput, format, args...)
}

func (svc *Runner) logArea(requested, actual model.Bounds) {
	if requested.XMin != math.MinInt32 || requested.XMax != math.MaxInt32 ||
		requested.YMin != math.MinInt32 || requested.YMax != math.MaxInt32 ||
		requested.ZMin != math.MinInt32 || requested.ZMax != math.MaxInt32 {
		svc.log("Requested bounds: x [%d..%d], y [%d..%d], z [%d..%d]\n",
			requested.XMin, requested.XMax, requested.YMin, requested.YMax, requested.ZMin, requested.ZMax)
	}
	svc.log("Actual bounds: x [%d..%d], y [%d..%d], z [%d..%d]\n",
		actual.XMin, actual.XMax, actual.YMin, actual.YMax, actual.ZMin, actual.ZMax)
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
		svc.log(" (%.1f%% of original blocks)\n", float64(regionCount)/float64(voxelCount)*percentMultiplier)
	} else {
		svc.log("\n")
	}
}

func (svc *Runner) logFaceVisibility(v model.FaceVisibility) {
	svc.log("Computed face visibility: %d block region(s), %d micro region(s), %d complex block(s)\n",
		len(v.BlockFaces), len(v.MicroFaces), len(v.ComplexFaces))
}

func (svc *Runner) logStyling(parts []model.Part) {
	styledParts := 0
	styledFaces := 0
	for _, p := range parts {
		faces := svc.styledFaceCount(p)
		styledFaces += faces
		if faces > 0 || p.Transparency != nil {
			styledParts++
		}
	}
	svc.log("Styled %d part(s), %d face(s) in total\n", styledParts, styledFaces)
}

func (svc *Runner) styledFaceCount(p model.Part) int {
	count := 0
	for _, face := range []*model.Surface{p.Top, p.Bottom, p.Front, p.Back, p.Left, p.Right} {
		if face != nil {
			count++
		}
	}
	return count
}
