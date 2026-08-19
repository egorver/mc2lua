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

type elementRotator interface {
	RunModel(elements []model.ModelElement, rotX, rotY float64) []model.ModelElement
}

const trapdoorIDSubstring = "trapdoor"

type BlockResolver struct {
	blockstateParser blockstateParser
	modelParser      modelParser
	textureResolver  textureResolver
	elementRotator   elementRotator
}

func NewBlockResolver(
	blockstateParser blockstateParser,
	modelParser modelParser,
	textureResolver textureResolver,
	elementRotator elementRotator,
) *BlockResolver {
	return &BlockResolver{
		blockstateParser: blockstateParser,
		modelParser:      modelParser,
		textureResolver:  textureResolver,
		elementRotator:   elementRotator,
	}
}

func (svc *BlockResolver) Run(id string, propsKey string, props map[string]string, namespaces map[string][]string) (*model.ResolvedBlock, error) {

	ns, blockID := svc.splitBlockID(id)

	matches, err := svc.blockstateParser.Run(ns, blockID, props, namespaces)
	if err != nil {
		return nil, fmt.Errorf("blockstate %s/%s: %w", ns, blockID, err)
	}

	fm, rotX, rotY := svc.resolveFlattened(matches, namespaces)

	if strings.Contains(blockID, trapdoorIDSubstring) {
		rotX, rotY = svc.correctTrapdoorRotation(props, rotX, rotY)
	}

	if len(fm.Elements) == 0 {
		return nil, fmt.Errorf("no elements found for block %s", id)
	}

	return &model.ResolvedBlock{
		ID:       id,
		PropsKey: propsKey,
		Elements: fm.Elements,
		Textures: svc.textureResolver.Run(fm.Textures),
		RotX:     rotX,
		RotY:     rotY,
	}, nil
}

func (svc *BlockResolver) splitBlockID(id string) (namespace, blockID string) {
	if parts := strings.SplitN(id, NamespaceSeparator, 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return DefaultNamespace, id
}

func (svc *BlockResolver) resolveFlattened(matches []blockstateMatch, namespaces map[string][]string) (*flattenedModel, float64, float64) {
	result := &flattenedModel{Textures: make(map[string]string)}
	seen := make(map[string]bool)

	bake, rotX, rotY := svc.rotationStrategy(matches)

	for _, match := range matches {
		key := svc.rotationKey(match)
		if seen[key] {
			continue
		}
		seen[key] = true
		fm := svc.resolveMatch(match, namespaces, bake)
		if fm == nil {
			continue
		}
		result.merge(fm)
	}
	return result, rotX, rotY
}

func (svc *BlockResolver) rotationStrategy(matches []blockstateMatch) (bake bool, rotX, rotY float64) {
	if len(matches) == 0 {
		return false, 0, 0
	}
	rotX, rotY = matches[0].RotX, matches[0].RotY
	for _, m := range matches[1:] {
		if m.RotX != rotX || m.RotY != rotY {
			return true, 0, 0
		}
	}
	return false, rotX, rotY
}

func (svc *BlockResolver) rotationKey(match blockstateMatch) string {
	return fmt.Sprintf("%s|%g|%g", match.Model, match.RotX, match.RotY)
}

func (svc *BlockResolver) resolveMatch(match blockstateMatch, namespaces map[string][]string, bake bool) *flattenedModel {
	fm, err := svc.modelParser.Run(match.Model, namespaces)
	if err != nil {
		return nil
	}
	if !bake || (match.RotX == 0 && match.RotY == 0) {
		return fm
	}
	return &flattenedModel{
		Elements: svc.elementRotator.RunModel(fm.Elements, match.RotX, match.RotY),
		Textures: fm.Textures,
	}
}

func (svc *BlockResolver) correctTrapdoorRotation(props map[string]string, rotX, rotY float64) (float64, float64) {
	if props["half"] != "top" || props["open"] != "true" {
		return rotX, rotY
	}
	switch rotY {
	case 90:
		return 0, 270
	case 270:
		return 0, 90
	default:
		return 0, rotY
	}
}
