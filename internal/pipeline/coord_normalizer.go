package pipeline

import (
	"fmt"

	"mc2lua/internal/model"
)

type CoordNormalizer struct{}

func NewCoordNormalizer() *CoordNormalizer {
	return &CoordNormalizer{}
}

func (svc *CoordNormalizer) Run(blocks []model.Block, noOffset bool) ([]model.Block, error) {
	if len(blocks) == 0 {
		return nil, fmt.Errorf("coord normalizer: empty block list")
	}

	minX, minY, minZ := computeMinCoords(blocks)
	xOff, yOff, zOff := computeOffsets(minX, minY, minZ, noOffset)
	result := applyOffset(blocks, xOff, yOff, zOff)

	return result, nil
}

func computeMinCoords(blocks []model.Block) (minX, minY, minZ int) {
	minX, minY, minZ = blocks[0].X, blocks[0].Y, blocks[0].Z
	for _, b := range blocks[1:] {
		if b.X < minX {
			minX = b.X
		}
		if b.Y < minY {
			minY = b.Y
		}
		if b.Z < minZ {
			minZ = b.Z
		}
	}
	return
}

func computeOffsets(minX, minY, minZ int, noOffset bool) (xOff, yOff, zOff int) {
	xOff, zOff = -minX, -minZ
	if !noOffset && minY != 0 {
		yOff = -minY
	}
	return
}

func applyOffset(blocks []model.Block, xOff, yOff, zOff int) []model.Block {
	out := make([]model.Block, len(blocks))
	for i, b := range blocks {
		out[i] = model.Block{
			ID:    b.ID,
			Props: b.Props,
			X:     b.X + xOff,
			Y:     b.Y + yOff,
			Z:     b.Z + zOff,
		}
	}
	return out
}
