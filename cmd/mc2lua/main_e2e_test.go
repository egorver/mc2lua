package main

import (
	"bytes"
	"compress/zlib"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"

	"mc2lua/internal/app"

	"github.com/Tnze/go-mc/nbt"
	"github.com/Tnze/go-mc/save/region"
	"github.com/stretchr/testify/require"
)

func TestStress_E2E(t *testing.T) {
	dir := t.TempDir()

	regionDir := filepath.Join(dir, "region")
	require.NoError(t, os.MkdirAll(regionDir, 0o755))

	chunkData, err := encodeChunkNBT(map[string]interface{}{
		"Status": "full",
		"sections": []interface{}{
			map[string]interface{}{
				"Y": int8(0),
				"block_states": map[string]interface{}{
					"palette": []interface{}{map[string]interface{}{"Name": "minecraft:stone"}},
					"data":    []int64{},
				},
			},
		},
	})
	require.NoError(t, err)
	writeRegionFile(t, regionDir, "r.0.0.mca", map[[2]int][]byte{{0, 0}: chunkData})

	assetsDir := filepath.Join(dir, "assets", "minecraft", "minecraft")
	require.NoError(t, os.MkdirAll(filepath.Join(assetsDir, "blockstates"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(assetsDir, "models", "block"), 0o755))
	require.NoError(t, os.MkdirAll(filepath.Join(assetsDir, "textures", "block"), 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(assetsDir, "blockstates", "stone.json"),
		[]byte(`{"variants":{"":{"model":"block/stone"}}}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(assetsDir, "models", "block", "stone.json"),
		[]byte(`{"elements":[{"from":[0,0,0],"to":[16,16,16],"shade":true}],"textures":{"particle":"block/stone"}}`), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(assetsDir, "textures", "block", "stone.png"),
		makePNG(t, 16, 16, func(x, y int) color.NRGBA { return color.NRGBA{125, 125, 125, 255} }), 0o644))

	configDir := filepath.Join(dir, "config")
	require.NoError(t, os.MkdirAll(configDir, 0o755))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "materials.yaml"),
		[]byte("mappings:\n  stone: Slate\nbrightness:\n  Slate: 1.0\n"), 0o644))
	require.NoError(t, os.WriteFile(filepath.Join(configDir, "parts.yaml"),
		[]byte("parts:\n  minecraft:stone:\n    color: [125, 125, 125]\n"), 0o644))

	outputPath := filepath.Join(dir, "output.lua")

	var stdout, stderr bytes.Buffer
	code := run([]string{
		"-input", regionDir,
		"-assets", filepath.Join(dir, "assets"),
		"-output", outputPath,
		"-parts-template", filepath.Join(dir, "template.yaml"),
		"-config", configDir,
	}, &stdout, &stderr, func() appRunner { return app.New() })

	require.Zero(t, code, "stderr: %s", stderr.String())
	require.FileExists(t, outputPath)
}

func encodeChunkNBT(chunk map[string]interface{}) ([]byte, error) {
	raw, err := nbt.Marshal(chunk)
	if err != nil {
		return nil, err
	}
	var buf bytes.Buffer
	buf.WriteByte(2)
	w := zlib.NewWriter(&buf)
	if _, err := w.Write(raw); err != nil {
		return nil, err
	}
	if err := w.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func writeRegionFile(t *testing.T, dir, name string, chunks map[[2]int][]byte) string {
	t.Helper()

	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	require.NoError(t, err)

	r, err := region.CreateWriter(f)
	require.NoError(t, err)

	for coords, data := range chunks {
		require.NoError(t, r.WriteSector(coords[0], coords[1], data))
	}
	require.NoError(t, r.Close())

	return path
}

func makePNG(t testing.TB, w, h int, px func(x, y int) color.NRGBA) []byte {
	t.Helper()
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			img.SetNRGBA(x, y, px(x, y))
		}
	}
	var buf bytes.Buffer
	require.NoError(t, png.Encode(&buf, img))
	return buf.Bytes()
}
