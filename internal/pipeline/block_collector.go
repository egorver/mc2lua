package pipeline

import (
	"slices"
	"strings"

	"mc2lua/internal/model"
)

type blockResolver interface {
	Run(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
}

type BlockCollector struct {
	blockResolver blockResolver
}

func NewBlockCollector(br blockResolver) *BlockCollector {
	return &BlockCollector{blockResolver: br}
}

func (svc *BlockCollector) Run(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string) {
	seen := make(map[string]bool)
	var result []model.ResolvedBlock
	var unresolved map[string]string

	for _, b := range blocks {
		propsKey := svc.propsToKey(b.Props)
		key := b.ID + "|" + propsKey
		if seen[key] {
			continue
		}
		seen[key] = true

		resolved, err := svc.blockResolver.Run(b.ID, propsKey, b.Props, namespaces)
		if err != nil {
			if unresolved == nil {
				unresolved = make(map[string]string)
			}
			if _, exists := unresolved[b.ID]; !exists {
				unresolved[b.ID] = err.Error()
			}
			continue
		}

		result = append(result, *resolved)
	}

	return result, unresolved
}

func (svc *BlockCollector) propsToKey(props map[string]string) string {
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
