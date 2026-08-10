package pipeline

import (
	"math"

	"mc2lua/internal/model"
)

type CuboidHelper struct{}

func NewCuboidHelper() *CuboidHelper {
	return &CuboidHelper{}
}

func (svc *CuboidHelper) Center(x, size int) float64 {
	return float64(x) + float64(size-1)/2.0
}

func (svc *CuboidHelper) MinCorner(c model.Cuboid) (x, y, z int) {
	x = int(math.Floor(c.X - float64(c.Width-1)/2.0))
	y = int(math.Floor(c.Y - float64(c.Height-1)/2.0))
	z = int(math.Floor(c.Z - float64(c.Depth-1)/2.0))
	return x, y, z
}
