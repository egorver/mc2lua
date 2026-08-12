package minecraft

import (
	"bytes"
	"fmt"
	"image"
	"image/png"

	"mc2lua/internal/model"
)

const (
	plainsTemperature = 0.8
	plainsDownfall    = 0.4
	colormapRange     = 255.0
)

type ColormapResolver struct {
	fs fsApi
}

func NewColormapResolver(fs fsApi) *ColormapResolver {
	return &ColormapResolver{fs: fs}
}

func (svc *ColormapResolver) Run(nsRoots map[string][]string) (*model.Colormap, error) {
	grass, err := svc.sample(nsRoots, GrassColormapFile)
	if err != nil {
		return nil, err
	}
	foliage, err := svc.sample(nsRoots, FoliageColormapFile)
	if err != nil {
		return nil, err
	}
	return &model.Colormap{Grass: grass, Foliage: foliage}, nil
}

func (svc *ColormapResolver) sample(nsRoots map[string][]string, fileName string) (model.Color, error) {
	for _, roots := range nsRoots {
		for _, root := range roots {
			data, err := svc.fs.ReadFile(root + ColormapDirPath + fileName)
			if err != nil {
				continue
			}
			img, err := png.Decode(bytes.NewReader(data))
			if err != nil {
				continue
			}
			return svc.sampleColormap(img, plainsTemperature, plainsDownfall), nil
		}
	}
	return model.Color{}, fmt.Errorf("colormap %s: not found in any namespace root", fileName)
}

func (svc *ColormapResolver) sampleColormap(img image.Image, temperature, downfall float64) model.Color {
	temp := svc.clamp01(temperature)
	hum := svc.clamp01(downfall) * temp

	x := int((1.0 - temp) * colormapRange)
	y := int((1.0 - hum) * colormapRange)

	r, g, b, _ := img.At(x, y).RGBA()
	return model.Color{
		uint8(r >> 8),
		uint8(g >> 8),
		uint8(b >> 8),
	}
}

func (svc *ColormapResolver) clamp01(v float64) float64 {
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}
