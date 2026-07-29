package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

type mockModelAnalyzer struct {
	runFn func(elements []model.ModelElement) bool
}

func (m *mockModelAnalyzer) Run(elements []model.ModelElement) bool {
	if m.runFn != nil {
		return m.runFn(elements)
	}
	return false
}

func TestStyleIndexer_New(t *testing.T) {
	t.Parallel()

	si := NewStyleIndexer(&mockModelAnalyzer{})
	require.NotNil(t, si)
}

func TestStyleIndexer_Run(t *testing.T) {
	t.Parallel()

	fullBlock := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true}
	halfBlock := model.ModelElement{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}}

	tests := []struct {
		name      string
		blocks    []model.ResolvedBlock
		analyzer  func([]model.ModelElement) bool
		wantLen   int
		wantCheck func(t *testing.T, idx *model.StyleIndex)
	}{
		{
			name:    "empty input",
			blocks:  nil,
			wantLen: 0,
		},
		{
			name: "single full block",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone", Elements: []model.ModelElement{fullBlock}},
			},
			analyzer: func(elements []model.ModelElement) bool { return true },
			wantLen: 1,
			wantCheck: func(t *testing.T, idx *model.StyleIndex) {
				b, ok := idx.Get("minecraft:stone", "")
				require.True(t, ok)
				require.True(t, b.IsFullBlock)
				require.Len(t, b.Elements, 1)
			},
		},
		{
			name: "non-full block with analyzer false",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:oak_fence", PropsKey: "water=true", Elements: []model.ModelElement{halfBlock}},
			},
			analyzer: func(elements []model.ModelElement) bool { return false },
			wantLen: 1,
			wantCheck: func(t *testing.T, idx *model.StyleIndex) {
				b, ok := idx.Get("minecraft:oak_fence", "water=true")
				require.True(t, ok)
				require.False(t, b.IsFullBlock)
				require.Len(t, b.Elements, 1)
			},
		},
		{
			name: "multiple blocks",
			blocks: []model.ResolvedBlock{
				{ID: "minecraft:stone", Elements: []model.ModelElement{fullBlock}},
				{ID: "minecraft:dirt", Elements: []model.ModelElement{fullBlock}},
			},
			analyzer: func(elements []model.ModelElement) bool { return true },
			wantLen: 2,
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
			analyzer: func(elements []model.ModelElement) bool { return true },
			wantLen: 1,
			wantCheck: func(t *testing.T, idx *model.StyleIndex) {
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
			ma := &mockModelAnalyzer{runFn: tt.analyzer}
			si := NewStyleIndexer(ma)

			idx := si.Run(tt.blocks)
			require.Equal(t, tt.wantLen, idx.Len())
			if tt.wantCheck != nil {
				tt.wantCheck(t, idx)
			}
		})
	}
}
