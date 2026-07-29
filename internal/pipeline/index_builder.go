package pipeline

import (
	"slices"
	"strings"

	"mc2lua/internal/index"
	"mc2lua/internal/model"
)

type blockResolver interface {
	Run(id string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
}

type IndexBuilder struct {
	blockResolver blockResolver
}

func NewIndexBuilder(br blockResolver) *IndexBuilder {
	return &IndexBuilder{blockResolver: br}
}

func (svc *IndexBuilder) Run(blocks []model.RawBlock, namespaces map[string][]string) (*index.BlockIndex, map[string]string, error) {
	idx := index.NewBlockIndex()
	var unresolvedErrs map[string]string

	for _, b := range blocks {
		propsKey := svc.propsToKey(b.Props)
		if _, ok := idx.Get(b.ID, propsKey); ok {
			continue
		}

		resolved, err := svc.blockResolver.Run(b.ID, b.Props, namespaces)
		if err != nil {
			if unresolvedErrs == nil {
				unresolvedErrs = make(map[string]string)
			}
			if _, exists := unresolvedErrs[b.ID]; !exists {
				unresolvedErrs[b.ID] = err.Error()
			}
			continue
		}
		idx.Add(b.ID, propsKey, resolved)
	}

	return idx, unresolvedErrs, nil
}

func (svc *IndexBuilder) propsToKey(props map[string]string) string {
	if len(props) == 0 {
		return ""
	}
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
