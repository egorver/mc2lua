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

type voxelIndexer interface {
	Run(blocks []model.RawBlock, styles index.StyleIndex) *index.VoxelIndex
}

type regionMerger interface {
	Run(grid *index.VoxelIndex) []model.Cuboid
}

type Runner struct {
	regionReader    regionReader
	coordNormalizer coordNormalizer
	assetScanner    assetScanner
	blockCollector  blockCollector
	styleIndexer    styleIndexer
	voxelIndexer    voxelIndexer
	regionMerger    regionMerger
	logOutput       io.Writer
}

func NewRunner(
	rr regionReader,
	cn coordNormalizer,
	as assetScanner,
	bc blockCollector,
	si styleIndexer,
	vi voxelIndexer,
	rm regionMerger,
	logOutput io.Writer,
) *Runner {
	return &Runner{
		regionReader:    rr,
		coordNormalizer: cn,
		assetScanner:    as,
		blockCollector:  bc,
		styleIndexer:    si,
		voxelIndexer:    vi,
		regionMerger:    rm,
		logOutput:       logOutput,
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
	if err != nil {
		return fmt.Errorf("read world: %w", err)
	}
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

	voxelIdx := svc.voxelIndexer.Run(blocks, *styledIdx)
	svc.log("Built voxel index: %d blocks\n", len(voxelIdx.Blocks()))

	regions := svc.regionMerger.Run(voxelIdx)
	svc.logMergeStats(len(voxelIdx.Blocks()), len(regions))

	return nil
}

func (svc *Runner) log(format string, args ...any) {
	fmt.Fprintf(svc.logOutput, format, args...)
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
