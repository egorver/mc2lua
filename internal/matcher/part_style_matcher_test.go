package matcher

import (
	"testing"

	"mc2lua/internal/model"

	"github.com/stretchr/testify/require"
)

func TestPartStyleMatcher_New(t *testing.T) {
	t.Parallel()

	t.Run("valid config", func(t *testing.T) {
		mock := &mockFS{data: []byte(`
parts:
  minecraft:oak_log:
    top:
      texture: rbxassetid://186400918
      color: [191, 134, 60]
    faces:
      texture: rbxassetid://186400919
`)}
		m, err := NewPartStyleMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("empty config", func(t *testing.T) {
		mock := &mockFS{data: []byte("")}
		m, err := NewPartStyleMatcher(mock, "test.yaml")
		require.NoError(t, err)
		require.NotNil(t, m)
	})

	t.Run("invalid yaml", func(t *testing.T) {
		mock := &mockFS{data: []byte("{{{")}
		_, err := NewPartStyleMatcher(mock, "test.yaml")
		require.Error(t, err)
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := NewPartStyleMatcher(&mockFS{data: nil}, "test.yaml")
		require.Error(t, err)
	})
}

func TestPartStyleMatcher_Run(t *testing.T) {
	t.Parallel()

	mock := &mockFS{data: []byte(`
parts:
  minecraft:oak_log:
    top:
      texture: rbxassetid://186400918
      color: [191, 134, 60]
    faces:
      texture: rbxassetid://186400919
  minecraft:glass:
    transparency: 0.5
    all:
      texture: rbxassetid://186400927
      color: [191, 191, 191]
  minecraft:water:
    color: [38, 94, 173]
  minecraft:tuff_bricks:
    all:
      texture: rbxassetid://186400940
      studs_per_tile: 12
`)}
	m, err := NewPartStyleMatcher(mock, "test.yaml")
	require.NoError(t, err)

	t.Run("oak log", func(t *testing.T) {
		got, ok := m.Run("minecraft:oak_log")
		require.True(t, ok)
		require.NotNil(t, got.Top)
		require.Equal(t, "rbxassetid://186400918", got.Top.Texture)
		require.Equal(t, &model.Color{191, 134, 60}, got.Top.Color)
		require.Nil(t, got.Bottom)
		require.NotNil(t, got.Faces)
	})

	t.Run("glass", func(t *testing.T) {
		got, ok := m.Run("minecraft:glass")
		require.True(t, ok)
		require.NotNil(t, got.Transparency)
		require.Equal(t, 0.5, *got.Transparency)
		require.NotNil(t, got.All)
	})

	t.Run("water", func(t *testing.T) {
		got, ok := m.Run("minecraft:water")
		require.True(t, ok)
		require.Equal(t, &model.Color{38, 94, 173}, got.Color)
		require.Nil(t, got.All)
	})

	t.Run("tuff bricks", func(t *testing.T) {
		got, ok := m.Run("minecraft:tuff_bricks")
		require.True(t, ok)
		require.NotNil(t, got.All)
		require.Equal(t, 12.0, *got.All.StudsPerTile)
	})

	t.Run("unlisted block", func(t *testing.T) {
		_, ok := m.Run("minecraft:stone")
		require.False(t, ok)
	})

	t.Run("custom mod block", func(t *testing.T) {
		_, ok := m.Run("cobblemon:gilded_chest")
		require.False(t, ok)
	})
}
