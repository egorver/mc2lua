package pipeline

import (
	"strings"

	"mc2lua/internal/index"
	"mc2lua/internal/minecraft"
	"mc2lua/internal/model"
)

type gridAnalyzer interface {
	Run(elements []model.StyledElement) model.GridAlignment
}

type elementRotator interface {
	Run(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement
}

type materialMatcher interface {
	Run(blockID string) string
}

type colorExtractor interface {
	Run(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error)
}

type StyleIndexer struct {
	gridAnalyzer    gridAnalyzer
	elementRotator  elementRotator
	materialMatcher materialMatcher
	colorExtractor  colorExtractor
}

func NewStyleIndexer(ga gridAnalyzer, er elementRotator, mm materialMatcher, ce colorExtractor) *StyleIndexer {
	return &StyleIndexer{
		gridAnalyzer:    ga,
		elementRotator:  er,
		materialMatcher: mm,
		colorExtractor:  ce,
	}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock, nsRoots map[string][]string) *index.StyleIndex {
	idx := index.NewStyleIndex()

	for _, b := range blocks {
		material := svc.materialMatcher.Run(b.ID)

		elements := make([]model.StyledElement, len(b.Elements))
		for i, el := range b.Elements {
			elements[i] = model.StyledElement{
				From:     el.From,
				To:       el.To,
				Rotation: el.Rotation,
				Shade:    el.Shade,
				Color:    svc.elementColor(el, b.Textures, nsRoots, b.ID),
				Material: material,
			}
		}

		rotated := svc.elementRotator.Run(elements, b.RotX, b.RotY)
		gridAlignment := svc.gridAnalyzer.Run(rotated)

		styled := model.StyledBlock{
			ID:            b.ID,
			PropsKey:      b.PropsKey,
			GridAlignment: gridAlignment,
			Elements:      rotated,
		}

		idx.Add(b.ID, b.PropsKey, styled)
	}

	return idx
}

func (svc *StyleIndexer) elementColor(el model.ModelElement, textures map[string]string, nsRoots map[string][]string, blockID string) model.Color {
	samples := svc.resolveTextureSamples(el.Faces, textures)
	if len(samples) == 0 {
		return model.DefaultColor
	}
	color, err := svc.colorExtractor.Run(samples, nsRoots, blockID)
	if err != nil {
		return model.DefaultColor
	}
	return color
}

func (svc *StyleIndexer) resolveTextureSamples(faces map[string]model.ElementFace, textures map[string]string) []minecraft.TextureSample {
	var samples []minecraft.TextureSample
	for _, face := range faces {
		texKey := strings.TrimPrefix(face.Texture, "#")
		texVar, ok := textures[texKey]
		if !ok {
			continue
		}
		samples = append(samples, minecraft.TextureSample{
			TextureVar: texVar,
			UV:         face.UV,
		})
	}
	return samples
}
