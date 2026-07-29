package minecraft

import (
	"io/fs"
	"testing"

	"mc2lua/internal/runtime"

	"github.com/stretchr/testify/require"
)

func TestAssetScannerRun(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(m *runtime.FSMock)
		want    map[string][]string
		wantErr string
	}{
		{
			name:    "no directories",
			setup:   func(m *runtime.FSMock) {},
			wantErr: "no Minecraft assets found",
		},
		{
			name: "no assets in any namespace",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/minecraft", 0755)
			},
			wantErr: "no Minecraft assets found",
		},
		{
			name: "single namespace with blockstates",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/minecraft", 0755)
				m.AddDir("assets/minecraft/stone", 0755)
				m.AddDir("assets/minecraft/stone/blockstates", 0755)
			},
			want: map[string][]string{"stone": {"assets/minecraft/stone"}},
		},
		{
			name: "single namespace with models",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/minecraft", 0755)
				m.AddDir("assets/minecraft/stone", 0755)
				m.AddDir("assets/minecraft/stone/models", 0755)
			},
			want: map[string][]string{"stone": {"assets/minecraft/stone"}},
		},
		{
			name: "namespace skipped when no blockstates or models",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/minecraft", 0755)
				m.AddDir("assets/minecraft/stone", 0755)
			},
			wantErr: "no Minecraft assets found",
		},
		{
			name: "non-directory entries skipped",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddFile("assets/file.txt", []byte("x"), 0644)
				m.AddDir("assets/minecraft", 0755)
				m.AddDir("assets/minecraft/stone", 0755)
				m.AddDir("assets/minecraft/stone/blockstates", 0755)
			},
			want: map[string][]string{"stone": {"assets/minecraft/stone"}},
		},
		{
			name: "ReadDir error for mod is skipped",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/minecraft", 0755)
				m.AddDir("assets/minecraft/stone", 0755)
				m.AddDir("assets/minecraft/stone/blockstates", 0755)
				m.ReadDirErrors["assets/minecraft"] = fs.ErrNotExist
			},
			wantErr: "no Minecraft assets found",
		},
		{
			name: "multiple mods and namespaces",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/minecraft", 0755)
				m.AddDir("assets/minecraft/stone", 0755)
				m.AddDir("assets/minecraft/stone/blockstates", 0755)
				m.AddDir("assets/minecraft/sand", 0755)
				m.AddDir("assets/mod2", 0755)
				m.AddDir("assets/mod2/foo", 0755)
				m.AddDir("assets/mod2/foo/models", 0755)
			},
			want: map[string][]string{
				"stone": {"assets/minecraft/stone"},
				"foo":   {"assets/mod2/foo"},
			},
		},
		{
			name: "same namespace from multiple mods appended",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.AddDir("assets/mod1", 0755)
				m.AddDir("assets/mod1/minecraft", 0755)
				m.AddDir("assets/mod1/minecraft/blockstates", 0755)
				m.AddDir("assets/mod2", 0755)
				m.AddDir("assets/mod2/minecraft", 0755)
				m.AddDir("assets/mod2/minecraft/models", 0755)
			},
			want: map[string][]string{"minecraft": {"assets/mod1/minecraft", "assets/mod2/minecraft"}},
		},
		{
			name: "outer ReadDir error",
			setup: func(m *runtime.FSMock) {
				m.AddDir("assets", 0755)
				m.ReadDirErrors["assets"] = fs.ErrNotExist
			},
			wantErr: "read assets",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := runtime.NewFSMock()
			tt.setup(mockFS)
			scanner := NewAssetScanner(mockFS)
			result, err := scanner.Run("assets")
			if tt.wantErr != "" {
				require.ErrorContains(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			if len(tt.want) == 1 {
				for k, v := range tt.want {
					require.Len(t, result, 1)
					require.ElementsMatch(t, v, result[k])
				}
			} else {
				require.Equal(t, tt.want, result)
			}
		})
	}
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
	tests := []struct {
		name  string
		setup func(m *runtime.FSMock)
		want  bool
	}{
		{
			name: "has blockstates",
			setup: func(m *runtime.FSMock) {
				m.AddDir("path/to/ns", 0755)
				m.AddDir("path/to/ns/blockstates", 0755)
			},
			want: true,
		},
		{
			name: "has models",
			setup: func(m *runtime.FSMock) {
				m.AddDir("path/to/ns", 0755)
				m.AddDir("path/to/ns/models", 0755)
			},
			want: true,
		},
		{
			name: "no blockstates or models",
			setup: func(m *runtime.FSMock) {
				m.AddDir("path/to/ns", 0755)
				m.AddFile("path/to/ns/random.txt", []byte("x"), 0644)
			},
			want: false,
		},
		{
			name: "empty directory",
			setup: func(m *runtime.FSMock) {
				m.AddDir("path/to/ns", 0755)
			},
			want: false,
		},
		{
			name: "ReadDir error returns false",
			setup: func(m *runtime.FSMock) {
				m.ReadDirErrors["path/to/ns"] = fs.ErrNotExist
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockFS := runtime.NewFSMock()
			tt.setup(mockFS)
			scanner := NewAssetScanner(mockFS)
			got := scanner.hasAssets("path/to/ns")
			require.Equal(t, tt.want, got)
		})
	}
}
