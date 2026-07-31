package minecraft

import (
	"fmt"
	"sort"

	"mc2lua/internal/model"
)

const maxTextureLift = 1.25

type pixelAccumulator struct {
	rSum, gSum, bSum int64
	count            int
	lums             []float64
}

func (a *pixelAccumulator) add(r, g, b, alpha uint32) {
	if alpha == 0 {
		return
	}
	r8, g8, b8 := uint8(r>>8), uint8(g>>8), uint8(b>>8)
	a.rSum += int64(r8)
	a.gSum += int64(g8)
	a.bSum += int64(b8)
	a.lums = append(a.lums, a.luminance(r8, g8, b8))
	a.count++
}

func (a *pixelAccumulator) result() (model.Color, error) {
	if a.count == 0 {
		return model.DefaultColor, fmt.Errorf("no valid pixels found")
	}
	lift := a.liftFactor()
	return model.Color{
		a.clampByte(int(float64(a.rSum/int64(a.count)) * lift)),
		a.clampByte(int(float64(a.gSum/int64(a.count)) * lift)),
		a.clampByte(int(float64(a.bSum/int64(a.count)) * lift)),
	}, nil
}

func (a *pixelAccumulator) liftFactor() float64 {
	if len(a.lums) < 2 {
		return 1.0
	}

	avgLum := a.averageLuminance()
	if avgLum <= 0 {
		return 1.0
	}

	return a.clampLift(a.upperHalfAverage() / avgLum)
}

func (a *pixelAccumulator) averageLuminance() float64 {
	var total float64
	for _, l := range a.lums {
		total += l
	}
	return total / float64(len(a.lums))
}

func (a *pixelAccumulator) upperHalfAverage() float64 {
	sorted := make([]float64, len(a.lums))
	copy(sorted, a.lums)
	sort.Float64s(sorted)

	mid := len(sorted) / 2
	var total float64
	for _, l := range sorted[mid:] {
		total += l
	}
	return total / float64(len(sorted)-mid)
}

func (a *pixelAccumulator) clampLift(lift float64) float64 {
	if lift < 1.0 {
		return 1.0
	}
	if lift > maxTextureLift {
		return maxTextureLift
	}
	return lift
}

func (a *pixelAccumulator) luminance(r, g, b uint8) float64 {
	return 0.2126*float64(r) + 0.7152*float64(g) + 0.0722*float64(b)
}

func (a *pixelAccumulator) clampByte(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > 255 {
		return 255
	}
	return uint8(v)
}
