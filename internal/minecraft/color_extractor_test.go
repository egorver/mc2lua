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

func makePNG(w, h int, px func(x, y int) color.NRGBA) []byte {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, px(x, y))
		}
	}
	var buf bytes.Buffer
	png.Encode(&buf, img)
	return buf.Bytes()
}

func setupColorExtractorFS(t *testing.T) (*runtime.FSMock, *ColorExtractor, map[string][]string) {
	t.Helper()
	fs := runtime.NewFSMock()
	extractor := NewColorExtractor(fs)
	return fs, extractor, map[string][]string{"minecraft": {"assets/minecraft"}}
}

func sample(path string) []TextureSample {
	return []TextureSample{{TextureVar: path, UV: [4]float64{0, 0, 16, 16}}}
}

func TestColorExtractor_UniformTexture(t *testing.T) {
	t.Parallel()

	fs, extractor, roots := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/uniform.png",
		makePNG(16, 16, func(x, y int) color.NRGBA { return color.NRGBA{150, 150, 150, 255} }), 0644)

	got, err := extractor.Run(sample("block/uniform"), roots, "minecraft:test_block")
	require.NoError(t, err)
	require.Equal(t, model.Color{150, 150, 150}, got)
}

func TestColorExtractor_ShadedTexture(t *testing.T) {
	t.Parallel()

	fs, extractor, roots := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/shaded.png",
		makePNG(16, 16, func(x, y int) color.NRGBA {
			if x < 8 {
				return color.NRGBA{100, 100, 100, 255}
			}
			return color.NRGBA{150, 150, 150, 255}
		}), 0644)

	got, err := extractor.Run(sample("block/shaded"), roots, "minecraft:test_block")
	require.NoError(t, err)
	require.Equal(t, model.Color{150, 150, 150}, got)
}

func TestColorExtractor_TransparentPixelsIgnored(t *testing.T) {
	t.Parallel()

	fs, extractor, roots := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/transparent.png",
		makePNG(16, 16, func(x, y int) color.NRGBA {
			if x < 8 {
				return color.NRGBA{0, 0, 0, 0}
			}
			return color.NRGBA{120, 120, 120, 255}
		}), 0644)

	got, err := extractor.Run(sample("block/transparent"), roots, "minecraft:test_block")
	require.NoError(t, err)
	require.Equal(t, model.Color{120, 120, 120}, got)
}

func TestColorExtractor_LiftCapped(t *testing.T) {
	t.Parallel()

	fs, extractor, roots := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/capped.png",
		makePNG(16, 16, func(x, y int) color.NRGBA {
			if x < 8 {
				return color.NRGBA{30, 30, 30, 255}
			}
			return color.NRGBA{255, 255, 255, 255}
		}), 0644)

	got, err := extractor.Run(sample("block/capped"), roots, "minecraft:test_block")
	require.NoError(t, err)
	require.Equal(t, model.Color{177, 177, 177}, got)
}

func TestColorExtractor_NoValidPixels(t *testing.T) {
	t.Parallel()

	fs, extractor, roots := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/empty.png",
		makePNG(16, 16, func(x, y int) color.NRGBA { return color.NRGBA{0, 0, 0, 0} }), 0644)

	_, err := extractor.Run(sample("block/empty"), roots, "minecraft:test_block")
	require.Error(t, err)
}
