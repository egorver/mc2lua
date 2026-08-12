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

type tintMatcher interface {
	Run(blockID string) (model.TintType, bool)
}

type colormapResolver interface {
	Run(nsRoots map[string][]string) (*model.Colormap, error)
}

var (
	waterTintColor    = model.Color{63, 118, 228}
	redstoneTintColor = model.Color{255, 0, 0}
)

type StyleIndexer struct {
	gridAnalyzer     gridAnalyzer
	elementRotator   elementRotator
	materialMatcher  materialMatcher
	partStyleMatcher indexerPartStyleMatcher
	colorExtractor   colorExtractor
	tintMatcher      tintMatcher
	colormapResolver colormapResolver
}

func NewStyleIndexer(ga gridAnalyzer, er elementRotator, mm materialMatcher, psm indexerPartStyleMatcher, ce colorExtractor, tm tintMatcher, cr colormapResolver) *StyleIndexer {
	return &StyleIndexer{
		gridAnalyzer:     ga,
		elementRotator:   er,
		materialMatcher:  mm,
		partStyleMatcher: psm,
		colorExtractor:   ce,
		tintMatcher:      tm,
		colormapResolver: cr,
	}
}

func (svc *StyleIndexer) Run(blocks []model.ResolvedBlock, nsRoots map[string][]string) *stateful.StyleIndex {
	idx := stateful.NewStyleIndex()

	colormap, _ := svc.colormapResolver.Run(nsRoots)

	for _, b := range blocks {
		material := svc.materialMatcher.Run(b.ID)

		elements := make([]model.StyledElement, len(b.Elements))
		for i, el := range b.Elements {
			elements[i] = model.StyledElement{
				From:     el.From,
				To:       el.To,
				Rotation: el.Rotation,
				Shade:    el.Shade,
				Color:    svc.elementColor(el, b.Textures, nsRoots, b.ID, colormap),
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

func (svc *StyleIndexer) elementColor(el model.ModelElement, textures map[string]string, nsRoots map[string][]string, blockID string, colormap *model.Colormap) model.Color {
	if style, ok := svc.partStyleMatcher.Run(blockID); ok && style.Color != nil {
		return *style.Color
	}
	samples := svc.resolveTextureSamples(el.Faces, textures, blockID, colormap)
	if len(samples) == 0 {
		if tex, ok := textures[minecraft.ParticleTextureKey]; ok && !strings.HasPrefix(tex, minecraft.TextureReferencePrefix) {
			samples = []minecraft.TextureSample{{
				TextureVar: tex,
				UV:         [4]float64{0, 0, minecraft.TexturePixelSize, minecraft.TexturePixelSize},
			}}
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

func (svc *StyleIndexer) resolveTextureSamples(faces map[string]model.ElementFace, textures map[string]string, blockID string, colormap *model.Colormap) []minecraft.TextureSample {
	var samples []minecraft.TextureSample
	for _, face := range faces {
		texKey := strings.TrimPrefix(face.Texture, minecraft.TextureReferencePrefix)
		texVar, ok := textures[texKey]
		if !ok {
			continue
		}
		sample := minecraft.TextureSample{
			TextureVar: texVar,
			UV:         face.UV,
		}
		if face.TintIndex != nil {
			if tint, ok := svc.tintColor(blockID, colormap); ok {
				sample.Tint = &tint
			}
		}
		samples = append(samples, sample)
	}
	return samples
}

func (svc *StyleIndexer) tintColor(blockID string, colormap *model.Colormap) (model.Color, bool) {
	tintType, ok := svc.tintMatcher.Run(blockID)
	if !ok {
		return model.Color{}, false
	}
	switch tintType {
	case model.TintGrass:
		if colormap == nil {
			return model.Color{}, false
		}
		return colormap.Grass, true
	case model.TintFoliage:
		if colormap == nil {
			return model.Color{}, false
		}
		return colormap.Foliage, true
	case model.TintWater:
		return waterTintColor, true
	case model.TintRedstone:
		return redstoneTintColor, true
	}
	return model.Color{}, false
}
