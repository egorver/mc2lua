package app

import (
	"math"
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/pipeline"

	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Parallel()

	a := New()
	require.NotNil(t, a)
}

func TestBuildDeps(t *testing.T) {
	t.Parallel()

	r := buildDeps()
	require.NotNil(t, r)
}

func TestApp_Run_WithNonExistentInput_ReturnsError(t *testing.T) {
	t.Parallel()

	a := New()
	err := a.Run(AppConfig{
		Input: t.TempDir() + "nonexistent",
	})
	require.Error(t, err)
}

func TestBuildConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cfg  AppConfig
		want pipeline.RunConfig
	}{
		{
			name: "zero values",
			cfg:  AppConfig{},
			want: pipeline.RunConfig{
				Bounds: model.Bounds{},
			},
		},
		{
			name: "all fields set",
			cfg: AppConfig{
				Input:    "region",
				Output:   "out.lua",
				Scale:    4,
				NoOffset: true,
				XMin:     -10, XMax: 10,
				YMin: 0, YMax: 255,
				ZMin: -20, ZMax: 20,
			},
			want: pipeline.RunConfig{
				Input:    "region",
				Output:   "out.lua",
				Scale:    4,
				NoOffset: true,
				Bounds: model.Bounds{
					XMin: -10, XMax: 10,
					YMin: 0, YMax: 255,
					ZMin: -20, ZMax: 20,
				},
			},
		},
		{
			name: "negative bounds",
			cfg: AppConfig{
				XMin: -100, XMax: -50,
				YMin: -64, YMax: 0,
				ZMin: -200, ZMax: -100,
			},
			want: pipeline.RunConfig{
				Bounds: model.Bounds{
					XMin: -100, XMax: -50,
					YMin: -64, YMax: 0,
					ZMin: -200, ZMax: -100,
				},
			},
		},
		{
			name: "extreme int bounds",
			cfg: AppConfig{
				XMin: math.MinInt32, XMax: math.MaxInt32,
				YMin: math.MinInt32, YMax: math.MaxInt32,
				ZMin: math.MinInt32, ZMax: math.MaxInt32,
			},
			want: pipeline.RunConfig{
				Bounds: model.Bounds{
					XMin: math.MinInt32, XMax: math.MaxInt32,
					YMin: math.MinInt32, YMax: math.MaxInt32,
					ZMin: math.MinInt32, ZMax: math.MaxInt32,
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := buildConfig(tt.cfg)
			require.Equal(t, tt.want, got)
		})
	}
}
