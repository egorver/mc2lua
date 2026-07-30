package minecraft

import (
	"fmt"
	"strings"

	"mc2lua/internal/model"
)

type blockstateParser interface {
	Run(ns, blockID string, props map[string]string, namespaces map[string][]string) ([]blockstateMatch, error)
}

type modelParser interface {
	Run(modelName string, namespaces map[string][]string) (*flattenedModel, error)
}

type textureResolver interface {
	Run(textures map[string]string) map[string]string
}

type BlockResolver struct {
	blockstateParser blockstateParser
	modelParser      modelParser
	textureResolver  textureResolver
}

func NewBlockResolver(
	blockstateParser blockstateParser,
	modelParser modelParser,
	textureResolver textureResolver,
) *BlockResolver {
	return &BlockResolver{
		blockstateParser: blockstateParser,
		modelParser:      modelParser,
		textureResolver:  textureResolver,
	}
}

func (svc *BlockResolver) Run(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error) {

	ns, blockID := svc.splitBlockID(id)

	matches, err := svc.blockstateParser.Run(ns, blockID, props, namespaces)
	if err != nil {
		return nil, fmt.Errorf("blockstate %s/%s: %w", ns, blockID, err)
	}

	fm := svc.resolveFlattened(matches, namespaces)

	if len(fm.Elements) == 0 {
		return nil, fmt.Errorf("no elements found for block %s", id)
	}

	textures := svc.textureResolver.Run(fm.Textures)

	var rotX, rotY float64
	for _, m := range matches {
		if m.RotX != 0 || m.RotY != 0 {
			rotX, rotY = m.RotX, m.RotY
			break
		}
	}

	return &model.ResolvedBlock{
		ID:       id,
		PropsKey: propsKey,
		Elements: fm.Elements,
		Textures: textures,
		RotX:     rotX,
		RotY:     rotY,
	}, nil
}

func (svc *BlockResolver) splitBlockID(id string) (namespace, blockID string) {
	if parts := strings.SplitN(id, ":", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "minecraft", id
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
