package pipeline

import (
	"fmt"

	"mc2lua/internal/model"
)

type regionReader interface {
	Run(input string, bounds model.Bounds) ([]model.Block, error)
}

type Runner struct {
	regionReader regionReader
}

func NewRunner(rr regionReader) *Runner {
	return &Runner{regionReader: rr}
}

type RunConfig struct {
	Input    string
	Output   string
	Scale    int
	NoOffset bool
	Bounds   model.Bounds
}

func (svc *Runner) Run(cfg RunConfig) error {
	_, err := svc.regionReader.Run(cfg.Input, cfg.Bounds)
	if err != nil {
		return fmt.Errorf("read world: %w", err)
	}

	return nil
}
