package pipeline

import (
	"fmt"
	"math"
	"math/bits"
	"path/filepath"

	"github.com/Tnze/go-mc/level"
	"github.com/Tnze/go-mc/save"
	"github.com/Tnze/go-mc/save/region"

	"mc2lua/internal/model"
)

type WorldReader struct {
	fs fsApi
}

func NewWorldReader(fs fsApi) *WorldReader {
	return &WorldReader{fs: fs}
}

func (svc *WorldReader) Run(input string, bounds model.Bounds) (*model.World, error) {
	files, err := svc.listRegionFiles(input)
	if err != nil {
		return nil, fmt.Errorf("list region files: %w", err)
	}

	if len(files) == 0 {
		return &model.World{}, nil
	}

	var blocks []model.Block
	minX, minY, minZ := math.MaxInt32, math.MaxInt32, math.MaxInt32
	maxX, maxY, maxZ := math.MinInt32, math.MinInt32, math.MinInt32

	for _, name := range files {
		rx, rz := svc.parseRegionCoord(name)

		if !svc.rangeOverlap(rx*512, rx*512+511, bounds.XMin, bounds.XMax) ||
			!svc.rangeOverlap(rz*512, rz*512+511, bounds.ZMin, bounds.ZMax) {
			continue
		}

		regionBlocks := svc.processRegion(input, name, rx, rz, bounds)
		for _, b := range regionBlocks {
			if b.X < minX {
				minX = b.X
			}
			if b.X > maxX {
				maxX = b.X
			}
			if b.Y < minY {
				minY = b.Y
			}
			if b.Y > maxY {
				maxY = b.Y
			}
			if b.Z < minZ {
				minZ = b.Z
			}
			if b.Z > maxZ {
				maxZ = b.Z
			}
		}
		blocks = append(blocks, regionBlocks...)
	}

	if len(blocks) == 0 {
		return &model.World{}, nil
	}

	lookup := make(map[[3]int]*model.Block, len(blocks))
	for i := range blocks {
		b := &blocks[i]
		lookup[[3]int{b.X, b.Y, b.Z}] = b
	}

	return &model.World{
		Blocks: blocks,
		Lookup: lookup,
		MinX:   minX,
		MinY:   minY,
		MinZ:   minZ,
		MaxX:   maxX,
		MaxY:   maxY,
		MaxZ:   maxZ,
	}, nil
}

func (svc *WorldReader) listRegionFiles(input string) ([]string, error) {
	entries, err := svc.fs.ReadDir(input)
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		ok, _ := filepath.Match("r.*.*.mca", name)
		if ok {
			files = append(files, name)
		}
	}
	return files, nil
}

func (svc *WorldReader) processRegion(input, name string, rx, rz int, bounds model.Bounds) []model.Block {
	r, err := region.Open(filepath.Join(input, name))
	if err != nil {
		return nil
	}
	defer r.Close()

	var blocks []model.Block
	for lx := 0; lx < 32; lx++ {
		for lz := 0; lz < 32; lz++ {
			if !r.ExistSector(lx, lz) {
				continue
			}

			chunkX := rx*32 + lx
			chunkZ := rz*32 + lz

			if !svc.rangeOverlap(chunkX*16, chunkX*16+15, bounds.XMin, bounds.XMax) ||
				!svc.rangeOverlap(chunkZ*16, chunkZ*16+15, bounds.ZMin, bounds.ZMax) {
				continue
			}

			chunkBlocks := svc.processChunk(r, lx, lz, chunkX, chunkZ, bounds)
			blocks = append(blocks, chunkBlocks...)
		}
	}
	return blocks
}

func (svc *WorldReader) processChunk(r *region.Region, lx, lz, chunkX, chunkZ int, bounds model.Bounds) []model.Block {
	data, err := r.ReadSector(lx, lz)
	if err != nil {
		return nil
	}

	var sc save.Chunk
	if err := sc.Load(data); err != nil {
		return nil
	}

	if sc.Status != "full" && sc.Status != "minecraft:full" {
		return nil
	}

	var blocks []model.Block
	for _, ssec := range sc.Sections {
		sectionBlocks := svc.processSection(ssec, chunkX, chunkZ, bounds)
		blocks = append(blocks, sectionBlocks...)
	}
	return blocks
}

func (svc *WorldReader) processSection(ssec save.Section, chunkX, chunkZ int, bounds model.Bounds) []model.Block {
	if len(ssec.BlockStates.Palette) == 0 {
		return nil
	}

	baseY := int(ssec.Y) * 16
	if !svc.rangeOverlap(baseY, baseY+15, bounds.YMin, bounds.YMax) {
		return nil
	}

	bs := svc.newBitStorage(len(ssec.BlockStates.Palette), ssec.BlockStates.Data)
	palette := ssec.BlockStates.Palette

	var blocks []model.Block
	for j := 0; j < 4096; j++ {
		block := svc.decodeBlock(bs, j, palette, chunkX, chunkZ, baseY, bounds)
		if block != nil {
			blocks = append(blocks, *block)
		}
	}
	return blocks
}

func (svc *WorldReader) newBitStorage(palSize int, data []uint64) *level.BitStorage {
	if palSize == 1 {
		return level.NewBitStorage(0, 4096, nil)
	}
	bitsPerEntry := bits.Len(uint(palSize - 1))
	if bitsPerEntry < 4 {
		bitsPerEntry = 4
	}
	return level.NewBitStorage(bitsPerEntry, 4096, data)
}

func (svc *WorldReader) decodeBlock(bs *level.BitStorage, j int, palette []save.BlockState, chunkX, chunkZ, baseY int, bounds model.Bounds) *model.Block {
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

	worldX := chunkX*16 + x
	worldY := baseY + y
	worldZ := chunkZ*16 + z

	if worldX < bounds.XMin || worldX > bounds.XMax ||
		worldY < bounds.YMin || worldY > bounds.YMax ||
		worldZ < bounds.ZMin || worldZ > bounds.ZMax {
		return nil
	}

	return &model.Block{
		ID:         blockName,
		Properties: props,
		X:          worldX,
		Y:          worldY,
		Z:          worldZ,
	}
}

func (svc *WorldReader) parseRegionCoord(name string) (int, int) {
	var rx, rz int
	fmt.Sscanf(name, "r.%d.%d.mca", &rx, &rz)
	return rx, rz
}

func (svc *WorldReader) rangeOverlap(min1, max1, min2, max2 int) bool {
	return max(min1, min2) <= min(max1, max2)
}
