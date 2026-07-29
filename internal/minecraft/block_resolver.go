package minecraft

import (
	"fmt"
	"mc2lua/internal/model"
	"strings"
)

type blockstateParser interface {
	Run(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error)
}

type modelAnalyzer interface {
	Run(elements []model.ModelElement) bool
}

type modelParser interface {
	Run(modelName string, namespaces map[string][]string) (*flattenedModel, error)
}

type BlockResolver struct {
	blockstateParser blockstateParser
	modelAnalyzer    modelAnalyzer
	modelParser      modelParser
}

func NewBlockResolver(
	blockstateParser blockstateParser,
	modelAnalyzer modelAnalyzer,
	modelParser modelParser,
) *BlockResolver {
	return &BlockResolver{
		blockstateParser: blockstateParser,
		modelAnalyzer:    modelAnalyzer,
		modelParser:      modelParser,
	}
}

func (svc *BlockResolver) Run(id string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error) {

	ns, blockID := svc.splitBlockID(id)

	matches, err := svc.blockstateParser.Run(ns, blockID, props, namespaces)
	if err != nil {
		return nil, fmt.Errorf("blockstate %s/%s: %w", ns, blockID, err)
	}

	fm := svc.resolveFlattened(matches, namespaces)

	if len(fm.Elements) == 0 {
		return nil, fmt.Errorf("no elements found for block %s", id)
	}

	isFullBlock := svc.modelAnalyzer.Run(fm.Elements)

	return &model.ResolvedBlock{
		IsFullBlock: isFullBlock,
		Elements:    fm.Elements,
		Textures:    fm.Textures,
	}, nil
}

func (svc *BlockResolver) resolveFlattened(matches []blockstateMatch, namespaces map[string][]string) *flattenedModel {
	result := &flattenedModel{Textures: make(map[string]string)}
	seenModels := make(map[string]bool)
	for _, match := range matches {
		if seenModels[match.Model] {
			continue
		}
		seenModels[match.Model] = true
		fm, err := svc.modelParser.Run(match.Model, namespaces)
		if err != nil {
			continue
		}
		result.merge(fm)
	}
	return result
}

func (svc *BlockResolver) splitBlockID(id string) (namespace, blockID string) {
	if parts := strings.SplitN(id, ":", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "minecraft", id
}
