package stateful

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestPixelAccumulator_New(t *testing.T) {
	t.Parallel()

	acc := NewPixelAccumulator()
	require.NotNil(t, acc)
	require.Equal(t, 0, acc.count)
	require.Empty(t, acc.lums)
}

func TestPixelAccumulator_Result(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		pixels  [][4]uint32
		want    model.Color
		wantErr string
	}{
		{
			name:   "single pixel keeps color",
			pixels: [][4]uint32{{150, 150, 150, 255}},
			want:   model.Color{150, 150, 150},
		},
		{
			name:   "uniform pixels keep color",
			pixels: [][4]uint32{{150, 150, 150, 255}, {150, 150, 150, 255}},
			want:   model.Color{150, 150, 150},
		},
		{
			name:   "shaded pixels get lifted average",
			pixels: [][4]uint32{{100, 100, 100, 255}, {150, 150, 150, 255}},
			want:   model.Color{150, 150, 150},
		},
		{
			name:   "strong contrast lift is capped",
			pixels: [][4]uint32{{30, 30, 30, 255}, {255, 255, 255, 255}},
			want:   model.Color{177, 177, 177},
		},
		{
			name:    "no valid pixels returns error",
			wantErr: "no valid pixels found",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acc := NewPixelAccumulator()
			addPixels(acc, tt.pixels)
			got, err := acc.Result()
			if tt.wantErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

// addPixels feeds 8-bit components shifted to the 16-bit range used by image.RGBA.
func addPixels(acc *PixelAccumulator, pixels [][4]uint32) {
	for _, p := range pixels {
		acc.Add(p[0]<<8, p[1]<<8, p[2]<<8, p[3]<<8)
	}
}

func TestPixelAccumulator_Add_IgnoresTransparent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		pixels    [][4]uint32
		wantCount int
	}{
		{name: "fully transparent ignored", pixels: [][4]uint32{{1, 2, 3, 0}}, wantCount: 0},
		{name: "opaque counted", pixels: [][4]uint32{{1, 2, 3, 255}}, wantCount: 1},
		{name: "mix of transparent and opaque", pixels: [][4]uint32{{1, 2, 3, 0}, {4, 5, 6, 255}, {7, 8, 9, 0}}, wantCount: 1},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acc := NewPixelAccumulator()
			addPixels(acc, tt.pixels)
			require.Equal(t, tt.wantCount, acc.count)
		})
	}
}

func TestPixelAccumulator_LiftFactor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		pixels [][4]uint32
		want   float64
	}{
		{
			name:   "less than two pixels returns one",
			pixels: [][4]uint32{{100, 100, 100, 255}},
			want:   1.0,
		},
		{
			name:   "zero average luminance returns one",
			pixels: [][4]uint32{{0, 0, 0, 255}, {0, 0, 0, 255}},
			want:   1.0,
		},
		{
			name:   "balanced shades lift to upper half",
			pixels: [][4]uint32{{100, 100, 100, 255}, {150, 150, 150, 255}},
			want:   1.2,
		},
		{
			name:   "extreme contrast capped",
			pixels: [][4]uint32{{30, 30, 30, 255}, {255, 255, 255, 255}},
			want:   maxTextureLift,
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			acc := NewPixelAccumulator()
			addPixels(acc, tt.pixels)
			require.InDelta(t, tt.want, acc.liftFactor(), 1e-9)
		})
	}
}

func TestPixelAccumulator_ClampLift(t *testing.T) {
	t.Parallel()

	acc := NewPixelAccumulator()

	tests := []struct {
		name string
		in   float64
		want float64
	}{
		{name: "below one clamped to one", in: 0.5, want: 1.0},
		{name: "one", in: 1.0, want: 1.0},
		{name: "within range", in: 1.2, want: 1.2},
		{name: "above max clamped", in: 2.0, want: maxTextureLift},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.InDelta(t, tt.want, acc.clampLift(tt.in), 1e-9)
		})
	}
}

func TestPixelAccumulator_ClampByte(t *testing.T) {
	t.Parallel()

	acc := NewPixelAccumulator()

	tests := []struct {
		name string
		in   int
		want uint8
	}{
		{name: "negative clamped to zero", in: -5, want: 0},
		{name: "zero", in: 0, want: 0},
		{name: "mid value", in: 128, want: 128},
		{name: "max value", in: 255, want: 255},
		{name: "overflow clamped to 255", in: 300, want: 255},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, acc.clampByte(tt.in))
		})
	}
}

func TestPixelAccumulator_Luminance(t *testing.T) {
	t.Parallel()

	acc := NewPixelAccumulator()

	tests := []struct {
		name string
		in   [3]uint8
		want float64
	}{
		{name: "black", in: [3]uint8{0, 0, 0}, want: 0},
		{name: "white", in: [3]uint8{255, 255, 255}, want: 255},
		{name: "gray", in: [3]uint8{150, 150, 150}, want: 150},
		{name: "pure red", in: [3]uint8{255, 0, 0}, want: 0.2126 * 255},
		{name: "pure green", in: [3]uint8{0, 255, 0}, want: 0.7152 * 255},
		{name: "pure blue", in: [3]uint8{0, 0, 255}, want: 0.0722 * 255},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.InDelta(t, tt.want, acc.luminance(tt.in[0], tt.in[1], tt.in[2]), 1e-9)
		})
	}
}
