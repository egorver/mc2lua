package model

type StyleIndex struct {
	entries map[string]StyledBlock
}

func NewStyleIndex() *StyleIndex {
	return &StyleIndex{entries: make(map[string]StyledBlock)}
}

func (i *StyleIndex) Add(id, props string, block StyledBlock) {
	i.entries[id+"|"+props] = block
}

func (i *StyleIndex) Get(id, props string) (StyledBlock, bool) {
	v, ok := i.entries[id+"|"+props]
	return v, ok
}

func (i *StyleIndex) Len() int {
	return len(i.entries)
}
