package pipeline

import (
	"math"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type FaceCuller struct {
	propsKeyBuilder generatorPropsKeyBuilder
}

func NewFaceCuller(pkb generatorPropsKeyBuilder) *FaceCuller {
	return &FaceCuller{propsKeyBuilder: pkb}
}

func (svc *FaceCuller) Run(occ *stateful.OccupancyIndex, blockRegions, microRegions []model.Cuboid, blocks []model.RawBlock, styles stateful.StyleIndex) model.FaceVisibility {
	vis := model.FaceVisibility{
		BlockFaces:   make([]model.FaceMask, len(blockRegions)),
		MicroFaces:   make([]model.FaceMask, len(microRegions)),
		ComplexFaces: make([]model.FaceMask, len(blocks)),
	}

	for i, c := range blockRegions {
		vis.BlockFaces[i] = svc.cuboidFaces(occ, c, model.SubGridSize)
	}
	for i, c := range microRegions {
		vis.MicroFaces[i] = svc.cuboidFaces(occ, c, 1)
	}
	for i, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
		style, ok := styles.Get(b.ID, propsKey)
		if !ok || style.GridAlignment != model.GridNotAligned {
			continue
		}
		vis.ComplexFaces[i] = svc.regionFaces(occ,
			b.X*model.SubGridSize, b.Y*model.SubGridSize, b.Z*model.SubGridSize,
			model.SubGridSize, model.SubGridSize, model.SubGridSize)
	}

	return vis
}

func (svc *FaceCuller) cuboidFaces(occ *stateful.OccupancyIndex, c model.Cuboid, scale int) model.FaceMask {
	xMin := int(math.Floor(c.X-float64(c.Width-1)/2.0)) * scale
	yMin := int(math.Floor(c.Y-float64(c.Height-1)/2.0)) * scale
	zMin := int(math.Floor(c.Z-float64(c.Depth-1)/2.0)) * scale
	return svc.regionFaces(occ, xMin, yMin, zMin, c.Width*scale, c.Height*scale, c.Depth*scale)
}

func (svc *FaceCuller) regionFaces(occ *stateful.OccupancyIndex, xMin, yMin, zMin, w, h, d int) model.FaceMask {
	xMax := xMin + w - 1
	yMax := yMin + h - 1
	zMax := zMin + d - 1

	var mask model.FaceMask
	mask[model.FaceIndexTop] = !svc.faceHidden(occ, xMin, xMax, yMax+1, yMax+1, zMin, zMax)
	mask[model.FaceIndexBottom] = !svc.faceHidden(occ, xMin, xMax, yMin-1, yMin-1, zMin, zMax)
	mask[model.FaceIndexFront] = !svc.faceHidden(occ, xMin, xMax, yMin, yMax, zMax+1, zMax+1)
	mask[model.FaceIndexBack] = !svc.faceHidden(occ, xMin, xMax, yMin, yMax, zMin-1, zMin-1)
	mask[model.FaceIndexLeft] = !svc.faceHidden(occ, xMin-1, xMin-1, yMin, yMax, zMin, zMax)
	mask[model.FaceIndexRight] = !svc.faceHidden(occ, xMax+1, xMax+1, yMin, yMax, zMin, zMax)
	return mask
}

func (svc *FaceCuller) faceHidden(occ *stateful.OccupancyIndex, minX, maxX, minY, maxY, minZ, maxZ int) bool {
	for x := minX; x <= maxX; x++ {
		for y := minY; y <= maxY; y++ {
			for z := minZ; z <= maxZ; z++ {
				if !occ.Occluding(x, y, z) {
					return false
				}
			}
		}
	}
	return true
}
