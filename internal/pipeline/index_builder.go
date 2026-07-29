package pipeline

import (
	"fmt"
	"slices"
	"strings"

	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type builderBlockResolver interface {
	Run(id, props string) (*model.ResolvedBlock, error)
}

type IndexBuilder struct {
	blockResolver builderBlockResolver
}

func NewIndexBuilder(br builderBlockResolver) *IndexBuilder {
	return &IndexBuilder{blockResolver: br}
}

func (svc *IndexBuilder) Run(blocks []model.Block) (*index.BlockIndex, error) {
	idx := index.NewBlockIndex()

	for _, b := range blocks {
		propsStr := svc.serializeProps(b.Props)
		if _, ok := idx.Get(b.ID, propsStr); ok {
			continue
		}

		resolved, err := svc.blockResolver.Run(b.ID, propsStr)
		if err != nil {
			return nil, fmt.Errorf("resolve block %s: %w", b.ID+"|"+propsStr, err)
		}
		idx.Add(b.ID, propsStr, resolved)
	}

	return idx, nil
}

func (svc *IndexBuilder) serializeProps(props map[string]string) string {
	keys := make([]string, 0, len(props))
	for k := range props {
		keys = append(keys, k)
	}
	slices.Sort(keys)

	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(k)
		b.WriteByte('=')
		b.WriteString(props[k])
	}
	return b.String()
}
