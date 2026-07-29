package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBounds_ZeroValues(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name               string
		bounds             Bounds
		wantXMin, wantXMax int
		wantYMin, wantYMax int
		wantZMin, wantZMax int
	}{
		{
			name:   "zero value",
			bounds: Bounds{},
		},
		{
			name: "set values",
			bounds: Bounds{
				XMin: -10, XMax: 10,
				YMin: 0, YMax: 255,
				ZMin: -100, ZMax: 100,
			},
			wantXMin: -10, wantXMax: 10,
			wantYMin: 0, wantYMax: 255,
			wantZMin: -100, wantZMax: 100,
		},
		{
			name: "negative bounds",
			bounds: Bounds{
				XMin: -100, XMax: -50,
				YMin: -64, YMax: -32,
				ZMin: -200, ZMax: -150,
			},
			wantXMin: -100, wantXMax: -50,
			wantYMin: -64, wantYMax: -32,
			wantZMin: -200, wantZMax: -150,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantXMin, tt.bounds.XMin)
			require.Equal(t, tt.wantXMax, tt.bounds.XMax)
			require.Equal(t, tt.wantYMin, tt.bounds.YMin)
			require.Equal(t, tt.wantYMax, tt.bounds.YMax)
			require.Equal(t, tt.wantZMin, tt.bounds.ZMin)
			require.Equal(t, tt.wantZMax, tt.bounds.ZMax)
		})
	}
}
