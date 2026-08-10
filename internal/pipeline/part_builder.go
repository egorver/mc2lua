package pipeline

import (
	"fmt"
	"strings"

	"mc2lua/internal/minecraft"
	"mc2lua/internal/model"
	"mc2lua/internal/stateful"
)

const elementNameFormat = "elem %d"

type partPropsKeyBuilder interface {
	Run(props map[string]string) string
}

type PartBuilder struct {
	propsKeyBuilder partPropsKeyBuilder
	groupCounter    int
}

func NewPartBuilder(pkb partPropsKeyBuilder) *PartBuilder {
	return &PartBuilder{propsKeyBuilder: pkb}
}

func (svc *PartBuilder) Run(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, visibility model.FaceVisibility, styleIndex stateful.StyleIndex, scale float64) ([]model.Part, error) {
	var parts []model.Part

	for _, cuboids := range [][]model.Cuboid{blockCuboids, microCuboids} {
		for _, c := range cuboids {
			style, ok := styleIndex.Get(c.ID, c.PropsKey)
			if !ok {
				continue
			}
			if style.GridAlignment == model.GridNotAligned {
				continue
			}
			part, err := svc.buildSimplePart(c, style, scale)
			if err != nil {
				return nil, fmt.Errorf("failed to build simple part for block ID %s with properties %s: %w", c.ID, c.PropsKey, err)
			}
			parts = append(parts, part)
		}
	}

	for _, r := range blocks {
		propsKey := svc.propsKeyBuilder.Run(r.Props)
		style, ok := styleIndex.Get(r.ID, propsKey)
		if ok && style.GridAlignment == model.GridNotAligned {
			parts = append(parts, svc.buildComplexParts(r, style, scale)...)
		}
	}

	return parts, nil
}

func (svc *PartBuilder) buildSimplePart(cuboid model.Cuboid, style model.StyledBlock, scale float64) (model.Part, error) {
	if len(style.Elements) == 0 {
		return model.Part{}, fmt.Errorf("no elements found for block ID %s with properties %s", cuboid.ID, cuboid.PropsKey)
	}

	elem := style.Elements[0]

	gs := 1.0
	if style.GridAlignment == model.GridSubBlock {
		gs = float64(model.SubGridSize)
	}

	x := (float64(cuboid.X) - (gs-1)/2) * scale / gs
	y := (float64(cuboid.Y) - (gs-1)/2) * scale / gs
	z := (float64(cuboid.Z) - (gs-1)/2) * scale / gs
	width := float64(cuboid.Width) * scale / gs
	height := float64(cuboid.Height) * scale / gs
	depth := float64(cuboid.Depth) * scale / gs

	return model.Part{
		Name:     svc.makePartName(cuboid.ID),
		Group:    "",
		GroupID:  0,
		BlockID:  cuboid.ID,
		PropsKey: cuboid.PropsKey,
		Size:     model.Vector3{width, height, depth},
		Position: model.Vector3{x, y, z},
		Color:    elem.Color,
		Material: elem.Material,
	}, nil
}

func (svc *PartBuilder) buildComplexParts(block model.RawBlock, style model.StyledBlock, scale float64) []model.Part {
	if len(style.Elements) == 0 {
		return nil
	}

	xBase := float64(block.X)*scale - scale/2.0
	yBase := float64(block.Y)*scale - scale/2.0
	zBase := float64(block.Z)*scale - scale/2.0

	groupID := 0
	groupName := ""
	if len(style.Elements) > 1 {
		firstPos := svc.elementKey(style.Elements[0], xBase, yBase, zBase, scale)
		distinct := false
		for _, elem := range style.Elements[1:] {
			if svc.elementKey(elem, xBase, yBase, zBase, scale) != firstPos {
				distinct = true
				break
			}
		}
		if distinct {
			svc.groupCounter++
			groupID = svc.groupCounter
			groupName = svc.makePartName(block.ID)
		}
	}

	parts := make([]model.Part, 0, len(style.Elements))
	for idx, elem := range style.Elements {
		sizeX := (elem.To[0] - elem.From[0]) / model.FullBlockSize * scale
		sizeY := (elem.To[1] - elem.From[1]) / model.FullBlockSize * scale
		sizeZ := (elem.To[2] - elem.From[2]) / model.FullBlockSize * scale

		cx := (elem.From[0] + elem.To[0]) / 2.0 / model.FullBlockSize * scale
		cy := (elem.From[1] + elem.To[1]) / 2.0 / model.FullBlockSize * scale
		cz := (elem.From[2] + elem.To[2]) / 2.0 / model.FullBlockSize * scale

		px := xBase + cx
		py := yBase + cy
		pz := zBase + cz

		partName := svc.makePartName(block.ID)
		if groupID != 0 {
			partName = fmt.Sprintf(elementNameFormat, idx+1)
		}

		parts = append(parts, model.Part{
			Name:     partName,
			Group:    groupName,
			GroupID:  groupID,
			BlockID:  block.ID,
			PropsKey: style.PropsKey,
			Size:     model.Vector3{sizeX, sizeY, sizeZ},
			Position: model.Vector3{px, py, pz},
			Color:    elem.Color,
			Material: elem.Material,
		})
	}
	return parts
}

func (svc *PartBuilder) elementKey(elem model.StyledElement, xBase, yBase, zBase, scale float64) string {
	cx := xBase + (elem.From[0]+elem.To[0])/2.0/model.FullBlockSize*scale
	cy := yBase + (elem.From[1]+elem.To[1])/2.0/model.FullBlockSize*scale
	cz := zBase + (elem.From[2]+elem.To[2])/2.0/model.FullBlockSize*scale
	return fmt.Sprintf("%.4f,%.4f,%.4f", cx, cy, cz)
}

func (svc *PartBuilder) makePartName(blockID string) string {
	return strings.TrimPrefix(blockID, minecraft.MinecraftNamespacePrefix)
}
