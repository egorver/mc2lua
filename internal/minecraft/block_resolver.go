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

type propertiesParser interface {
	Run(jsonStr string) map[string]string
}

type BlockResolver struct {
	blockstateParser blockstateParser
	modelAnalyzer    modelAnalyzer
	modelParser      modelParser
	propertiesParser propertiesParser
}

func NewBlockResolver(
	blockstateParser blockstateParser,
	modelAnalyzer modelAnalyzer,
	modelParser modelParser,
	propertiesParser propertiesParser,
) *BlockResolver {
	return &BlockResolver{
		blockstateParser: blockstateParser,
		modelAnalyzer:    modelAnalyzer,
		modelParser:      modelParser,
		propertiesParser: propertiesParser,
	}
}

func (svc *BlockResolver) Run(id, props string, namespaces map[string][]string) (*model.ResolvedBlock, error) {

	ns, blockID := svc.splitBlockID(id)
	parsed := svc.propertiesParser.Run(props)

	matches, err := svc.blockstateParser.Run(ns, blockID, parsed, namespaces)
	if err != nil {
		return nil, fmt.Errorf("blockstate %s/%s: %w", ns, blockID, err)
	}

	elements := svc.resolveElements(matches, namespaces)

	if len(elements) == 0 {
		return nil, fmt.Errorf("no elements found for block %s", id)
	}

	isFullBlock := svc.modelAnalyzer.Run(elements)

	return &model.ResolvedBlock{
		IsFullBlock: isFullBlock,
		Elements:    elements,
	}, nil
}

func (svc *BlockResolver) resolveElements(matches []blockstateMatch, namespaces map[string][]string) []model.ModelElement {
	var elements []model.ModelElement
	seenModels := make(map[string]bool)
	for _, match := range matches {
		if seenModels[match.Model] {
			continue
		}
		seenModels[match.Model] = true
		model, err := svc.modelParser.Run(match.Model, namespaces)
		if err != nil {
			continue
		}
		elements = append(elements, model.Elements...)
	}
	return elements
}

func (svc *BlockResolver) splitBlockID(id string) (namespace, blockID string) {
	if parts := strings.SplitN(id, ":", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "minecraft", id
}
