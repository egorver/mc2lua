package index

import "mc2lua/internal/model"

type BlockLookup interface {
	Get(id, props string) (*model.ResolvedBlock, bool)
}

type BlockIndex struct {
	entries map[string]*model.ResolvedBlock
}

func NewBlockIndex() *BlockIndex {
	return &BlockIndex{entries: make(map[string]*model.ResolvedBlock)}
}

func (i *BlockIndex) Add(id, props string, resolved *model.ResolvedBlock) {
	i.entries[id+"|"+props] = resolved
}

func (i *BlockIndex) Get(id, props string) (*model.ResolvedBlock, bool) {
	key := id + "|" + props
	v, ok := i.entries[key]
	return v, ok
}
