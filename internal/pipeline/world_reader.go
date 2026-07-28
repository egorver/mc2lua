package pipeline

import "mc2lua/internal/model"

type WorldReader struct {
	fs fsApi
}

func NewWorldReader(fs fsApi) *WorldReader {
	return &WorldReader{fs: fs}
}

func (svc *WorldReader) Run(input string, bounds model.Bounds) (*model.World, error) {
	return nil, nil
}
