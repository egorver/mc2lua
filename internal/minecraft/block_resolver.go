package minecraft

import (
	"mc2lua/internal/model"
)

type BlockResolver struct{}

func NewBlockResolver() *BlockResolver {
	return &BlockResolver{}
}

func (svc *BlockResolver) Run(id, props string) (*model.ResolvedBlock, error) {
	return &model.ResolvedBlock{}, nil
}
