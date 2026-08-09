package minecraft

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestGridAnalyzer_Run(t *testing.T) {
	t.Parallel()

	svc := NewGridAnalyzer()

	tests := []struct {
		name     string
		elements []model.StyledElement
		want     model.GridAlignment
	}{
		{
			name:     "empty elements",
			elements: nil,
			want:     model.GridNotAligned,
		},
		{
			name: "full block",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
			},
			want: model.GridFullBlock,
		},
		{
			name: "multiple full blocks",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
			},
			want: model.GridFullBlock,
		},
		{
			name: "half slab",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}, Shade: true},
			},
			want: model.GridSubBlock,
		},
		{
			name: "full block plus partial element",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Shade: true},
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}, Shade: true},
			},
			want: model.GridSubBlock,
		},
		{
			name: "stairs two elements",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}, Shade: true},
				{From: model.Vector3{0, 8, 0}, To: model.Vector3{16, 16, 8}, Shade: true},
			},
			want: model.GridSubBlock,
		},
		{
			name: "has rotation",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 16, 16}, Rotation: &model.ElementRotation{}, Shade: true},
			},
			want: model.GridNotAligned,
		},
		{
			name: "shade false",
			elements: []model.StyledElement{
				{From: model.Vector3{0, 0, 0}, To: model.Vector3{16, 8, 16}, Shade: false},
			},
			want: model.GridNotAligned,
		},
		{
			name: "not aligned to sub-grid",
			elements: []model.StyledElement{
				{From: model.Vector3{1, 0, 0}, To: model.Vector3{15, 8, 16}, Shade: true},
			},
			want: model.GridNotAligned,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.Run(tt.elements)
			require.Equal(t, tt.want, got)
		})
	}
}
