package pipeline

import (
	"testing"

	"mc2lua/internal/index"
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

func (m *mockElementRotator) Run(elements []model.StyledElement, rotX, rotY float64) []model.StyledElement {
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

func TestStyleIndexer_New(t *testing.T) {
	t.Parallel()

	si := NewStyleIndexer(&mockGridAnalyzer{}, &mockElementRotator{}, &mockMaterialMatcher{}, &mockBrightnessMatcher{}, &mockColorExtractor{})
	require.NotNil(t, si)
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
		wantCheck  func(t *testing.T, idx *index.StyleIndex)
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
			wantCheck: func(t *testing.T, idx *index.StyleIndex) {
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
			wantCheck: func(t *testing.T, idx *index.StyleIndex) {
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
			wantCheck: func(t *testing.T, idx *index.StyleIndex) {
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
			wantCheck: func(t *testing.T, idx *index.StyleIndex) {
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
			wantCheck: func(t *testing.T, idx *index.StyleIndex) {
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
			wantCheck: func(t *testing.T, idx *index.StyleIndex) {
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
			si := NewStyleIndexer(ga, &mockElementRotator{}, mm, bm, &mockColorExtractor{})

			idx := si.Run(tt.blocks, nil)
			require.Equal(t, tt.wantLen, idx.Len())
			if tt.wantCheck != nil {
				tt.wantCheck(t, idx)
			}
		})
	}
}
