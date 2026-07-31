package stateful

import "mc2lua/internal/model"

type StyleIndex struct {
	entries map[string]model.StyledBlock
}

func NewStyleIndex() *StyleIndex {
	return &StyleIndex{entries: make(map[string]model.StyledBlock)}
}

func (i *StyleIndex) Add(id, props string, block model.StyledBlock) {
	i.entries[id+"|"+props] = block
}

func (i *StyleIndex) Get(id, props string) (model.StyledBlock, bool) {
	v, ok := i.entries[id+"|"+props]
	return v, ok
}

func (i *StyleIndex) Len() int {
	return len(i.entries)
}
