package minecraft

import (
	"io/fs"
	"testing"

	"mc2lua/internal/runtime"

	"github.com/stretchr/testify/require"
)

func TestAssetScannerRun(t *testing.T) {
	t.Run("no directories", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		scanner := NewAssetScanner(mockFS)
		_, err := scanner.Run("assets")
		require.ErrorContains(t, err, "no Minecraft assets found")
	})

	t.Run("no assets in any namespace", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/minecraft", 0755)
		scanner := NewAssetScanner(mockFS)
		_, err := scanner.Run("assets")
		require.ErrorContains(t, err, "no Minecraft assets found")
	})

	t.Run("single namespace with blockstates", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/minecraft", 0755)
		mockFS.AddDir("assets/minecraft/stone", 0755)
		mockFS.AddDir("assets/minecraft/stone/blockstates", 0755)

		scanner := NewAssetScanner(mockFS)
		result, err := scanner.Run("assets")
		require.NoError(t, err)
		require.Equal(t, map[string][]string{"stone": {"assets/minecraft/stone"}}, result)
	})

	t.Run("single namespace with models", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/minecraft", 0755)
		mockFS.AddDir("assets/minecraft/stone", 0755)
		mockFS.AddDir("assets/minecraft/stone/models", 0755)

		scanner := NewAssetScanner(mockFS)
		result, err := scanner.Run("assets")
		require.NoError(t, err)
		require.Equal(t, map[string][]string{"stone": {"assets/minecraft/stone"}}, result)
	})

	t.Run("namespace skipped when no blockstates or models", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/minecraft", 0755)
		mockFS.AddDir("assets/minecraft/stone", 0755)

		scanner := NewAssetScanner(mockFS)
		_, err := scanner.Run("assets")
		require.ErrorContains(t, err, "no Minecraft assets found")
	})

	t.Run("non-directory entries skipped", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddFile("assets/file.txt", []byte("x"), 0644)
		mockFS.AddDir("assets/minecraft", 0755)
		mockFS.AddDir("assets/minecraft/stone", 0755)
		mockFS.AddDir("assets/minecraft/stone/blockstates", 0755)

		scanner := NewAssetScanner(mockFS)
		result, err := scanner.Run("assets")
		require.NoError(t, err)
		require.Equal(t, map[string][]string{"stone": {"assets/minecraft/stone"}}, result)
	})

	t.Run("ReadDir error for mod is skipped", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/minecraft", 0755)
		mockFS.AddDir("assets/minecraft/stone", 0755)
		mockFS.AddDir("assets/minecraft/stone/blockstates", 0755)
		mockFS.ReadDirErrors["assets/minecraft"] = fs.ErrNotExist

		scanner := NewAssetScanner(mockFS)
		_, err := scanner.Run("assets")
		require.ErrorContains(t, err, "no Minecraft assets found")
	})

	t.Run("multiple mods and namespaces", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/minecraft", 0755)
		mockFS.AddDir("assets/minecraft/stone", 0755)
		mockFS.AddDir("assets/minecraft/stone/blockstates", 0755)
		mockFS.AddDir("assets/minecraft/sand", 0755)
		mockFS.AddDir("assets/mod2", 0755)
		mockFS.AddDir("assets/mod2/foo", 0755)
		mockFS.AddDir("assets/mod2/foo/models", 0755)

		scanner := NewAssetScanner(mockFS)
		result, err := scanner.Run("assets")
		require.NoError(t, err)
		require.Equal(t, map[string][]string{
			"stone": {"assets/minecraft/stone"},
			"foo":   {"assets/mod2/foo"},
		}, result)
	})

	t.Run("same namespace from multiple mods appended", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.AddDir("assets/mod1", 0755)
		mockFS.AddDir("assets/mod1/minecraft", 0755)
		mockFS.AddDir("assets/mod1/minecraft/blockstates", 0755)
		mockFS.AddDir("assets/mod2", 0755)
		mockFS.AddDir("assets/mod2/minecraft", 0755)
		mockFS.AddDir("assets/mod2/minecraft/models", 0755)

		scanner := NewAssetScanner(mockFS)
		result, err := scanner.Run("assets")
		require.NoError(t, err)
		require.Len(t, result, 1)
		require.ElementsMatch(t, []string{"assets/mod1/minecraft", "assets/mod2/minecraft"}, result["minecraft"])
	})

	t.Run("outer ReadDir error", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("assets", 0755)
		mockFS.ReadDirErrors["assets"] = fs.ErrNotExist

		scanner := NewAssetScanner(mockFS)
		_, err := scanner.Run("assets")
		require.ErrorContains(t, err, "read assets")
	})
}

func TestAssetScannerCollectNamespaces(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *runtime.FSMock)
		modPath string
		want    map[string][]string
	}{
		{
			name: "ReadDir error returns without modifying map",
			setup: func(m *runtime.FSMock) {
				m.ReadDirErrors["mods/test"] = fs.ErrNotExist
			},
			modPath: "mods/test",
			want:    map[string][]string{},
		},
		{
			name: "non-directory entries skipped",
			setup: func(m *runtime.FSMock) {
				m.AddFile("mods/test/file.txt", nil, 0644)
			},
			modPath: "mods/test",
			want:    map[string][]string{},
		},
		{
			name: "namespace without assets skipped",
			setup: func(m *runtime.FSMock) {
				m.AddDir("mods/test/minecraft", 0755)
			},
			modPath: "mods/test",
			want:    map[string][]string{},
		},
		{
			name: "namespace with blockstates added",
			setup: func(m *runtime.FSMock) {
				m.AddDir("mods/test/minecraft", 0755)
				m.AddDir("mods/test/minecraft/blockstates", 0755)
			},
			modPath: "mods/test",
			want:    map[string][]string{"minecraft": {"mods/test/minecraft"}},
		},
		{
			name: "namespace with models added",
			setup: func(m *runtime.FSMock) {
				m.AddDir("mods/test/minecraft", 0755)
				m.AddDir("mods/test/minecraft/models", 0755)
			},
			modPath: "mods/test",
			want:    map[string][]string{"minecraft": {"mods/test/minecraft"}},
		},
		{
			name: "multiple namespaces",
			setup: func(m *runtime.FSMock) {
				m.AddDir("mods/test/minecraft", 0755)
				m.AddDir("mods/test/minecraft/blockstates", 0755)
				m.AddDir("mods/test/forge", 0755)
				m.AddDir("mods/test/forge/models", 0755)
			},
			modPath: "mods/test",
			want:    map[string][]string{"minecraft": {"mods/test/minecraft"}, "forge": {"mods/test/forge"}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := runtime.NewFSMock()
			tt.setup(m)
			scanner := NewAssetScanner(m)

			got := make(map[string][]string)
			scanner.collectNamespaces(tt.modPath, got)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestAssetScannerHasAssets(t *testing.T) {
	t.Run("has blockstates", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("path/to/ns", 0755)
		mockFS.AddDir("path/to/ns/blockstates", 0755)
		scanner := NewAssetScanner(mockFS)
		require.True(t, scanner.hasAssets("path/to/ns"))
	})

	t.Run("has models", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("path/to/ns", 0755)
		mockFS.AddDir("path/to/ns/models", 0755)
		scanner := NewAssetScanner(mockFS)
		require.True(t, scanner.hasAssets("path/to/ns"))
	})

	t.Run("no blockstates or models", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("path/to/ns", 0755)
		mockFS.AddFile("path/to/ns/random.txt", []byte("x"), 0644)
		scanner := NewAssetScanner(mockFS)
		require.False(t, scanner.hasAssets("path/to/ns"))
	})

	t.Run("empty directory", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.AddDir("path/to/ns", 0755)
		scanner := NewAssetScanner(mockFS)
		require.False(t, scanner.hasAssets("path/to/ns"))
	})

	t.Run("ReadDir error returns false", func(t *testing.T) {
		mockFS := runtime.NewFSMock()
		mockFS.ReadDirErrors["path/to/ns"] = fs.ErrNotExist
		scanner := NewAssetScanner(mockFS)
		require.False(t, scanner.hasAssets("path/to/ns"))
	})
}
