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
	logOutput       io.Writer
}

func NewRunner(
	rr regionReader,
	cn coordNormalizer,
	as assetScanner,
	ib indexBuilder,
	logOutput io.Writer,
) *Runner {
	return &Runner{
		regionReader:    rr,
		coordNormalizer: cn,
		assetScanner:    as,
		indexBuilder:    ib,
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

	idx, unresolvedErrs, err := svc.indexBuilder.Run(blocks, namespaces)
	if err != nil {
		return fmt.Errorf("build index: %w", err)
	}
	svc.logUnresolved(unresolvedErrs)
	svc.log("Built index: %d unique block variant(s)\n", idx.Len())

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
