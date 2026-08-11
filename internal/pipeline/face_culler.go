package pipeline

import (
	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

type cullerPropsKeyBuilder interface {
	Run(props map[string]string) string
}

type cullerCuboidHelper interface {
	MinCorner(c model.Cuboid) (x, y, z int)
}

type FaceCuller struct {
	propsKeyBuilder cullerPropsKeyBuilder
	cuboidHelper    cullerCuboidHelper
}

func NewFaceCuller(pkb cullerPropsKeyBuilder, ch cullerCuboidHelper) *FaceCuller {
	return &FaceCuller{propsKeyBuilder: pkb, cuboidHelper: ch}
}

func (svc *FaceCuller) Run(occ *stateful.OccupancyIndex, blockRegions, microRegions []model.Cuboid, blocks []model.RawBlock, styles stateful.StyleIndex) model.FaceVisibility {
	blockFaces := svc.cullBlockRegions(occ, blockRegions)
	microFaces := svc.cullMicroRegions(occ, microRegions)
	complexFaces := svc.cullComplexBlocks(occ, blocks, styles)

	vis := model.FaceVisibility{
		BlockFaces:   blockFaces,
		MicroFaces:   microFaces,
		ComplexFaces: complexFaces,
	}
	return vis
}

func (svc *FaceCuller) cullBlockRegions(occ *stateful.OccupancyIndex, regions []model.Cuboid) []model.FaceMask {
	masks := make([]model.FaceMask, len(regions))
	for i, c := range regions {
		masks[i] = svc.cuboidFaces(occ, c, model.SubGridSize)
	}
	return masks
}

func (svc *FaceCuller) cullMicroRegions(occ *stateful.OccupancyIndex, regions []model.Cuboid) []model.FaceMask {
	masks := make([]model.FaceMask, len(regions))
	for i, c := range regions {
		masks[i] = svc.cuboidFaces(occ, c, 1)
	}
	return masks
}

func (svc *FaceCuller) cullComplexBlocks(occ *stateful.OccupancyIndex, blocks []model.RawBlock, styles stateful.StyleIndex) []model.FaceMask {
	masks := make([]model.FaceMask, len(blocks))
	for i, b := range blocks {
		propsKey := svc.propsKeyBuilder.Run(b.Props)
		style, ok := styles.Get(b.ID, propsKey)
		if !ok || style.GridAlignment != model.GridNotAligned {
			continue
		}
		masks[i] = svc.regionFaces(occ,
			b.X*model.SubGridSize, b.Y*model.SubGridSize, b.Z*model.SubGridSize,
			model.SubGridSize, model.SubGridSize, model.SubGridSize)
	}
	return masks
}

func (svc *FaceCuller) cuboidFaces(occ *stateful.OccupancyIndex, c model.Cuboid, scale int) model.FaceMask {
	xMin, yMin, zMin := svc.cuboidHelper.MinCorner(c)
	xMin *= scale
	yMin *= scale
	zMin *= scale
	return svc.regionFaces(occ, xMin, yMin, zMin, c.Width*scale, c.Height*scale, c.Depth*scale)
}

func (svc *FaceCuller) regionFaces(occ *stateful.OccupancyIndex, xMin, yMin, zMin, w, h, d int) model.FaceMask {
	xMax := xMin + w - 1
	yMax := yMin + h - 1
	zMax := zMin + d - 1

	var mask model.FaceMask
	mask[model.FaceIndexTop] = !svc.faceHidden(occ, xMin, xMax, yMax+1, yMax+1, zMin, zMax)
	mask[model.FaceIndexBottom] = !svc.faceHidden(occ, xMin, xMax, yMin-1, yMin-1, zMin, zMax)
	mask[model.FaceIndexFront] = !svc.faceHidden(occ, xMin, xMax, yMin, yMax, zMin-1, zMin-1)
	mask[model.FaceIndexBack] = !svc.faceHidden(occ, xMin, xMax, yMin, yMax, zMax+1, zMax+1)
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
