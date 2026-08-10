package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFaceIndexConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  int
		want int
	}{
		{name: "top", got: FaceIndexTop, want: 0},
		{name: "bottom", got: FaceIndexBottom, want: 1},
		{name: "front", got: FaceIndexFront, want: 2},
		{name: "back", got: FaceIndexBack, want: 3},
		{name: "left", got: FaceIndexLeft, want: 4},
		{name: "right", got: FaceIndexRight, want: 5},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestFaceMask_ZeroValue(t *testing.T) {
	t.Parallel()

	var m FaceMask
	require.Len(t, m, 6)
	for i, v := range m {
		require.False(t, v, "index %d should be false", i)
	}
}

func TestFaceMask_AllFaces(t *testing.T) {
	t.Parallel()

	var m FaceMask
	m[FaceIndexTop] = true
	m[FaceIndexBottom] = true
	m[FaceIndexFront] = true
	m[FaceIndexBack] = true
	m[FaceIndexLeft] = true
	m[FaceIndexRight] = true

	for i, v := range m {
		require.True(t, v, "index %d should be true", i)
	}
}

func TestFaceMask_PartialFaces(t *testing.T) {
	t.Parallel()

	m := FaceMask{FaceIndexTop: true, FaceIndexRight: true}
	require.True(t, m[FaceIndexTop])
	require.True(t, m[FaceIndexRight])
	require.False(t, m[FaceIndexBottom])
	require.False(t, m[FaceIndexFront])
	require.False(t, m[FaceIndexBack])
	require.False(t, m[FaceIndexLeft])
}

func TestFaceMask_LiteralCompare(t *testing.T) {
	t.Parallel()

	require.Equal(t, FaceMask{true, false, false, false, false, false}, FaceMask{FaceIndexTop: true})
}

func TestFaceVisibility_ZeroValue(t *testing.T) {
	t.Parallel()

	var v FaceVisibility
	require.Nil(t, v.BlockFaces)
	require.Nil(t, v.MicroFaces)
	require.Nil(t, v.ComplexFaces)
}

func TestFaceVisibility_FullInit(t *testing.T) {
	t.Parallel()

	v := FaceVisibility{
		BlockFaces:   []FaceMask{{FaceIndexTop: true}},
		MicroFaces:   []FaceMask{{FaceIndexLeft: true}, {FaceIndexRight: true}},
		ComplexFaces: []FaceMask{},
	}
	require.Len(t, v.BlockFaces, 1)
	require.True(t, v.BlockFaces[0][FaceIndexTop])
	require.Len(t, v.MicroFaces, 2)
	require.True(t, v.MicroFaces[0][FaceIndexLeft])
	require.True(t, v.MicroFaces[1][FaceIndexRight])
	require.Empty(t, v.ComplexFaces)
}

func TestFaceVisibility_Mixed(t *testing.T) {
	t.Parallel()

	v := FaceVisibility{
		BlockFaces:   []FaceMask{{}, {FaceIndexBottom: true}},
		MicroFaces:   nil,
		ComplexFaces: []FaceMask{{FaceIndexFront: true, FaceIndexBack: true}},
	}
	require.Len(t, v.BlockFaces, 2)
	require.False(t, v.BlockFaces[0][FaceIndexBottom])
	require.True(t, v.BlockFaces[1][FaceIndexBottom])
	require.Nil(t, v.MicroFaces)
	require.Len(t, v.ComplexFaces, 1)
	require.True(t, v.ComplexFaces[0][FaceIndexFront])
	require.True(t, v.ComplexFaces[0][FaceIndexBack])
}
