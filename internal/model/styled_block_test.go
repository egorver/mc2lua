package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		color   Color
		wantR   uint8
		wantG   uint8
		wantB   uint8
	}{
		{name: "zero value"},
		{name: "red", color: Color{255, 0, 0}, wantR: 255, wantG: 0, wantB: 0},
		{name: "green", color: Color{0, 255, 0}, wantR: 0, wantG: 255, wantB: 0},
		{name: "blue", color: Color{0, 0, 255}, wantR: 0, wantG: 0, wantB: 255},
		{name: "white", color: Color{255, 255, 255}, wantR: 255, wantG: 255, wantB: 255},
		{name: "black", color: Color{0, 0, 0}, wantR: 0, wantG: 0, wantB: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantR, tt.color[0])
			require.Equal(t, tt.wantG, tt.color[1])
			require.Equal(t, tt.wantB, tt.color[2])
		})
	}
}

func TestStyledElement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		element    StyledElement
		wantFrom   Vector3
		wantTo     Vector3
		wantShade  bool
		wantColor  Color
		wantMat    string
	}{
		{name: "zero value"},
		{
			name: "full cube",
			element: StyledElement{
				From: Vector3{0, 0, 0}, To: Vector3{16, 16, 16}, Shade: true,
				Color: Color{255, 255, 255}, Material: "Stone",
			},
			wantFrom: Vector3{0, 0, 0}, wantTo: Vector3{16, 16, 16},
			wantShade: true, wantColor: Color{255, 255, 255}, wantMat: "Stone",
		},
		{
			name: "with rotation",
			element: StyledElement{
				From:     Vector3{0, 0, 0},
				To:       Vector3{8, 16, 8},
				Rotation: &ElementRotation{Origin: Vector3{8, 8, 8}, Axis: "y", Angle: 45},
				Shade:    false,
				Color:    Color{100, 50, 0},
				Material: "Wood",
			},
			wantFrom: Vector3{0, 0, 0}, wantTo: Vector3{8, 16, 8},
			wantShade: false, wantColor: Color{100, 50, 0}, wantMat: "Wood",
		},
		{
			name: "no material",
			element: StyledElement{
				From: Vector3{0, 0, 0}, To: Vector3{16, 16, 16},
				Color: Color{200, 200, 200},
			},
			wantFrom: Vector3{0, 0, 0}, wantTo: Vector3{16, 16, 16},
			wantColor: Color{200, 200, 200},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantFrom, tt.element.From)
			require.Equal(t, tt.wantTo, tt.element.To)
			require.Equal(t, tt.wantShade, tt.element.Shade)
			require.Equal(t, tt.wantColor, tt.element.Color)
			require.Equal(t, tt.wantMat, tt.element.Material)
		})
	}
}

func TestStyledBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		block      StyledBlock
		wantID     string
		wantProps  string
		wantFull   bool
		wantElems  int
	}{
		{name: "zero value"},
		{
			name: "full block single element",
			block: StyledBlock{
				ID: "minecraft:stone", PropsKey: "", IsFullBlock: true,
				Elements: []StyledElement{
					{From: Vector3{0, 0, 0}, To: Vector3{16, 16, 16}, Shade: true},
				},
			},
			wantID: "minecraft:stone", wantProps: "", wantFull: true, wantElems: 1,
		},
		{
			name: "non-full block with props",
			block: StyledBlock{
				ID: "minecraft:oak_stairs", PropsKey: "facing=north,half=bottom",
				IsFullBlock: false,
				Elements: []StyledElement{
					{From: Vector3{0, 0, 0}, To: Vector3{16, 8, 16}, Shade: true},
					{From: Vector3{0, 8, 0}, To: Vector3{8, 16, 16}, Shade: true},
				},
			},
			wantID: "minecraft:oak_stairs", wantProps: "facing=north,half=bottom",
			wantFull: false, wantElems: 2,
		},
		{
			name: "block with color and material in elements",
			block: StyledBlock{
				ID: "minecraft:grass_block", IsFullBlock: true,
				Elements: []StyledElement{
					{
						From: Vector3{0, 0, 0}, To: Vector3{16, 16, 16}, Shade: true,
						Color: Color{100, 180, 50}, Material: "Grass",
					},
				},
			},
			wantID: "minecraft:grass_block", wantFull: true, wantElems: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantID, tt.block.ID)
			require.Equal(t, tt.wantProps, tt.block.PropsKey)
			require.Equal(t, tt.wantFull, tt.block.IsFullBlock)
			require.Len(t, tt.block.Elements, tt.wantElems)
		})
	}
}
