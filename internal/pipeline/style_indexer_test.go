package pipeline

import (
	"os"
	"testing"

	"mc2lua/internal/stateful"
	"mc2lua/internal/minecraft"
	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockGridAnalyzer struct {
	runFn func(elements []model.StyledElement) model.GridAlignment
}

func (m *mockGridAnalyzer) Run(elements []model.StyledElement) model.GridAlignment {
	if m.runFn != nil {
		return m.runFn(elements)
	}
	return model.GridNotAligned
}

type mockMaterialMatcher struct {
	runFn func(blockID string) string
}

func (m *mockMaterialMatcher) Run(blockID string) string {
	if m.runFn != nil {
		return m.runFn(blockID)
	}
	return model.DefaultMaterial
}

type mockBrightnessMatcher struct {
	runFn func(material string) float64
}

func (m *mockBrightnessMatcher) Run(material string) float64 {
	if m.runFn != nil {
		return m.runFn(material)
	}
	return 1.0
}

type mockElementRotator struct {
	runFn func(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement
}

func (m *mockElementRotator) RunStyled(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement {
	if m.runFn != nil {
		return m.runFn(elements, rotX, rotY)
	}
	return elements
}

type mockColorExtractor struct {
	runFn func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error)
}

func (m *mockColorExtractor) Run(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
	if m.runFn != nil {
		return m.runFn(samples, nsRoots, blockID)
	}
	return model.DefaultColor, nil
}

type mockColorMatcher struct {
	runFn func(blockID string) (model.Color, bool)
}

func (m *mockColorMatcher) Run(blockID string) (model.Color, bool) {
	if m.runFn != nil {
		return m.runFn(blockID)
	}
	return model.Color{}, false
}

func TestStyleIndexer_ElementColorParticleFallback(t *testing.T) {
	t.Parallel()

	noFaces := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}

	tests := []struct {
		name       string
		el         model.ModelElement
		textures   map[string]string
		brightness float64
		extractor  func(samples []minecraft.TextureSample) model.Color
		want       model.Color
	}{
		{
			name:       "no faces uses particle texture",
			el:         noFaces,
			textures:   map[string]string{"particle": "block/water_still"},
			brightness: 1.0,
			extractor: func(samples []minecraft.TextureSample) model.Color {
				require.Len(t, samples, 1)
				require.Equal(t, "block/water_still", samples[0].TextureVar)
				require.Equal(t, [4]float64{0, 0, 16, 16}, samples[0].UV)
				return model.Color{10, 20, 30}
			},
			want: model.Color{10, 20, 30},
		},
		{
			name:       "brightness applied to particle color",
			el:         noFaces,
			textures:   map[string]string{"particle": "block/lava_still"},
			brightness: 1.2,
			extractor: func(samples []minecraft.TextureSample) model.Color {
				return model.Color{200, 100, 50}
			},
			want: model.Color{240, 120, 60},
		},
		{
			name:       "unresolved particle reference falls back to default",
			el:         noFaces,
			textures:   map[string]string{"particle": "#missing"},
			brightness: 1.0,
			extractor: func(samples []minecraft.TextureSample) model.Color {
				require.Fail(t, "extractor should not be called")
				return model.DefaultColor
			},
			want: model.DefaultColor,
		},
		{
			name:       "no particle texture falls back to default",
			el:         noFaces,
			textures:   nil,
			brightness: 1.0,
			extractor: func(samples []minecraft.TextureSample) model.Color {
				require.Fail(t, "extractor should not be called")
				return model.DefaultColor
			},
			want: model.DefaultColor,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ce := &mockColorExtractor{runFn: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
				return tt.extractor(samples), nil
			}}
			si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, ce, &mockColorMatcher{})
			got := si.elementColor(tt.el, tt.textures, nil, "minecraft:water", tt.brightness)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStyleIndexer_New(t *testing.T) {
	t.Parallel()

	si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, &mockColorExtractor{}, &mockColorMatcher{})
	require.NotNil(t, si)
}

func TestStyleIndexer_ElementColorOverride(t *testing.T) {
	t.Parallel()

	el := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}

	t.Run("configured color wins over textures", func(t *testing.T) {
		cm := &mockColorMatcher{runFn: func(blockID string) (model.Color, bool) {
			require.Equal(t, "minecraft:water", blockID)
			return model.Color{38, 94, 173}, true
		}}
		ce := &mockColorExtractor{runFn: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
			require.Fail(t, "extractor should not be called when override exists")
			return model.DefaultColor, nil
		}}
		si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, ce, cm)
		got := si.elementColor(el, map[string]string{"particle": "block/water_still"}, nil, "minecraft:water", 1.0)
		require.Equal(t, model.Color{38, 94, 173}, got)
	})

	t.Run("brightness applied to override color", func(t *testing.T) {
		cm := &mockColorMatcher{runFn: func(blockID string) (model.Color, bool) {
			return model.Color{200, 82, 18}, true
		}}
		si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, &mockColorExtractor{}, cm)
		got := si.elementColor(el, nil, nil, "minecraft:lava", 1.2)
		require.Equal(t, model.Color{240, 98, 21}, got)
	})

	t.Run("unconfigured block falls through to textures", func(t *testing.T) {
		cm := &mockColorMatcher{runFn: func(blockID string) (model.Color, bool) {
			return model.Color{}, false
		}}
		ce := &mockColorExtractor{runFn: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
			return model.Color{10, 20, 30}, nil
		}}
		si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, ce, cm)
		got := si.elementColor(el, map[string]string{"particle": "block/stone"}, nil, "minecraft:stone", 1.0)
		require.Equal(t, model.Color{10, 20, 30}, got)
	})
}

func TestStyleIndexer_Run(t *testing.T) {
	t.Parallel()

	fullBlock := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}
	halfBlock := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}}

	tests := []struct {
		name       string
		blocks     []model.ResolvedBlock
		analyzer   func([]model.StyledElement) model.GridAlignment
		matcher    func(blockID string) string
		brightness func(material string) float64
		wantLen    int
		wantCheck  func(t *testing.T, idx *stateful.StyleIndex)
	}{
		{
			name:    "empty input",
			blocks:  nil,
			wantLen: 0,
		},
		{
			name: "material resolved from matcher",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone", Elements: []model.ModelElement{fullBlock}},
				{ID: "minecraft:oak_planks", Elements: []model.ModelElement{fullBlock}},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridFullBlock },
			matcher: func(blockID string) string {
				if blockID == "minecraft:stone" {
					return "Slate"
				}
				return "Wood"
			},
			wantLen: 2,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				stone, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.Equal(t, "Slate", stone.Elements[0].Material)

				planks, ok := idx.Get("minecraft:oak_planks", "")
				require.True(t, ok)
				require.Equal(t, "Wood", planks.Elements[0].Material)
			},
		},
		{
			name: "single full block",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone", Elements: []model.ModelElement{fullBlock}},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridFullBlock },
			wantLen:  1,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				b, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.Equal(t, model.GridFullBlock, b.GridAlignment)
				require.Len(t, b.Elements, 1)
			},
		},
		{
			name: "sub-block returns GridSubBlock",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone_slab", PropsKey: "type=bottom", Elements: []model.ModelElement{halfBlock}},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridSubBlock },
			wantLen:  1,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				b, ok := idx.Get("minecraft:stone_slab", "type=bottom")
				require.True(t, ok)
				require.Equal(t, model.GridSubBlock, b.GridAlignment)
				require.Len(t, b.Elements, 1)
			},
		},
		{
			name: "non-aligned block",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:oak_fence", PropsKey: "water=true", Elements: []model.ModelElement{halfBlock}},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridNotAligned },
			wantLen:  1,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				b, ok := idx.Get("minecraft:oak_fence", "water=true")
				require.True(t, ok)
				require.Equal(t, model.GridNotAligned, b.GridAlignment)
				require.Len(t, b.Elements, 1)
			},
		},
		{
			name: "multiple blocks",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone", Elements: []model.ModelElement{fullBlock}},
				{ID: "minecraft:dirt", Elements: []model.ModelElement{fullBlock}},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridFullBlock },
			wantLen:  2,
		},
		{
			name: "element faces stripped in styled element",
			blocks: []model.ResolvedBlock{
				{
					ID: "minecraft:stone",
					Elements: []model.ModelElement{
						{
							From:  model.Vector3{0, 0, 0},
							To:    model.Vector3{16, 16, 16},
							Shade: true,
							Faces: map[string]model.ElementFace{"up": {UV: [4]float64{0, 0, 16, 16}}},
						},
					},
				},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridFullBlock },
			wantLen:  1,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				b, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.Len(t, b.Elements, 1)
				require.Equal(t, model.Vector3{0, 0, 0}, b.Elements[0].From)
				require.Equal(t, model.Vector3{16, 16, 16}, b.Elements[0].To)
				require.True(t, b.Elements[0].Shade)
			},
		},
		{
			name: "brightness applied to element color",
			blocks: []model.ResolvedBlock{
				{
					ID: "minecraft:stone",
					Elements: []model.ModelElement{
						{
							From:  model.Vector3{0, 0, 0},
							To:    model.Vector3{16, 16, 16},
							Shade: true,
							Faces: map[string]model.ElementFace{"up": {Texture: "#up", UV: [4]float64{0, 0, 16, 16}}},
						},
					},
					Textures: map[string]string{"up": "block/stone"},
				},
			},
			analyzer:   func(elements []model.StyledElement) model.GridAlignment { return model.GridFullBlock },
			matcher:    func(blockID string) string { return "Slate" },
			brightness: func(material string) float64 { return 1.2 },
			wantLen:    1,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				b, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.Equal(t, model.Color{229, 229, 229}, b.Elements[0].Color)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ga := &mockGridAnalyzer{runFn: tt.analyzer}
			mm := &mockMaterialMatcher{runFn: tt.matcher}
			bm := &mockBrightnessMatcher{runFn: tt.brightness}
			si := NewStyleIndexer(ga, &mockElementRotator{}, mm, bm, &mockColorExtractor{}, &mockColorMatcher{})

			idx := si.Run(tt.blocks, nil)
			require.Equal(t, tt.wantLen, idx.Len())
			if tt.wantCheck != nil {
				tt.wantCheck(t, idx)
			}
		})
	}
}

func TestStyleIndexer_ScaleColor(t *testing.T) {
	t.Parallel()

	svc := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, &mockColorExtractor{}, &mockColorMatcher{})

	tests := []struct {
		name   string
		color  model.Color
		factor float64
		want   model.Color
	}{
		{name: "zero factor returns original", color: model.Color{100, 100, 100}, factor: 0, want: model.Color{100, 100, 100}},
		{name: "negative factor returns original", color: model.Color{100, 100, 100}, factor: -1, want: model.Color{100, 100, 100}},
		{name: "unit factor returns original", color: model.Color{100, 100, 100}, factor: 1, want: model.Color{100, 100, 100}},
		{name: "scales up color", color: model.Color{100, 100, 100}, factor: 1.5, want: model.Color{150, 150, 150}},
		{name: "rounds down on scale", color: model.Color{100, 100, 100}, factor: 1.25, want: model.Color{125, 125, 125}},
		{name: "clamps overflow to 255", color: model.Color{200, 200, 200}, factor: 2, want: model.Color{255, 255, 255}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, svc.scaleColor(tt.color, tt.factor))
		})
	}
}

func TestStyleIndexer_ClampByte(t *testing.T) {
	t.Parallel()

	svc := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, &mockColorExtractor{}, &mockColorMatcher{})

	tests := []struct {
		name string
		in   int
		want uint8
	}{
		{name: "negative clamped to zero", in: -10, want: 0},
		{name: "zero", in: 0, want: 0},
		{name: "mid value", in: 128, want: 128},
		{name: "max value", in: 255, want: 255},
		{name: "overflow clamped to 255", in: 300, want: 255},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, svc.clampByte(tt.in))
		})
	}
}

func TestStyleIndexer_ElementColor(t *testing.T) {
	t.Parallel()

	elementWithFace := func(texture string) model.ModelElement {
		return model.ModelElement{
			Faces: map[string]model.ElementFace{"up": {Texture: texture, UV: [4]float64{0, 0, 16, 16}}},
		}
	}

	tests := []struct {
		name       string
		el         model.ModelElement
		textures   map[string]string
		colorRun   func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error)
		brightness float64
		want       model.Color
	}{
		{
			name:     "no faces returns default color",
			el:       model.ModelElement{},
			textures: map[string]string{"up": "block/stone"},
			want:     model.DefaultColor,
		},
		{
			name:     "missing texture key returns default color",
			el:       elementWithFace("#up"),
			textures: map[string]string{"down": "block/dirt"},
			want:     model.DefaultColor,
		},
		{
			name:     "color extractor error returns default color",
			el:       elementWithFace("#up"),
			textures: map[string]string{"up": "block/stone"},
			colorRun: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
				return model.DefaultColor, os.ErrPermission
			},
			want: model.DefaultColor,
		},
		{
			name:     "successful extraction returns color",
			el:       elementWithFace("#up"),
			textures: map[string]string{"up": "block/stone"},
			colorRun: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
				return model.Color{100, 100, 100}, nil
			},
			brightness: 1.0,
			want:       model.Color{100, 100, 100},
		},
		{
			name:     "brightness scales extracted color",
			el:       elementWithFace("#up"),
			textures: map[string]string{"up": "block/stone"},
			colorRun: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
				return model.Color{100, 100, 100}, nil
			},
			brightness: 1.5,
			want:       model.Color{150, 150, 150},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ce := &mockColorExtractor{runFn: tt.colorRun}
			svc := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, ce, &mockColorMatcher{})

			got := svc.elementColor(tt.el, tt.textures, nil, "minecraft:stone", tt.brightness)
			require.Equal(t, tt.want, got)
		})
	}
}
