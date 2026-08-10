package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVector3(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		vector Vector3
		wantX  float64
		wantY  float64
		wantZ  float64
	}{
		{name: "zero", vector: Vector3{0, 0, 0}},
		{name: "positive", vector: Vector3{1.5, 2.5, 3.5}, wantX: 1.5, wantY: 2.5, wantZ: 3.5},
		{name: "negative", vector: Vector3{-1, -2, -3}, wantX: -1, wantY: -2, wantZ: -3},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantX, tt.vector[0])
			require.Equal(t, tt.wantY, tt.vector[1])
			require.Equal(t, tt.wantZ, tt.vector[2])
		})
	}
}

func TestElementRotation(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		rotation ElementRotation
		wantOrig Vector3
		wantAxis string
		wantAng  float64
		wantRes  bool
	}{
		{name: "zero value"},
		{
			name:     "with values",
			rotation: ElementRotation{Origin: Vector3{8, 8, 8}, Axis: "y", Angle: 45, Rescale: true},
			wantOrig: Vector3{8, 8, 8}, wantAxis: "y", wantAng: 45, wantRes: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantOrig, tt.rotation.Origin)
			require.Equal(t, tt.wantAxis, tt.rotation.Axis)
			require.Equal(t, tt.wantAng, tt.rotation.Angle)
			require.Equal(t, tt.wantRes, tt.rotation.Rescale)
		})
	}
}

func TestElementFace(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		face    ElementFace
		wantUV  [4]float64
		wantTex string
	}{
		{name: "zero value"},
		{
			name:   "with values",
			face:   ElementFace{UV: [4]float64{0, 0, 16, 16}, Texture: "block/stone"},
			wantUV: [4]float64{0, 0, 16, 16}, wantTex: "block/stone",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantUV, tt.face.UV)
			require.Equal(t, tt.wantTex, tt.face.Texture)
		})
	}
}

func TestModelElement(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		element   ModelElement
		wantFrom  Vector3
		wantTo    Vector3
		wantShade bool
	}{
		{name: "zero value"},
		{
			name: "full block", element: ModelElement{From: Vector3{0, 0, 0}, To: Vector3{16, 16, 16}, Shade: true},
			wantFrom: Vector3{0, 0, 0}, wantTo: Vector3{16, 16, 16}, wantShade: true,
		},
		{
			name: "no shade", element: ModelElement{From: Vector3{0, 0, 0}, To: Vector3{16, 16, 16}, Shade: false},
			wantFrom: Vector3{0, 0, 0}, wantTo: Vector3{16, 16, 16}, wantShade: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantFrom, tt.element.From)
			require.Equal(t, tt.wantTo, tt.element.To)
			require.Equal(t, tt.wantShade, tt.element.Shade)
			require.Nil(t, tt.element.Rotation)
			require.Nil(t, tt.element.Faces)
		})
	}
}

func TestModelElement_RotationAndFaces(t *testing.T) {
	t.Parallel()

	rot := &ElementRotation{Origin: Vector3{8, 8, 8}, Axis: "x", Angle: 22.5, Rescale: true}
	faces := map[string]ElementFace{
		"up":    {UV: [4]float64{0, 0, 16, 16}, Texture: "block/stone"},
		"down":  {UV: [4]float64{0, 0, 16, 16}, Texture: "block/stone"},
		"north": {UV: [4]float64{0, 0, 16, 16}, Texture: "block/stone"},
	}
	element := ModelElement{
		From:     Vector3{0, 0, 0},
		To:       Vector3{16, 16, 16},
		Rotation: rot,
		Shade:    true,
		Faces:    faces,
	}

	require.Equal(t, rot, element.Rotation)
	require.Equal(t, faces, element.Faces)
	require.Len(t, element.Faces, 3)
	require.Equal(t, ElementFace{UV: [4]float64{0, 0, 16, 16}, Texture: "block/stone"}, element.Faces["up"])
}

func TestResolvedBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		block     ResolvedBlock
		wantID    string
		wantProps string
		wantElems int
	}{
		{name: "zero value"},
		{
			name: "full block", block: ResolvedBlock{Elements: []ModelElement{{}}},
			wantElems: 1,
		},
		{
			name: "not full block", block: ResolvedBlock{Elements: []ModelElement{{}, {}}},
			wantElems: 2,
		},
		{
			name: "with id and props",
			block: ResolvedBlock{
				ID: "minecraft:stone", PropsKey: "variant=andesite",
				Elements: []ModelElement{{}},
			},
			wantID: "minecraft:stone", wantProps: "variant=andesite", wantElems: 1,
		},
		{
			name: "with textures",
			block: ResolvedBlock{
				ID:       "minecraft:stone",
				Elements: []ModelElement{{}},
				Textures: map[string]string{"#all": "block/stone"},
			},
			wantID: "minecraft:stone", wantElems: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.wantID, tt.block.ID)
			require.Equal(t, tt.wantProps, tt.block.PropsKey)
			require.Len(t, tt.block.Elements, tt.wantElems)
		})
	}
}

func TestResolvedBlock_Textures(t *testing.T) {
	t.Parallel()

	block := ResolvedBlock{
		ID:       "minecraft:stone",
		PropsKey: "",
		Textures: map[string]string{"#all": "block/stone", "#top": "block/stone_top"},
	}
	require.Len(t, block.Textures, 2)
	require.Equal(t, "block/stone", block.Textures["#all"])
	require.Equal(t, "block/stone_top", block.Textures["#top"])
}

func TestResolvedBlock_Rotations(t *testing.T) {
	t.Parallel()

	block := ResolvedBlock{ID: "minecraft:oak_log", RotX: 90, RotY: 180}
	require.Equal(t, 90.0, block.RotX)
	require.Equal(t, 180.0, block.RotY)
}
