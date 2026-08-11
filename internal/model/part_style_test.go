package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFaceConstants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "top", got: FaceTop, want: "top"},
		{name: "bottom", got: FaceBottom, want: "bottom"},
		{name: "front", got: FaceFront, want: "front"},
		{name: "back", got: FaceBack, want: "back"},
		{name: "left", got: FaceLeft, want: "left"},
		{name: "right", got: FaceRight, want: "right"},
		{name: "sides", got: FaceSides, want: "sides"},
		{name: "faces", got: FaceFaces, want: "faces"},
		{name: "walls", got: FaceWalls, want: "walls"},
		{name: "all", got: FaceAll, want: "all"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			require.Equal(t, tt.want, tt.got)
		})
	}
}

func TestSurface(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		var s Surface
		require.Empty(t, s.Texture)
		require.Nil(t, s.Color)
		require.Nil(t, s.Transparency)
	})

	t.Run("full surface", func(t *testing.T) {
		alpha := 0.5
		s := Surface{
			Texture:      "rbxassetid://186400918",
			Color:        &Color{191, 134, 60},
			Transparency: &alpha,
		}
		require.Equal(t, "rbxassetid://186400918", s.Texture)
		require.Equal(t, &Color{191, 134, 60}, s.Color)
		require.Equal(t, &alpha, s.Transparency)
	})

	t.Run("texture only", func(t *testing.T) {
		s := Surface{Texture: "rbxassetid://186400919"}
		require.Equal(t, "rbxassetid://186400919", s.Texture)
		require.Nil(t, s.Color)
		require.Nil(t, s.Transparency)
	})

	t.Run("color only", func(t *testing.T) {
		s := Surface{Color: &Color{0, 0, 0}}
		require.Equal(t, &Color{0, 0, 0}, s.Color)
		require.Empty(t, s.Texture)
		require.Nil(t, s.Transparency)
	})
}

func TestPartStyle(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		var ps PartStyle
		require.Nil(t, ps.Color)
		require.Nil(t, ps.Transparency)
		require.Nil(t, ps.Top)
		require.Nil(t, ps.Bottom)
		require.Nil(t, ps.Front)
		require.Nil(t, ps.Back)
		require.Nil(t, ps.Left)
		require.Nil(t, ps.Right)
		require.Nil(t, ps.Sides)
		require.Nil(t, ps.Faces)
		require.Nil(t, ps.Walls)
		require.Nil(t, ps.All)
	})

	t.Run("full style", func(t *testing.T) {
		alpha := 0.25
		ps := PartStyle{
			Color:        &Color{38, 94, 173},
			Transparency: &alpha,
			Top:          &Surface{Texture: "rbxassetid://1"},
			Bottom:       &Surface{Texture: "rbxassetid://2"},
			Front:        &Surface{Texture: "rbxassetid://3"},
			Back:         &Surface{Texture: "rbxassetid://4"},
			Left:         &Surface{Texture: "rbxassetid://5"},
			Right:        &Surface{Texture: "rbxassetid://6"},
			Sides:        &Surface{Texture: "rbxassetid://7"},
			Faces:        &Surface{Texture: "rbxassetid://8"},
			Walls:        &Surface{Texture: "rbxassetid://9"},
			All:          &Surface{Texture: "rbxassetid://10"},
		}
		require.Equal(t, &Color{38, 94, 173}, ps.Color)
		require.Equal(t, &alpha, ps.Transparency)
		require.NotNil(t, ps.Top)
		require.NotNil(t, ps.Bottom)
		require.NotNil(t, ps.Front)
		require.NotNil(t, ps.Back)
		require.NotNil(t, ps.Left)
		require.NotNil(t, ps.Right)
		require.NotNil(t, ps.Sides)
		require.NotNil(t, ps.Faces)
		require.NotNil(t, ps.Walls)
		require.NotNil(t, ps.All)
	})
}
