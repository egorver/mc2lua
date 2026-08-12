package minecraft

import (
	"bytes"
	"image"
	"image/png"
	"strings"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type ColorExtractor struct {
	fs fsApi
}

const rgbaChannelShift = 8

func NewColorExtractor(fs fsApi) *ColorExtractor {
	return &ColorExtractor{fs: fs}
}

type TextureSample struct {
	TextureVar string
	UV         [4]float64
	Tint       *model.Color
}

func (svc *ColorExtractor) Run(samples []TextureSample, nsRoots map[string][]string, blockID string) (model.Color, error) {
	acc := stateful.NewPixelAccumulator()

	for _, sample := range samples {
		filePath := svc.buildTexturePath(sample.TextureVar, nsRoots, blockID)
		if filePath == "" {
			continue
		}

		img, err := svc.decodeTexture(filePath)
		if err != nil {
			continue
		}

		svc.accumulateRegion(img, sample.UV, sample.Tint, acc)
	}

	return acc.Result()
}

func (svc *ColorExtractor) decodeTexture(filePath string) (image.Image, error) {
	data, err := svc.fs.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	return png.Decode(bytes.NewReader(data))
}

func (svc *ColorExtractor) accumulateRegion(img image.Image, uv [4]float64, tint *model.Color, acc *stateful.PixelAccumulator) {
	bounds := img.Bounds()
	w, h := bounds.Dx(), bounds.Dy()

	uMin := svc.clamp(int(uv[0]*float64(w)/TexturePixelSize), 0, w-1)
	uMax := svc.clamp(int(uv[2]*float64(w)/TexturePixelSize), uMin+1, w)
	vMin := svc.clamp(int(uv[1]*float64(h)/TexturePixelSize), 0, h-1)
	vMax := svc.clamp(int(uv[3]*float64(h)/TexturePixelSize), vMin+1, h)

	for y := vMin; y < vMax; y++ {
		for x := uMin; x < uMax; x++ {
			r, g, b, a := img.At(x, y).RGBA()
			r8, g8, b8 := uint8(r>>rgbaChannelShift), uint8(g>>rgbaChannelShift), uint8(b>>rgbaChannelShift)
			if tint != nil {
				r8, g8, b8 = svc.tintPixel(r8, g8, b8, *tint)
			}
			acc.Add(uint32(r8)<<rgbaChannelShift, uint32(g8)<<rgbaChannelShift, uint32(b8)<<rgbaChannelShift, a)
		}
	}
}

func (svc *ColorExtractor) tintPixel(r, g, b uint8, tint model.Color) (uint8, uint8, uint8) {
	return svc.tintChannel(r, tint[0]), svc.tintChannel(g, tint[1]), svc.tintChannel(b, tint[2])
}

func (svc *ColorExtractor) tintChannel(v, tint uint8) uint8 {
	if tint == 0 {
		return 0
	}
	return uint8((int(v)*int(tint) + 127) / 255)
}

func (svc *ColorExtractor) buildTexturePath(textureVar string, nsRoots map[string][]string, blockID string) string {
	var ns, relPath string
	if strings.Contains(textureVar, NamespaceSeparator) {
		parts := strings.SplitN(textureVar, NamespaceSeparator, 2)
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
		candidate := root + TextureDirPath + relPath + PNGExtension
		if _, err := svc.fs.ReadFile(candidate); err == nil {
			return candidate
		}
	}

	return ""
}

func (svc *ColorExtractor) splitBlockID(id string) (namespace, blockID string) {
	if parts := strings.SplitN(id, NamespaceSeparator, 2); len(parts) == 2 {
		return parts[0], parts[1]
	}
	return DefaultNamespace, id
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
