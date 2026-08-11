package pipeline

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestBoundsResolver_New(t *testing.T) {
	t.Parallel()

	require.NotNil(t, NewBoundsResolver())
}

func TestBoundsResolver_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		blocks []model.RawBlock
		want   model.Bounds
	}{
		{
			name:   "empty input",
			blocks: nil,
			want:   model.Bounds{},
		},
		{
			name: "single block",
			blocks: []model.RawBlock{
				{X: 5, Y: 10, Z: 15},
			},
			want: model.Bounds{
				XMin: 5, XMax: 5,
				YMin: 10, YMax: 10,
				ZMin: 15, ZMax: 15,
			},
		},
		{
			name: "negative coordinates",
			blocks: []model.RawBlock{
				{X: -100, Y: -64, Z: -200},
			},
			want: model.Bounds{
				XMin: -100, XMax: -100,
				YMin: -64, YMax: -64,
				ZMin: -200, ZMax: -200,
			},
		},
		{
			name: "scattered blocks",
			blocks: []model.RawBlock{
				{X: 0, Y: 0, Z: 0},
				{X: -50, Y: 255, Z: 100},
				{X: 30, Y: -10, Z: 20},
			},
			want: model.Bounds{
				XMin: -50, XMax: 30,
				YMin: -10, YMax: 255,
				ZMin: 0, ZMax: 100,
			},
		},
		{
			name: "z min from later block",
			blocks: []model.RawBlock{
				{X: 0, Y: 0, Z: 10},
				{X: 5, Y: 5, Z: -5},
			},
			want: model.Bounds{
				XMin: 0, XMax: 5,
				YMin: 0, YMax: 5,
				ZMin: -5, ZMax: 10,
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			svc := NewBoundsResolver()
			require.Equal(t, tt.want, svc.Run(tt.blocks))
		})
	}
}
