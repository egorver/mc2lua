package model

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPart(t *testing.T) {
	t.Parallel()

	t.Run("zero value", func(t *testing.T) {
		var p Part
		require.Nil(t, p.Transparency)
		require.Nil(t, p.Top)
		require.Nil(t, p.Bottom)
		require.Nil(t, p.Front)
		require.Nil(t, p.Back)
		require.Nil(t, p.Left)
		require.Nil(t, p.Right)
	})

	t.Run("full part", func(t *testing.T) {
		alpha := 0.25
		p := Part{
			Name:         "stone",
			Group:        "grp",
			GroupID:      1,
			BlockID:      "minecraft:stone",
			PropsKey:     "props",
			Size:         Vector3{1, 2, 3},
			Position:     Vector3{4, 5, 6},
			Color:        Color{191, 134, 60},
			Material:     "Stone",
			Transparency: &alpha,
			Top:          &Surface{Texture: "rbxassetid://1"},
			Bottom:       &Surface{Texture: "rbxassetid://2"},
			Front:        &Surface{Texture: "rbxassetid://3"},
			Back:         &Surface{Texture: "rbxassetid://4"},
			Left:         &Surface{Texture: "rbxassetid://5"},
			Right:        &Surface{Texture: "rbxassetid://6"},
		}
		require.Equal(t, "stone", p.Name)
		require.Equal(t, "grp", p.Group)
		require.Equal(t, 1, p.GroupID)
		require.Equal(t, "minecraft:stone", p.BlockID)
		require.Equal(t, "props", p.PropsKey)
		require.Equal(t, Vector3{1, 2, 3}, p.Size)
		require.Equal(t, Vector3{4, 5, 6}, p.Position)
		require.Equal(t, Color{191, 134, 60}, p.Color)
		require.Equal(t, "Stone", p.Material)
		require.Equal(t, &alpha, p.Transparency)
		require.Equal(t, &Surface{Texture: "rbxassetid://1"}, p.Top)
		require.Equal(t, &Surface{Texture: "rbxassetid://2"}, p.Bottom)
		require.Equal(t, &Surface{Texture: "rbxassetid://3"}, p.Front)
		require.Equal(t, &Surface{Texture: "rbxassetid://4"}, p.Back)
		require.Equal(t, &Surface{Texture: "rbxassetid://5"}, p.Left)
		require.Equal(t, &Surface{Texture: "rbxassetid://6"}, p.Right)
	})
}
