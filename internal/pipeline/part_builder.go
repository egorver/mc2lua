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
	if err := svc.validateVisibility(blocks, blockCuboids, microCuboids, visibility); err != nil {
		return nil, err
	}
	parts, err := svc.buildCuboidParts(blockCuboids, microCuboids, visibility, styleIndex, scale)
	if err != nil {
		return nil, err
	}
	parts = append(parts, svc.buildBlockParts(blocks, visibility, styleIndex, scale)...)
	return parts, nil
}

func (svc *PartBuilder) validateVisibility(blocks []model.RawBlock, blockCuboids, microCuboids []model.Cuboid, visibility model.FaceVisibility) error {
	if len(visibility.BlockFaces) != len(blockCuboids) {
		return fmt.Errorf("visibility mismatch: got %d block face mask(s) for %d block cuboid(s)",
			len(visibility.BlockFaces), len(blockCuboids))
	}
	if len(visibility.MicroFaces) != len(microCuboids) {
		return fmt.Errorf("visibility mismatch: got %d micro face mask(s) for %d micro cuboid(s)",
			len(visibility.MicroFaces), len(microCuboids))
	}
	if len(visibility.ComplexFaces) != len(blocks) {
		return fmt.Errorf("visibility mismatch: got %d complex face mask(s) for %d block(s)",
			len(visibility.ComplexFaces), len(blocks))
	}
	return nil
}

func (svc *PartBuilder) buildCuboidParts(blockCuboids, microCuboids []model.Cuboid, visibility model.FaceVisibility, styleIndex stateful.StyleIndex, scale float64) ([]model.Part, error) {
	var parts []model.Part

	sources := []struct {
		cuboids []model.Cuboid
		faces   []model.FaceMask
	}{
		{blockCuboids, visibility.BlockFaces},
		{microCuboids, visibility.MicroFaces},
	}

	for _, src := range sources {
		for i, c := range src.cuboids {
			style, ok := styleIndex.Get(c.ID, c.PropsKey)
			if !ok {
				continue
			}
			if style.GridAlignment == model.GridNotAligned {
				continue
			}
			part, err := svc.buildSimplePart(c, style, src.faces[i], scale)
			if err != nil {
				return nil, fmt.Errorf("failed to build simple part for block ID %s with properties %s: %w", c.ID, c.PropsKey, err)
			}
			parts = append(parts, part)
		}
	}

	return parts, nil
}

func (svc *PartBuilder) buildBlockParts(blocks []model.RawBlock, visibility model.FaceVisibility, styleIndex stateful.StyleIndex, scale float64) []model.Part {
	var parts []model.Part

	for i, r := range blocks {
		propsKey := svc.propsKeyBuilder.Run(r.Props)
		style, ok := styleIndex.Get(r.ID, propsKey)
		if ok && style.GridAlignment == model.GridNotAligned {
			parts = append(parts, svc.buildComplexParts(r, style, visibility.ComplexFaces[i], scale)...)
		}
	}

	return parts
}

func (svc *PartBuilder) buildSimplePart(cuboid model.Cuboid, style model.StyledBlock, visible model.FaceMask, scale float64) (model.Part, error) {
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
		Name:         svc.makePartName(cuboid.ID),
		Group:        "",
		GroupID:      0,
		BlockID:      cuboid.ID,
		PropsKey:     cuboid.PropsKey,
		Size:         model.Vector3{width, height, depth},
		Position:     model.Vector3{x, y, z},
		Color:        elem.Color,
		Material:     elem.Material,
		VisibleFaces: visible,
	}, nil
}

func (svc *PartBuilder) buildComplexParts(block model.RawBlock, style model.StyledBlock, visible model.FaceMask, scale float64) []model.Part {
	if len(style.Elements) == 0 {
		return nil
	}

	xBase := float64(block.X)*scale - scale/2.0
	yBase := float64(block.Y)*scale - scale/2.0
	zBase := float64(block.Z)*scale - scale/2.0

	groupID, groupName := svc.resolveGroup(block, style, xBase, yBase, zBase, scale)

	parts := make([]model.Part, 0, len(style.Elements))
	for idx, elem := range style.Elements {
		parts = append(parts, svc.buildElementPart(block, elem, style.PropsKey, idx, groupID, groupName, visible, xBase, yBase, zBase, scale))
	}
	return parts
}

func (svc *PartBuilder) resolveGroup(block model.RawBlock, style model.StyledBlock, xBase, yBase, zBase, scale float64) (int, string) {
	if len(style.Elements) < 2 {
		return 0, ""
	}

	firstPos := svc.elementKey(style.Elements[0], xBase, yBase, zBase, scale)
	distinct := false
	for _, elem := range style.Elements[1:] {
		if svc.elementKey(elem, xBase, yBase, zBase, scale) != firstPos {
			distinct = true
			break
		}
	}
	if !distinct {
		return 0, ""
	}

	svc.groupCounter++
	return svc.groupCounter, svc.makePartName(block.ID)
}

func (svc *PartBuilder) buildElementPart(block model.RawBlock, elem model.StyledElement, propsKey string, idx, groupID int, groupName string, visible model.FaceMask, xBase, yBase, zBase, scale float64) model.Part {
	sizeX := (elem.To[0] - elem.From[0]) / model.FullBlockSize * scale
	sizeY := (elem.To[1] - elem.From[1]) / model.FullBlockSize * scale
	sizeZ := (elem.To[2] - elem.From[2]) / model.FullBlockSize * scale

	center := svc.elementCenter(elem, xBase, yBase, zBase, scale)

	partName := svc.makePartName(block.ID)
	if groupID != 0 {
		partName = fmt.Sprintf(elementNameFormat, idx+1)
	}

	return model.Part{
		Name:         partName,
		Group:        groupName,
		GroupID:      groupID,
		BlockID:      block.ID,
		PropsKey:     propsKey,
		Size:         model.Vector3{sizeX, sizeY, sizeZ},
		Position:     model.Vector3{center[0], center[1], center[2]},
		Color:        elem.Color,
		Material:     elem.Material,
		VisibleFaces: visible,
	}
}

func (svc *PartBuilder) elementCenter(elem model.StyledElement, xBase, yBase, zBase, scale float64) model.Vector3 {
	return model.Vector3{
		xBase + (elem.From[0]+elem.To[0])/2.0/model.FullBlockSize*scale,
		yBase + (elem.From[1]+elem.To[1])/2.0/model.FullBlockSize*scale,
		zBase + (elem.From[2]+elem.To[2])/2.0/model.FullBlockSize*scale,
	}
}

func (svc *PartBuilder) elementKey(elem model.StyledElement, xBase, yBase, zBase, scale float64) string {
	center := svc.elementCenter(elem, xBase, yBase, zBase, scale)
	return fmt.Sprintf("%.4f,%.4f,%.4f", center[0], center[1], center[2])
}

func (svc *PartBuilder) makePartName(blockID string) string {
	return strings.TrimPrefix(blockID, minecraft.MinecraftNamespacePrefix)
}
