package pipeline

import (
	"fmt"

	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type regionReader interface {
	Run(input string, bounds model.Bounds) ([]model.Block, error)
}

type coordNormalizer interface {
	Run(blocks []model.Block, noOffset bool) ([]model.Block, error)
}

type indexBuilder interface {
	Run(blocks []model.Block) (*index.BlockIndex, error)
}

type Runner struct {
	regionReader    regionReader
	coordNormalizer coordNormalizer
	indexBuilder    indexBuilder
}

func NewRunner(
	rr regionReader,
	cn coordNormalizer,
	ib indexBuilder,
) *Runner {
	return &Runner{
		regionReader:    rr,
		coordNormalizer: cn,
		indexBuilder:    ib,
	}
}

type RunConfig struct {
	Input    string
	Output   string
	Scale    int
	NoOffset bool
	Bounds   model.Bounds
}

func (svc *Runner) Run(cfg RunConfig) error {
	blocks, err := svc.regionReader.Run(cfg.Input, cfg.Bounds)
	if err != nil {
		return fmt.Errorf("read world: %w", err)
	}

	blocks, err = svc.coordNormalizer.Run(blocks, cfg.NoOffset)
	if err != nil {
		return fmt.Errorf("normalize coords: %w", err)
	}

	_, err = svc.indexBuilder.Run(blocks)
	if err != nil {
		return fmt.Errorf("build index: %w", err)
	}

	return nil
}
