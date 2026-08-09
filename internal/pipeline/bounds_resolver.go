package pipeline

import (
	"mc2lua/internal/model"
)

type BoundsResolver struct{}

func NewBoundsResolver() *BoundsResolver {
	return &BoundsResolver{}
}

func (svc *BoundsResolver) Run(blocks []model.RawBlock) model.Bounds {
	if len(blocks) == 0 {
		return model.Bounds{}
	}

	bounds := model.Bounds{
		XMin: blocks[0].X, XMax: blocks[0].X,
		YMin: blocks[0].Y, YMax: blocks[0].Y,
		ZMin: blocks[0].Z, ZMax: blocks[0].Z,
	}
	for _, b := range blocks[1:] {
		if b.X < bounds.XMin {
			bounds.XMin = b.X
		}
		if b.X > bounds.XMax {
			bounds.XMax = b.X
		}
		if b.Y < bounds.YMin {
			bounds.YMin = b.Y
		}
		if b.Y > bounds.YMax {
			bounds.YMax = b.Y
		}
		if b.Z < bounds.ZMin {
			bounds.ZMin = b.Z
		}
		if b.Z > bounds.ZMax {
			bounds.ZMax = b.Z
		}
	}
	return bounds
}
