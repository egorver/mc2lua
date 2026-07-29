package minecraft

import (
	"math/bits"

	"mc2lua/internal/model"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
)

type ChunkDecoder struct{}

func NewChunkDecoder() *ChunkDecoder {
	return &ChunkDecoder{}
}

func (svc *ChunkDecoder) Run(data []byte, chunkX, chunkZ int) ([]model.RawBlock, error) {
	var sc save.Chunk
	if err := sc.Load(data); err != nil {
		return nil, err
	}

	if sc.Status != "full" && sc.Status != "minecraft:full" {
		return nil, nil
	}

	var blocks []model.RawBlock
	for _, ssec := range sc.Sections {
		sectionBlocks := svc.decodeSection(ssec, chunkX, chunkZ)
		blocks = append(blocks, sectionBlocks...)
	}
	return blocks, nil
}

func (svc *ChunkDecoder) decodeSection(ssec save.Section, chunkX, chunkZ int) []model.RawBlock {
	if len(ssec.BlockStates.Palette) == 0 {
		return nil
	}

	baseY := int(ssec.Y) * 16
	bs := svc.newBitStorage(len(ssec.BlockStates.Palette), ssec.BlockStates.Data)
	palette := ssec.BlockStates.Palette

	var blocks []model.RawBlock
	for j := 0; j < 4096; j++ {
		block := svc.decodeBlock(bs, j, palette, chunkX, chunkZ, baseY)
		if block != nil {
			blocks = append(blocks, *block)
		}
	}
	return blocks
}

func (svc *ChunkDecoder) newBitStorage(palSize int, data []uint64) *level.BitStorage {
	if palSize <= 1 {
		return level.NewBitStorage(0, 4096, nil)
	}
	bitsPerEntry := bits.Len(uint(palSize - 1)) //nolint:gosec
	if bitsPerEntry < 4 {
		bitsPerEntry = 4
	}
	return level.NewBitStorage(bitsPerEntry, 4096, data)
}

func (svc *ChunkDecoder) decodeBlock(bs *level.BitStorage, j int, palette []save.BlockState, chunkX, chunkZ, baseY int) *model.RawBlock {
	idx := bs.Get(j)
	if idx >= len(palette) {
		return nil
	}

	blockName := palette[idx].Name
	if blockName == "minecraft:air" || blockName == "minecraft:cave_air" || blockName == "minecraft:void_air" {
		return nil
	}

	var props map[string]string
	if p := palette[idx].Properties; p.Type != 0 {
		var m map[string]string
		if err := p.Unmarshal(&m); err == nil && len(m) > 0 {
			props = m
		}
	}

	x := j & 15
	z := (j >> 4) & 15
	y := j >> 8

	return &model.RawBlock{
		ID:    blockName,
		Props: props,
		X:     chunkX*16 + x,
		Y:     baseY + y,
		Z:     chunkZ*16 + z,
	}
}
