package pipeline

import (
	"fmt"
	"os"
	"testing"

	"mc2lua/internal/minecraft"
	"mc2lua/internal/model"
	"mc2lua/internal/stateful"

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

type mockTintMatcher struct {
	runFn func(blockID string) (model.TintType, bool)
}

func (m *mockTintMatcher) Run(blockID string) (model.TintType, bool) {
	if m.runFn != nil {
		return m.runFn(blockID)
	}
	return model.TintGrass, false
}

type mockColormapResolver struct {
	runFn func(nsRoots map[string][]string) (*model.Colormap, error)
}

func (m *mockColormapResolver) Run(nsRoots map[string][]string) (*model.Colormap, error) {
	if m.runFn != nil {
		return m.runFn(nsRoots)
	}
	return nil, nil
}

func TestStyleIndexer_ElementColorParticleFallback(t *testing.T) {
	t.Parallel()

	noFaces := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}

	tests := []struct {
		name      string
		el        model.ModelElement
		textures  map[string]string
		extractor func(samples []minecraft.TextureSample) model.Color
		want      model.Color
	}{
		{
			name:     "no faces uses particle texture",
			el:       noFaces,
			textures: map[string]string{"particle": "block/water_still"},
			extractor: func(samples []minecraft.TextureSample) model.Color {
				require.Len(t, samples, 1)
				require.Equal(t, "block/water_still", samples[0].TextureVar)
				require.Equal(t, [4]float64{0, 0, 16, 16}, samples[0].UV)
				return model.Color{10, 20, 30}
			},
			want: model.Color{10, 20, 30},
		},
		{
			name:     "particle color returned raw",
			el:       noFaces,
			textures: map[string]string{"particle": "block/lava_still"},
			extractor: func(samples []minecraft.TextureSample) model.Color {
				return model.Color{200, 100, 50}
			},
			want: model.Color{200, 100, 50},
		},
		{
			name:     "unresolved particle reference falls back to default",
			el:       noFaces,
			textures: map[string]string{"particle": "#missing"},
			extractor: func(samples []minecraft.TextureSample) model.Color {
				require.Fail(t, "extractor should not be called")
				return model.DefaultColor
			},
			want: model.DefaultColor,
		},
		{
			name:     "no particle texture falls back to default",
			el:       noFaces,
			textures: nil,
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
			si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, ce, &mockTintMatcher{}, &mockColormapResolver{})
			got := si.elementColor(tt.el, tt.textures, nil, "minecraft:water", nil)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestStyleIndexer_New(t *testing.T) {
	t.Parallel()

	si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, &mockColorExtractor{}, &mockTintMatcher{}, &mockColormapResolver{})
	require.NotNil(t, si)
}

func TestStyleIndexer_ElementColorOverride(t *testing.T) {
	t.Parallel()

	el := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}

	t.Run("configured color wins over textures", func(t *testing.T) {
		psm := &mockPartStyleMatcher{styles: map[string]model.PartStyle{
			"minecraft:water": {Color: &model.Color{38, 94, 173}},
		}}
		ce := &mockColorExtractor{runFn: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
			require.Fail(t, "extractor should not be called when override exists")
			return model.DefaultColor, nil
		}}
		si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, psm, ce, &mockTintMatcher{}, &mockColormapResolver{})
		got := si.elementColor(el, map[string]string{"particle": "block/water_still"}, nil, "minecraft:water", nil)
		require.Equal(t, model.Color{38, 94, 173}, got)
	})

	t.Run("style without color falls through to textures", func(t *testing.T) {
		psm := &mockPartStyleMatcher{styles: map[string]model.PartStyle{
			"minecraft:stone": {Transparency: floatPtr(0.5)},
		}}
		ce := &mockColorExtractor{runFn: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
			return model.Color{10, 20, 30}, nil
		}}
		si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, psm, ce, &mockTintMatcher{}, &mockColormapResolver{})
		got := si.elementColor(el, map[string]string{"particle": "block/stone"}, nil, "minecraft:stone", nil)
		require.Equal(t, model.Color{10, 20, 30}, got)
	})

	t.Run("unconfigured block falls through to textures", func(t *testing.T) {
		psm := &mockPartStyleMatcher{}
		ce := &mockColorExtractor{runFn: func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
			return model.Color{10, 20, 30}, nil
		}}
		si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, psm, ce, &mockTintMatcher{}, &mockColormapResolver{})
		got := si.elementColor(el, map[string]string{"particle": "block/stone"}, nil, "minecraft:stone", nil)
		require.Equal(t, model.Color{10, 20, 30}, got)
	})
}

func TestStyleIndexer_Run(t *testing.T) {
	t.Parallel()

	fullBlock := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}
	halfBlock := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}}

	tests := []struct {
		name      string
		blocks    []model.ResolvedBlock
		analyzer  func([]model.StyledElement) model.GridAlignment
		matcher   func(blockID string) string
		wantLen   int
		wantCheck func(t *testing.T, idx *stateful.StyleIndex)
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
			name: "rotation angles preserved in styled block",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone", PropsKey: "facing=north", RotX: 90, RotY: 180, Elements: []model.ModelElement{fullBlock}},
			},
			analyzer: func(elements []model.StyledElement) model.GridAlignment { return model.GridFullBlock },
			wantLen:  1,
			wantCheck: func(t *testing.T, idx *stateful.StyleIndex) {
				b, ok := idx.Get("minecraft:stone", "facing=north")
				require.True(t, ok)
				require.Equal(t, 90.0, b.RotX)
				require.Equal(t, 180.0, b.RotY)
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ga := &mockGridAnalyzer{runFn: tt.analyzer}
			mm := &mockMaterialMatcher{runFn: tt.matcher}
			si := NewStyleIndexer(ga, &mockElementRotator{}, mm, &mockPartStyleMatcher{}, &mockColorExtractor{}, &mockTintMatcher{}, &mockColormapResolver{})

			idx := si.Run(tt.blocks, nil)
			require.Equal(t, tt.wantLen, idx.Len())
			if tt.wantCheck != nil {
				tt.wantCheck(t, idx)
			}
		})
	}
}

func TestStyleIndexer_RotationPassedUnchangedToRotator(t *testing.T) {
	t.Parallel()

	var gotRotX, gotRotY float64
	rotator := &mockElementRotator{runFn: func(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement {
		gotRotX, gotRotY = rotX, rotY
		return elements
	}}

	si := NewStyleIndexer(&mockGridAnalyzer{}, rotator, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, &mockColorExtractor{}, &mockTintMatcher{}, &mockColormapResolver{})

	block := model.ResolvedBlock{
		ID:   "minecraft:stone",
		RotX: 90,
		RotY: 180,
		Elements: []model.ModelElement{
			{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
		},
	}

	idx := si.Run([]model.ResolvedBlock{block}, nil)

	require.Equal(t, 90.0, gotRotX)
	require.Equal(t, 180.0, gotRotY)

	b, ok := idx.Get("minecraft:stone", "")
	require.True(t, ok)
	require.Equal(t, 90.0, b.RotX)
	require.Equal(t, 180.0, b.RotY)
}

func TestStyleIndexer_ZeroRotationKeptZero(t *testing.T) {
	t.Parallel()

	var rotatorCalls int
	rotator := &mockElementRotator{runFn: func(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement {
		rotatorCalls++
		return elements
	}}

	si := NewStyleIndexer(&mockGridAnalyzer{}, rotator, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, &mockColorExtractor{}, &mockTintMatcher{}, &mockColormapResolver{})

	block := model.ResolvedBlock{
		ID: "minecraft:stone",
		Elements: []model.ModelElement{
			{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
		},
	}

	idx := si.Run([]model.ResolvedBlock{block}, nil)

	require.Equal(t, 1, rotatorCalls)

	b, ok := idx.Get("minecraft:stone", "")
	require.True(t, ok)
	require.Equal(t, 0.0, b.RotX)
	require.Equal(t, 0.0, b.RotY)
}

func TestStyleIndexer_ElementColor(t *testing.T) {
	t.Parallel()

	elementWithFace := func(texture string) model.ModelElement {
		return model.ModelElement{
			Faces: map[string]model.ElementFace{"up": {Texture: texture, UV: [4]float64{0, 0, 16, 16}}},
		}
	}

	tests := []struct {
		name     string
		el       model.ModelElement
		textures map[string]string
		colorRun func(samples []minecraft.TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error)
		want     model.Color
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
			want: model.Color{100, 100, 100},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ce := &mockColorExtractor{runFn: tt.colorRun}
			svc := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, ce, &mockTintMatcher{}, &mockColormapResolver{})

			got := svc.elementColor(tt.el, tt.textures, nil, "minecraft:stone", nil)
			require.Equal(t, tt.want, got)
		})
	}
}

func tintMultiset(tints []*model.Color) map[string]int {
	multiset := make(map[string]int, len(tints))
	for _, tint := range tints {
		if tint == nil {
			multiset["nil"]++
			continue
		}
		multiset[fmt.Sprintf("%d,%d,%d", tint[0], tint[1], tint[2])]++
	}
	return multiset
}

func TestStyleIndexer_ResolveTextureSamplesTint(t *testing.T) {
	t.Parallel()

	textures := map[string]string{"all": "block/stone"}
	tintIndex := 0

	tests := []struct {
		name     string
		faces    map[string]model.ElementFace
		matcher  func(blockID string) (model.TintType, bool)
		colormap *model.Colormap
		wantTint []*model.Color
	}{
		{
			name:     "no tintindex leaves tint nil",
			faces:    map[string]model.ElementFace{"up": {Texture: "#all", UV: [4]float64{0, 0, 16, 16}}},
			wantTint: []*model.Color{nil},
		},
		{
			name:  "grass tint applied",
			faces: map[string]model.ElementFace{"up": {Texture: "#all", UV: [4]float64{0, 0, 16, 16}, TintIndex: &tintIndex}},
			matcher: func(blockID string) (model.TintType, bool) {
				return model.TintGrass, true
			},
			colormap: &model.Colormap{Grass: model.Color{145, 189, 89}, Foliage: model.Color{119, 171, 47}},
			wantTint: []*model.Color{{145, 189, 89}},
		},
		{
			name:  "unmatched block keeps tint nil",
			faces: map[string]model.ElementFace{"up": {Texture: "#all", UV: [4]float64{0, 0, 16, 16}, TintIndex: &tintIndex}},
			matcher: func(blockID string) (model.TintType, bool) {
				return model.TintGrass, false
			},
			colormap: &model.Colormap{Grass: model.Color{145, 189, 89}},
			wantTint: []*model.Color{nil},
		},
		{
			name:  "grass without colormap keeps tint nil",
			faces: map[string]model.ElementFace{"up": {Texture: "#all", UV: [4]float64{0, 0, 16, 16}, TintIndex: &tintIndex}},
			matcher: func(blockID string) (model.TintType, bool) {
				return model.TintGrass, true
			},
			colormap: nil,
			wantTint: []*model.Color{nil},
		},
		{
			name:  "water tint works without colormap",
			faces: map[string]model.ElementFace{"up": {Texture: "#all", UV: [4]float64{0, 0, 16, 16}, TintIndex: &tintIndex}},
			matcher: func(blockID string) (model.TintType, bool) {
				return model.TintWater, true
			},
			colormap: nil,
			wantTint: []*model.Color{{63, 118, 228}},
		},
		{
			name: "multiple faces tint individually",
			faces: map[string]model.ElementFace{
				"up":   {Texture: "#all", UV: [4]float64{0, 0, 16, 16}, TintIndex: &tintIndex},
				"down": {Texture: "#all", UV: [4]float64{0, 0, 16, 16}},
			},
			matcher: func(blockID string) (model.TintType, bool) {
				return model.TintFoliage, true
			},
			colormap: &model.Colormap{Foliage: model.Color{119, 171, 47}},
			wantTint: []*model.Color{{119, 171, 47}, nil},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			tm := &mockTintMatcher{runFn: tt.matcher}
			si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, &mockColorExtractor{}, tm, &mockColormapResolver{})

			got := si.resolveTextureSamples(tt.faces, textures, "minecraft:grass_block", tt.colormap)
			require.Len(t, got, len(tt.wantTint))
			gotTints := make([]*model.Color, len(got))
			for i, s := range got {
				gotTints[i] = s.Tint
			}
			require.Equal(t, tintMultiset(tt.wantTint), tintMultiset(gotTints))
		})
	}
}

func TestStyleIndexer_Run_ResolvesColormap(t *testing.T) {
	t.Parallel()

	var gotRoots map[string][]string
	cr := &mockColormapResolver{runFn: func(nsRoots map[string][]string) (*model.Colormap, error) {
		gotRoots = nsRoots
		return &model.Colormap{Grass: model.Color{145, 189, 89}}, nil
	}}
	si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockPartStyleMatcher{}, &mockColorExtractor{}, &mockTintMatcher{}, cr)

	si.Run([]model.ResolvedBlock{
		{ID: "minecraft:stone", Elements: []model.ModelElement{{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}}},
	}, map[string][]string{"minecraft": {"assets/minecraft"}})

	require.Equal(t, map[string][]string{"minecraft": {"assets/minecraft"}}, gotRoots)
}
