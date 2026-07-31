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
	Run(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement
}

type materialMatcher interface {
	Run(blockID string) string
}

type brightnessMatcher interface {
	Run(material string) float64
}

type colorMatcher interface {
	Run(blockID string) (model.Color, bool)
}

type colorExtractor interface {
	Run(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error)
}

type StyleIndexer struct {
	gridAnalyzer      gridAnalyzer
	elementRotator    elementRotator
	materialMatcher   materialMatcher
	brightnessMatcher brightnessMatcher
	colorMatcher      colorMatcher
	colorExtractor    colorExtractor
}

func NewStyleIndexer(ga gridAnalyzer, er elementRotator, mm materialMatcher, bm brightnessMatcher, ce colorExtractor, cm colorMatcher) *StyleIndexer {
	return &StyleIndexer{
		gridAnalyzer:      ga,
		elementRotator:    er,
		materialMatcher:   mm,
		brightnessMatcher: bm,
		colorExtractor:    ce,
		colorMatcher:      cm,
	}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock, nsRoots map[string][]string) *stateful.StyleIndex {
	idx := stateful.NewStyleIndex()

	for _, b := range blocks {
		material := svc.materialMatcher.Run(b.ID)
		brightness := svc.brightnessMatcher.Run(material)

		elements := make([]model.StyledElement, len(b.Elements))
		for i, el := range b.Elements {
			elements[i] = model.StyledElement{
				From:     el.From,
				To:       el.To,
				Rotation: el.Rotation,
				Shade:    el.Shade,
				Color:    svc.elementColor(el, b.Textures, nsRoots, b.ID, brightness),
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

func (svc *StyleIndexer) elementColor(el model.ModelElement, textures map[string]string, nsRoots map[string][]string, blockID string, brightness float64) model.Color {
	if c, ok := svc.colorMatcher.Run(blockID); ok {
		return svc.scaleColor(c, brightness)
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
	return svc.scaleColor(color, brightness)
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

func (svc *StyleIndexer) scaleColor(c model.Color, factor float64) model.Color {
	if factor <= 0 {
		return c
	}
	return model.Color{
		svc.clampByte(int(float64(c[0]) * factor)),
		svc.clampByte(int(float64(c[1]) * factor)),
		svc.clampByte(int(float64(c[2]) * factor)),
	}
}

func (svc *StyleIndexer) clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
