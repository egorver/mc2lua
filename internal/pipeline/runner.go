package pipeline

import (
	"fmt"
	"sort"
	"strings"

	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type regionReader interface {
	Run(input string, bounds model.Bounds) ([]model.Block, error)
}

type coordNormalizer interface {
	Run(blocks []model.Block, noOffset bool) ([]model.Block, error)
}

type assetScanner interface {
	Run(root string) (map[string][]string, error)
}

type indexBuilder interface {
	Run(blocks []model.Block, namespaces map[string][]string) (*index.BlockIndex, map[string]string, error)
}

type Runner struct {
	regionReader    regionReader
	coordNormalizer coordNormalizer
	assetScanner    assetScanner
	indexBuilder    indexBuilder
}

func NewRunner(
	rr regionReader,
	cn coordNormalizer,
	as assetScanner,
	ib indexBuilder,
) *Runner {
	return &Runner{
		regionReader:    rr,
		coordNormalizer: cn,
		assetScanner:    as,
		indexBuilder:    ib,
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
	fmt.Printf("Read %d blocks\n", len(blocks))

	blocks, err = svc.coordNormalizer.Run(blocks, cfg.NoOffset)
	if err != nil {
		return fmt.Errorf("normalize coords: %w", err)
	}
	fmt.Printf("Adjusted coordinates for %d blocks\n", len(blocks))

	namespaces, err := svc.assetScanner.Run(cfg.AssetsDir)
	if err != nil {
		return fmt.Errorf("scan assets: %w", err)
	}
	nsNames := make([]string, 0, len(namespaces))
	for ns := range namespaces {
		nsNames = append(nsNames, ns)
	}
	fmt.Printf("Found %d namespace(s): %s\n", len(namespaces), strings.Join(nsNames, ", "))

	idx, unresolvedErrs, err := svc.indexBuilder.Run(blocks, namespaces)
	if err != nil {
		return fmt.Errorf("build index: %w", err)
	}
	ids := make([]string, 0, len(unresolvedErrs))
	for id := range unresolvedErrs {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		fmt.Printf("Warning: skipping %s: %s\n", id, unresolvedErrs[id])
	}
	fmt.Printf("Built index: %d unique block variant(s)\n", idx.Len())

	return nil
}
