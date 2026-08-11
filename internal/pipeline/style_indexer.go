package pipeline

import (
	"strings"

	"mc2lua/internal/minecraft"
	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type gridAnalyzer interface {
	Run(elements []model.StyledElement) model.GridAlignment
}

type elementRotator interface {
	RunStyled(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement
}

type materialMatcher interface {
	Run(blockID string) string
}

type indexerPartStyleMatcher interface {
	Run(blockID string) (model.PartStyle, bool)
}

type colorExtractor interface {
	Run(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error)
}

type StyleIndexer struct {
	gridAnalyzer     gridAnalyzer
	elementRotator   elementRotator
	materialMatcher  materialMatcher
	partStyleMatcher indexerPartStyleMatcher
	colorExtractor   colorExtractor
}

func NewStyleIndexer(ga gridAnalyzer, er elementRotator, mm materialMatcher, psm indexerPartStyleMatcher, ce colorExtractor) *StyleIndexer {
	return &StyleIndexer{
		gridAnalyzer:     ga,
		elementRotator:   er,
		materialMatcher:  mm,
		partStyleMatcher: psm,
		colorExtractor:   ce,
	}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock, nsRoots map[string][]string) *stateful.StyleIndex {
	idx := stateful.NewStyleIndex()

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

		rotated := svc.elementRotator.RunStyled(elements, b.RotX, b.RotY)
		gridAlignment := svc.gridAnalyzer.Run(rotated)

		styled := model.StyledBlock{
			ID:            b.ID,
			PropsKey:      b.PropsKey,
			GridAlignment: gridAlignment,
			RotX:          b.RotX,
			RotY:          b.RotY,
			Elements:      rotated,
		}

		idx.Add(b.ID, b.PropsKey, styled)
	}

	return idx
}

func (svc *StyleIndexer) elementColor(el model.ModelElement, textures map[string]string, nsRoots map[string][]string, blockID string) model.Color {
	if style, ok := svc.partStyleMatcher.Run(blockID); ok && style.Color != nil {
		return *style.Color
	}
	samples := svc.resolveTextureSamples(el.Faces, textures)
	if len(samples) == 0 {
		if tex, ok := textures["particle"]; ok && !strings.HasPrefix(tex, "#") {
			samples = []minecraft.TextureSample{{TextureVar: tex, UV: [4]float64{0, 0, 16, 16}}}
		}
	}
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
