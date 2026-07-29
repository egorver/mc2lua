package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestVector3(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		vector   Vector3
		wantX    float64
		wantY    float64
		wantZ    float64
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
			name:    "with values",
			face:    ElementFace{UV: [4]float64{0, 0, 16, 16}, Texture: "block/stone"},
			wantUV:  [4]float64{0, 0, 16, 16}, wantTex: "block/stone",
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
		})
	}
}

func TestResolvedBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		block     ResolvedBlock
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
			name: "with textures",
			block: ResolvedBlock{
				Elements: []ModelElement{{}},
				Textures: map[string]string{"#all": "block/stone"},
			},
			wantElems: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Len(t, tt.block.Elements, tt.wantElems)
		})
	}
}
