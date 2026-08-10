package pipeline

import (
	"testing"

	"mc2lua/internal/model"
	"mc2lua/internal/stateful"

	"github.com/stretchr/testify/require"
)

type mockGeneratorPropsKeyBuilder struct {
	runFn func(props map[string]string) string
}

func (m *mockGeneratorPropsKeyBuilder) Run(props map[string]string) string {
	if m.runFn != nil {
		return m.runFn(props)
	}
	return ""
}

func testPartBuilder(t *testing.T, pkb *mockGeneratorPropsKeyBuilder) *PartBuilder {
	t.Helper()
	if pkb == nil {
		pkb = &mockGeneratorPropsKeyBuilder{}
	}
	return NewPartBuilder(pkb)
}

func makeStyledBlock(id, propsKey string, alignment model.GridAlignment, elems ...model.StyledElement) model.StyledBlock {
	return model.StyledBlock{ID: id, PropsKey: propsKey, GridAlignment: alignment, Elements: elems}
}

func makeStyledElement(fromX, fromY, fromZ, toX, toY, toZ float64, color model.Color, material string) model.StyledElement {
	return model.StyledElement{
		From:     model.Vector3{fromX, fromY, fromZ},
		To:       model.Vector3{toX, toY, toZ},
		Color:    color,
		Material: material,
	}
}

func TestPartBuilder_New(t *testing.T) {
	t.Parallel()

	pb := NewPartBuilder(&mockGeneratorPropsKeyBuilder{})
	require.NotNil(t, pb)
}

func TestMakePartName(t *testing.T) {
	t.Parallel()

	svc := testPartBuilder(t, nil)

	tests := []struct {
		name    string
		blockID string
		want    string
	}{
		{name: "with minecraft prefix", blockID: "minecraft:stone", want: "stone"},
		{name: "without prefix", blockID: "stone", want: "stone"},
		{name: "empty", blockID: "", want: ""},
		{name: "custom namespace", blockID: "cobblemon:ore", want: "cobblemon:ore"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := svc.makePartName(tt.blockID)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestElementKey(t *testing.T) {
	t.Parallel()

	svc := testPartBuilder(t, nil)
	elem := makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{}, "")

	key := svc.elementKey(elem, -2, 6, 2, 4)
	require.Equal(t, "0.0000,7.0000,4.0000", key)
}

func TestBuildSimplePart(t *testing.T) {
	t.Parallel()

	svc := testPartBuilder(t, nil)

	tests := []struct {
		name       string
		cuboid     model.Cuboid
		style      model.StyledBlock
		scale      float64
		wantErr    bool
		wantErrMsg string
		wantCheck  func(t *testing.T, part model.Part)
	}{
		{
			name:   "normal",
			cuboid: model.Cuboid{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 2, Z: 1, Width: 3, Height: 2, Depth: 4},
			style: makeStyledBlock("minecraft:stone", "", model.GridFullBlock,
				makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{131, 84, 50}, "Wood"),
			),
			scale: 4,
			wantCheck: func(t *testing.T, part model.Part) {
				require.Equal(t, "stone", part.Name)
				require.Equal(t, "", part.Group)
				require.Equal(t, 0, part.GroupID)
				require.Equal(t, "minecraft:stone", part.BlockID)
				require.Equal(t, "", part.PropsKey)
				require.Equal(t, model.Vector3{12, 8, 16}, part.Size)
				require.Equal(t, model.Vector3{0, 8, 4}, part.Position)
				require.Equal(t, model.Color{131, 84, 50}, part.Color)
				require.Equal(t, "Wood", part.Material)
			},
		},
		{
			name:   "micro block sub grid",
			cuboid: model.Cuboid{ID: "minecraft:stone_slab", PropsKey: "", X: 1.5, Y: 0.5, Z: 1.5, Width: 4, Height: 2, Depth: 4},
			style: makeStyledBlock("minecraft:stone_slab", "", model.GridSubBlock,
				makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{128, 128, 128}, "Stone"),
			),
			scale: 4,
			wantCheck: func(t *testing.T, part model.Part) {
				require.Equal(t, model.Vector3{4, 2, 4}, part.Size)
				require.Equal(t, model.Vector3{0, -1, 0}, part.Position)
			},
		},
		{
			name:       "empty elements",
			cuboid:     model.Cuboid{ID: "minecraft:stone", PropsKey: ""},
			style:      makeStyledBlock("minecraft:stone", "", model.GridFullBlock),
			scale:      4,
			wantErr:    true,
			wantErrMsg: "no elements found for block ID minecraft:stone with properties ",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			part, err := svc.buildSimplePart(tt.cuboid, tt.style, tt.scale)
			if tt.wantErr {
				require.Error(t, err)
				if tt.wantErrMsg != "" {
					require.Contains(t, err.Error(), tt.wantErrMsg)
				}
				return
			}
			require.NoError(t, err)
			if tt.wantCheck != nil {
				tt.wantCheck(t, part)
			}
		})
	}
}

func TestBuildComplexParts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		block     model.RawBlock
		style     model.StyledBlock
		scale     float64
		wantCount int
		wantCheck func(t *testing.T, parts []model.Part)
	}{
		{
			name:  "single element no group",
			block: model.RawBlock{ID: "minecraft:stone", X: 0, Y: 2, Z: 1},
			style: makeStyledBlock("minecraft:stone", "", model.GridNotAligned,
				makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{128, 128, 128}, "Slate"),
			),
			scale:     4,
			wantCount: 1,
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, 0, parts[0].GroupID)
				require.Equal(t, "", parts[0].Group)
				require.Equal(t, "stone", parts[0].Name)
			},
		},
		{
			name:  "elements at same position no group",
			block: model.RawBlock{X: 0, Y: 0, Z: 0},
			style: makeStyledBlock("minecraft:stone", "", model.GridNotAligned,
				makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{128, 128, 128}, "Slate"),
				makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{128, 128, 128}, "Slate"),
			),
			scale:     4,
			wantCount: 2,
			wantCheck: func(t *testing.T, parts []model.Part) {
				for _, p := range parts {
					require.Equal(t, 0, p.GroupID)
					require.Equal(t, "", p.Group)
				}
			},
		},
		{
			name:  "elements at different positions with group",
			block: model.RawBlock{ID: "minecraft:oak_stairs", X: 0, Y: 0, Z: 0},
			style: makeStyledBlock("minecraft:oak_stairs", "half=bottom,facing=north", model.GridNotAligned,
				makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{131, 84, 50}, "Wood"),
				makeStyledElement(0, 8, 0, 16, 16, 8, model.Color{131, 84, 50}, "Wood"),
			),
			scale:     4,
			wantCount: 2,
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, 1, parts[0].GroupID)
				require.Equal(t, "oak_stairs", parts[0].Group)
				require.Equal(t, "elem 1", parts[0].Name)
				require.Equal(t, "elem 2", parts[1].Name)
			},
		},
		{
			name:      "empty elements",
			block:     model.RawBlock{X: 0, Y: 0, Z: 0},
			style:     makeStyledBlock("minecraft:stone", "", model.GridNotAligned),
			scale:     4,
			wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := testPartBuilder(t, nil)
			parts := svc.buildComplexParts(tt.block, tt.style, tt.scale)
			require.Len(t, parts, tt.wantCount)
			if tt.wantCheck != nil {
				tt.wantCheck(t, parts)
			}
		})
	}
}

func TestPartBuilder_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		blocks       []model.RawBlock
		blockCuboids []model.Cuboid
		microCuboids []model.Cuboid
		styleIdx     *stateful.StyleIndex
		pkbFn        func(props map[string]string) string
		scale        float64
		wantCount    int
		wantErrMsg   string
		wantCheck    func(t *testing.T, parts []model.Part)
	}{
		{
			name:      "empty input",
			blocks:    nil,
			styleIdx:  stateful.NewStyleIndex(),
			scale:     4,
			wantCount: 0,
		},
		{
			name: "only full blocks",
			blockCuboids: []model.Cuboid{
				{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 0, Z: 0, Width: 1, Depth: 1, Height: 1},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:stone", "", makeStyledBlock("minecraft:stone", "", model.GridFullBlock,
					makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{128, 128, 128}, "Slate"),
				))
				return idx
			}(),
			scale:     4,
			wantCount: 1,
			wantCheck: func(t *testing.T, parts []model.Part) {
				p := parts[0]
				require.Equal(t, "stone", p.Name)
				require.Equal(t, "", p.Group)
				require.Equal(t, 0, p.GroupID)
				require.Equal(t, "minecraft:stone", p.BlockID)
				require.Equal(t, "", p.PropsKey)
				require.Equal(t, model.Vector3{4, 4, 4}, p.Size)
				require.Equal(t, model.Vector3{0, 0, 0}, p.Position)
				require.Equal(t, model.Color{128, 128, 128}, p.Color)
				require.Equal(t, "Slate", p.Material)
			},
		},
		{
			name: "only micro blocks",
			microCuboids: []model.Cuboid{
				{ID: "minecraft:stone_slab", PropsKey: "", X: 1.5, Y: 0.5, Z: 1.5, Width: 4, Height: 2, Depth: 4},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:stone_slab", "", makeStyledBlock("minecraft:stone_slab", "", model.GridSubBlock,
					makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{128, 128, 128}, "Stone"),
				))
				return idx
			}(),
			scale:     4,
			wantCount: 1,
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, model.Vector3{4, 2, 4}, parts[0].Size)
				require.Equal(t, model.Vector3{0, -1, 0}, parts[0].Position)
			},
		},
		{
			name: "only complex blocks",
			blocks: []model.RawBlock{
				{ID: "minecraft:oak_stairs", Props: map[string]string{"half": "bottom", "facing": "north"}, X: 0, Y: 0, Z: 0},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:oak_stairs", "facing=north,half=bottom", makeStyledBlock("minecraft:oak_stairs", "facing=north,half=bottom", model.GridNotAligned,
					makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{131, 84, 50}, "Wood"),
					makeStyledElement(0, 8, 0, 16, 16, 8, model.Color{131, 84, 50}, "Wood"),
				))
				return idx
			}(),
			pkbFn: func(props map[string]string) string {
				return "facing=north,half=bottom"
			},
			scale:     4,
			wantCount: 2,
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, 1, parts[0].GroupID)
				require.Equal(t, "oak_stairs", parts[0].Group)
				require.Equal(t, "elem 1", parts[0].Name)
				require.Equal(t, "elem 2", parts[1].Name)
			},
		},
		{
			name: "skips cuboids with missing style or not-aligned style",
			blockCuboids: []model.Cuboid{
				{ID: "minecraft:oak_stairs", PropsKey: "facing=north", X: 0, Y: 0, Z: 0, Width: 1, Depth: 1, Height: 1},
				{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 0, Z: 0, Width: 1, Depth: 1, Height: 1},
				{ID: "minecraft:unknown", PropsKey: "", X: 0, Y: 0, Z: 0, Width: 1, Depth: 1, Height: 1},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:oak_stairs", "facing=north", makeStyledBlock(
					"minecraft:oak_stairs", "facing=north", model.GridNotAligned,
				))
				idx.Add("minecraft:stone", "", makeStyledBlock(
					"minecraft:stone", "", model.GridFullBlock,
					makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{128, 128, 128}, "Slate"),
				))
				return idx
			}(),
			scale:     4,
			wantCount: 1,
		},
		{
			name: "skips raw block with full block style",
			blocks: []model.RawBlock{
				{ID: "minecraft:stone", Props: map[string]string{}, X: 0, Y: 0, Z: 0},
				{ID: "minecraft:unknown", Props: map[string]string{}, X: 0, Y: 0, Z: 0},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:stone", "", makeStyledBlock(
					"minecraft:stone", "", model.GridFullBlock,
				))
				return idx
			}(),
			pkbFn:     func(props map[string]string) string { return "" },
			scale:     4,
			wantCount: 0,
		},
		{
			name: "aligned cuboid without elements returns error",
			blockCuboids: []model.Cuboid{
				{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 0, Z: 0, Width: 1, Depth: 1, Height: 1},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:stone", "", makeStyledBlock(
					"minecraft:stone", "", model.GridFullBlock,
				))
				return idx
			}(),
			scale:      4,
			wantErrMsg: "failed to build simple part for block ID minecraft:stone with properties ",
		},
		{
			name: "full and complex blocks",
			blocks: []model.RawBlock{
				{ID: "minecraft:oak_stairs", Props: map[string]string{"half": "bottom", "facing": "north"}, X: 0, Y: 0, Z: 0},
			},
			blockCuboids: []model.Cuboid{
				{ID: "minecraft:stone", PropsKey: "", X: 0, Y: 0, Z: 0, Width: 1, Depth: 1, Height: 1},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:stone", "", makeStyledBlock("minecraft:stone", "", model.GridFullBlock,
					makeStyledElement(0, 0, 0, 16, 16, 16, model.Color{128, 128, 128}, "Slate"),
				))
				idx.Add("minecraft:oak_stairs", "facing=north,half=bottom", makeStyledBlock("minecraft:oak_stairs", "facing=north,half=bottom", model.GridNotAligned,
					makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{131, 84, 50}, "Wood"),
					makeStyledElement(0, 8, 0, 16, 16, 8, model.Color{131, 84, 50}, "Wood"),
				))
				return idx
			}(),
			pkbFn: func(props map[string]string) string {
				return "facing=north,half=bottom"
			},
			scale:     4,
			wantCount: 3,
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, 0, parts[0].GroupID)
				require.Equal(t, 1, parts[1].GroupID)
				require.Equal(t, 1, parts[2].GroupID)
			},
		},
		{
			name: "distinct groups get increasing ids",
			blocks: []model.RawBlock{
				{ID: "minecraft:oak_stairs", Props: map[string]string{"half": "bottom", "facing": "north"}, X: 0, Y: 0, Z: 0},
				{ID: "minecraft:oak_stairs", Props: map[string]string{"half": "bottom", "facing": "south"}, X: 3, Y: 0, Z: 0},
			},
			styleIdx: func() *stateful.StyleIndex {
				idx := stateful.NewStyleIndex()
				idx.Add("minecraft:oak_stairs", "facing=north,half=bottom", makeStyledBlock("minecraft:oak_stairs", "facing=north,half=bottom", model.GridNotAligned,
					makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{131, 84, 50}, "Wood"),
					makeStyledElement(0, 8, 0, 16, 16, 8, model.Color{131, 84, 50}, "Wood"),
				))
				idx.Add("minecraft:oak_stairs", "facing=south,half=bottom", makeStyledBlock("minecraft:oak_stairs", "facing=south,half=bottom", model.GridNotAligned,
					makeStyledElement(0, 0, 0, 16, 8, 16, model.Color{131, 84, 50}, "Wood"),
					makeStyledElement(0, 8, 0, 16, 16, 8, model.Color{131, 84, 50}, "Wood"),
				))
				return idx
			}(),
			pkbFn: func(props map[string]string) string {
				if props["facing"] == "south" {
					return "facing=south,half=bottom"
				}
				return "facing=north,half=bottom"
			},
			scale:     4,
			wantCount: 4,
			wantCheck: func(t *testing.T, parts []model.Part) {
				require.Equal(t, 1, parts[0].GroupID)
				require.Equal(t, 1, parts[1].GroupID)
				require.Equal(t, 2, parts[2].GroupID)
				require.Equal(t, 2, parts[3].GroupID)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pkb := &mockGeneratorPropsKeyBuilder{runFn: tt.pkbFn}
			svc := testPartBuilder(t, pkb)

			parts, err := svc.Run(tt.blocks, tt.blockCuboids, tt.microCuboids, model.FaceVisibility{}, *tt.styleIdx, tt.scale)
			if tt.wantErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tt.wantErrMsg)
				return
			}
			require.NoError(t, err)
			require.Len(t, parts, tt.wantCount)

			if tt.wantCheck != nil {
				tt.wantCheck(t, parts)
			}
		})
	}
}
