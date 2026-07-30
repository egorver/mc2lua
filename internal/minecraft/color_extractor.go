package minecraft

import (
	"bytes"
	"fmt"
	"image"
	"image/png"
	"strings"

	"mc2lua/internal/model"
)

type ColorExtractor struct {
	fs fsApi
}

func NewColorExtractor(fs fsApi) *ColorExtractor {
	return &ColorExtractor{fs: fs}
}

type TextureSample struct {
	TextureVar string
	UV         [4]float64
}

type pixelAccumulator struct {
	rSum, gSum, bSum int64
	count            int
}

func (a *pixelAccumulator) add(r, g, b, alpha uint32) {
	if alpha == 0 {
		return
	}
	a.rSum += int64(r >> 8)
	a.gSum += int64(g >> 8)
	a.bSum += int64(b >> 8)
	a.count++
}

func (a *pixelAccumulator) result() (model.Color, error) {
	if a.count == 0 {
		return model.DefaultColor, fmt.Errorf("no valid pixels found")
	}
	return model.Color{
		uint8(a.rSum / int64(a.count)),
		uint8(a.gSum / int64(a.count)),
		uint8(a.bSum / int64(a.count)),
	}, nil
}

func (svc *ColorExtractor) Run(samples []TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
	var acc pixelAccumulator

	for _, sample := range samples {
		filePath := svc.buildTexturePath(sample.TextureVar, nsRoots, blockID)
		if filePath == "" {
			continue
		}

		img, err := svc.decodeTexture(filePath)
		if err != nil {
			continue
		}

		svc.accumulateRegion(img, sample.UV, &acc)
	}

	return acc.result()
}

func (svc *ColorExtractor) decodeTexture(filePath string) (image.Image, error) {
	data, err := svc.fs.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

func (svc *ColorExtractor) accumulateRegion(img image.Image, uv [4]float64, acc *pixelAccumulator) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	uMin := svc.clamp(int(uv[0]*float64(w)/16), 0, w-1)
	uMax := svc.clamp(int(uv[2]*float64(w)/16), uMin+1, w)
	vMin := svc.clamp(int(uv[1]*float64(h)/16), 0, h-1)
	vMax := svc.clamp(int(uv[3]*float64(h)/16), vMin+1, h)

	for y := vMin; y < vMax; y++ {
		for x := uMin; x < uMax; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			acc.add(r, g, b, a)
		}
	}
}

func (svc *ColorExtractor) buildTexturePath(textureVar string, nsRoots map[string][]string, blockID string) string {
	var ns, relPath string
	if strings.Contains(textureVar, ":") {
		parts := strings.SplitN(textureVar, ":", 2)
		ns, relPath = parts[0], parts[1]
	} else {
		ns, _ = svc.splitBlockID(blockID)
		relPath = textureVar
	}

	roots, ok := nsRoots[ns]
	if !ok || len(roots) == 0 {
		return ""
	}

	for _, root := range roots {
		candidate := root + "/textures/" + relPath + ".png"
		if _, err := svc.fs.ReadFile(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func (svc *ColorExtractor) splitBlockID(id string) (namespace, blockID string) {
	if parts := strings.SplitN(id, ":", 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return "minecraft", id
}

func (svc *ColorExtractor) clamp(v, low, high int) int {
	if v < low {
		return low
	}
	if v > high {
		return high
	}
	return v
}
