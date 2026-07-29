package pipeline

import (
	"mc2lua/internal/model"
)

type blockResolver interface {
	Run(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error)
}

type propsKeyBuilder interface {
	Run(props map[string]string) string
}

type BlockCollector struct {
	blockResolver   blockResolver
	propsKeyBuilder propsKeyBuilder
}

func NewBlockCollector(br blockResolver, pkb propsKeyBuilder) *BlockCollector {
	return &BlockCollector{blockResolver: br, propsKeyBuilder: pkb}
}

func (svc *BlockCollector) Run(blocks []model.RawBlock, namespaces map[string][]string) ([]model.ResolvedBlock, map[string]string) {
	seen := make(map[string]bool)
	var result []model.ResolvedBlock
	var unresolved map[string]string

	for _, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
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
