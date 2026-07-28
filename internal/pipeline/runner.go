package pipeline

import "mc2lua/internal/model"

type worldReader interface {
	Run(input string, bounds model.Bounds) (*model.World, error)
}

type Runner struct {
	worldReader worldReader
}

func NewRunner(wr worldReader) *Runner {
	return &Runner{worldReader: wr}
}

type RunConfig struct {
	Input    string
	Output   string
	Scale    int
	NoOffset bool
	Bounds   model.Bounds
}

func (svc *Runner) Run(cfg RunConfig) error {
	_, err := svc.worldReader.Run(cfg.Input, cfg.Bounds)
	return err
}
