package minecraft

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/runtime"

	"github.com/stretchr/testify/require"
)

func makeColormap(t testing.TB, encode func(x, y int) color.NRGBA) image.Image {
	t.Helper()
	data := makePNG(t, 256, 256, encode)
	img, err := png.Decode(bytes.NewReader(data))
	require.NoError(t, err)
	return img
}

func makeColormapPNG(t testing.TB, encode func(x, y int) color.NRGBA) []byte {
	t.Helper()
	return makePNG(t, 256, 256, encode)
}

func TestSampleColormap(t *testing.T) {
	t.Parallel()

	svc := NewColormapResolver(runtime.NewFSMock())

	img := makeColormap(t, func(x, y int) color.NRGBA {
		return color.NRGBA{uint8(x), uint8(y), 0, 255}
	})

	tests := []struct {
		name        string
		temperature float64
		downfall    float64
		want        model.Color
	}{
		{name: "plains climate", temperature: 0.8, downfall: 0.4, want: model.Color{50, 173, 0}},
		{name: "cold and dry samples far corner", temperature: 0, downfall: 0, want: model.Color{255, 255, 0}},
		{name: "hot and wet samples lower corner", temperature: 1, downfall: 1, want: model.Color{0, 0, 0}},
		{name: "below zero temperature clamps", temperature: -0.5, downfall: 0.5, want: model.Color{255, 255, 0}},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := svc.sampleColormap(img, tt.temperature, tt.downfall)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestColormapResolver_PlainsGolden(t *testing.T) {
	t.Parallel()

	svc := NewColormapResolver(runtime.NewFSMock())

	grassImg := makeColormap(t, func(x, y int) color.NRGBA {
		if x == 50 && y == 173 {
			return color.NRGBA{145, 189, 89, 255}
		}
		return color.NRGBA{0, 0, 0, 255}
	})
	require.Equal(t, model.Color{145, 189, 89}, svc.sampleColormap(grassImg, plainsTemperature, plainsDownfall))

	foliageImg := makeColormap(t, func(x, y int) color.NRGBA {
		if x == 50 && y == 173 {
			return color.NRGBA{119, 171, 47, 255}
		}
		return color.NRGBA{0, 0, 0, 255}
	})
	require.Equal(t, model.Color{119, 171, 47}, svc.sampleColormap(foliageImg, plainsTemperature, plainsDownfall))
}

func TestColormapResolver_Run(t *testing.T) {
	t.Parallel()

	fs := runtime.NewFSMock()
	fs.AddFile("assets/minecraft/textures/colormap/grass.png",
		makeColormapPNG(t, func(x, y int) color.NRGBA {
			return color.NRGBA{uint8(x), uint8(y), 0, 255}
		}), 0644)
	fs.AddFile("assets/minecraft/textures/colormap/foliage.png",
		makeColormapPNG(t, func(x, y int) color.NRGBA {
			return color.NRGBA{0, uint8(y), uint8(x), 255}
		}), 0644)

	svc := NewColormapResolver(fs)
	cm, err := svc.Run(map[string][]string{"minecraft": {"assets/minecraft"}})
	require.NoError(t, err)
	require.Equal(t, model.Color{50, 173, 0}, cm.Grass)
	require.Equal(t, model.Color{0, 173, 50}, cm.Foliage)
}

func TestColormapResolver_Run_MissingColormap(t *testing.T) {
	t.Parallel()

	fs := runtime.NewFSMock()
	fs.AddFile("assets/minecraft/textures/colormap/foliage.png",
		makeColormapPNG(t, func(x, y int) color.NRGBA { return color.NRGBA{0, 0, 0, 255} }), 0644)

	svc := NewColormapResolver(fs)
	_, err := svc.Run(map[string][]string{"minecraft": {"assets/minecraft"}})
	require.ErrorContains(t, err, "grass.png")
}

func TestColormapResolver_Run_SecondRootWins(t *testing.T) {
	t.Parallel()

	fs := runtime.NewFSMock()
	fs.AddFile("assets/mod2/textures/colormap/grass.png",
		makeColormapPNG(t, func(x, y int) color.NRGBA {
			return color.NRGBA{uint8(x), uint8(y), 0, 255}
		}), 0644)
	fs.AddFile("assets/mod2/textures/colormap/foliage.png",
		makeColormapPNG(t, func(x, y int) color.NRGBA {
			return color.NRGBA{0, uint8(y), uint8(x), 255}
		}), 0644)

	svc := NewColormapResolver(fs)
	cm, err := svc.Run(map[string][]string{"minecraft": {"assets/mod1", "assets/mod2"}})
	require.NoError(t, err)
	require.Equal(t, model.Color{50, 173, 0}, cm.Grass)
	require.Equal(t, model.Color{0, 173, 50}, cm.Foliage)
}
