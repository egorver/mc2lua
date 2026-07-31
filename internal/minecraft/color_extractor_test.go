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

func TestColorExtractor_SplitBlockID(t *testing.T) {
	t.Parallel()

	svc := NewColorExtractor(runtime.NewFSMock())

	tests := []struct {
		name        string
		id          string
		wantNS      string
		wantBlockID string
	}{
		{name: "with namespace", id: "minecraft:stone", wantNS: "minecraft", wantBlockID: "stone"},
		{name: "custom namespace", id: "cobblemon:ore", wantNS: "cobblemon", wantBlockID: "ore"},
		{name: "without namespace", id: "stone", wantNS: "minecraft", wantBlockID: "stone"},
		{name: "empty id", id: "", wantNS: "minecraft", wantBlockID: ""},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ns, blockID := svc.splitBlockID(tt.id)
			require.Equal(t, tt.wantNS, ns)
			require.Equal(t, tt.wantBlockID, blockID)
		})
	}
}

func TestColorExtractor_Clamp(t *testing.T) {
	t.Parallel()

	svc := NewColorExtractor(runtime.NewFSMock())

	tests := []struct {
		name string
		v    int
		low  int
		high int
		want int
	}{
		{name: "below low", v: -5, low: 0, high: 10, want: 0},
		{name: "at low", v: 0, low: 0, high: 10, want: 0},
		{name: "inside range", v: 5, low: 0, high: 10, want: 5},
		{name: "at high", v: 10, low: 0, high: 10, want: 10},
		{name: "above high", v: 15, low: 0, high: 10, want: 10},
		{name: "negative range", v: -8, low: -10, high: -5, want: -8},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, svc.clamp(tt.v, tt.low, tt.high))
		})
	}
}

func TestColorExtractor_BuildTexturePath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		textureVar string
		nsRoots    map[string][]string
		blockID    string
		files      map[string][]byte
		want       string
	}{
		{
			name:       "texture var with namespace and existing root",
			textureVar: "minecraft:block/stone",
			nsRoots:    map[string][]string{"minecraft": {"assets/minecraft"}},
			files:      map[string][]byte{"assets/minecraft/textures/block/stone.png": {}},
			want:       "assets/minecraft/textures/block/stone.png",
		},
		{
			name:       "texture var without namespace uses block namespace",
			textureVar: "block/stone",
			nsRoots:    map[string][]string{"minecraft": {"assets/minecraft"}},
			blockID:    "minecraft:stone",
			files:      map[string][]byte{"assets/minecraft/textures/block/stone.png": {}},
			want:       "assets/minecraft/textures/block/stone.png",
		},
		{
			name:       "namespace missing in roots",
			textureVar: "mod:block/ore",
			nsRoots:    map[string][]string{"minecraft": {"assets/minecraft"}},
			want:       "",
		},
		{
			name:       "empty roots for namespace",
			textureVar: "minecraft:block/stone",
			nsRoots:    map[string][]string{"minecraft": {}},
			want:       "",
		},
		{
			name:       "first root missing file falls through to second",
			textureVar: "minecraft:block/stone",
			nsRoots:    map[string][]string{"minecraft": {"assets/first", "assets/second"}},
			files:      map[string][]byte{"assets/second/textures/block/stone.png": {}},
			want:       "assets/second/textures/block/stone.png",
		},
		{
			name:       "no root contains the texture",
			textureVar: "minecraft:block/stone",
			nsRoots:    map[string][]string{"minecraft": {"assets/first", "assets/second"}},
			files:      map[string][]byte{"assets/first/textures/block/dirt.png": {}},
			want:       "",
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			fs := runtime.NewFSMock()
			for path, data := range tt.files {
				fs.AddFile(path, data, 0644)
			}
			svc := NewColorExtractor(fs)

			got := svc.buildTexturePath(tt.textureVar, tt.nsRoots, tt.blockID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestColorExtractor_DecodeTexture_InvalidPNG(t *testing.T) {
	t.Parallel()

	fs := runtime.NewFSMock()
	fs.AddFile("assets/minecraft/textures/block/broken.png", []byte("not a png"), 0644)
	svc := NewColorExtractor(fs)

	_, err := svc.decodeTexture("assets/minecraft/textures/block/broken.png")
	require.Error(t, err)
}

func TestColorExtractor_DecodeTexture_MissingFile(t *testing.T) {
	t.Parallel()

	svc := NewColorExtractor(runtime.NewFSMock())

	_, err := svc.decodeTexture("assets/minecraft/textures/block/missing.png")
	require.Error(t, err)
}

func TestColorExtractor_Run_MissingTextures(t *testing.T) {
	t.Parallel()

	_, extractor, roots := setupColorExtractorFS(t)

	_, err := extractor.Run(sample("block/missing"), roots, "minecraft:test_block")
	require.Error(t, err)
}

func TestColorExtractor_Run_BrokenTextureSkipped(t *testing.T) {
	t.Parallel()

	fs, extractor, roots := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/broken.png", []byte("not a png"), 0644)

	_, err := extractor.Run(sample("block/broken"), roots, "minecraft:test_block")
	require.Error(t, err)
}

func TestColorExtractor_Run_UnknownNamespaceSkipped(t *testing.T) {
	t.Parallel()

	fs, extractor, _ := setupColorExtractorFS(t)
	fs.AddFile("assets/minecraft/textures/block/present.png",
		makePNG(16, 16, func(x, y int) color.NRGBA { return color.NRGBA{120, 120, 120, 255} }), 0644)

	_, err := extractor.Run(sample("block/present"), map[string][]string{"other": {"assets/minecraft"}}, "minecraft:test_block")
	require.Error(t, err)
}
