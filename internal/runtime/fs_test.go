package runtime

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestFS_NewFS_NotNil(t *testing.T) {
	t.Parallel()
	f := NewFS()
	require.NotNil(t, f)
}

func TestFS_Create_Error(t *testing.T) {
	t.Parallel()
	f := NewFS()

	tests := []struct {
		name string
		path string
	}{
		{
			name: "parent is file",
			path: func() string {
				tmp := t.TempDir()
				fpath := filepath.Join(tmp, "file")
				require.NoError(t, os.WriteFile(fpath, []byte("x"), 0600))
				return filepath.Join(fpath, "sub", "created.txt")
			}(),
		},
		{
			name: "empty path",
			path: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := f.Create(tt.path)
			require.Error(t, err)
		})
	}
}

func TestFS_ReadFile(t *testing.T) {
	t.Parallel()
	f := NewFS()
	td := t.TempDir()
	fn := td + string(os.PathSeparator) + "readfile_test.txt"
	err := os.WriteFile(fn, []byte("readfile content"), 0600)
	require.NoError(t, err)

	b, err := f.ReadFile(fn)
	require.NoError(t, err)
	require.Equal(t, "readfile content", string(b))
}

func TestFS_Create_WriteAndRead(t *testing.T) {
	t.Parallel()
	f := NewFS()
	fn := filepath.Join(t.TempDir(), "created.txt")

	w, err := f.Create(fn)
	require.NoError(t, err)
	_, err = w.Write([]byte("hello world"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	b, err := f.ReadFile(fn)
	require.NoError(t, err)
	require.Equal(t, "hello world", string(b))
}

func TestFS_ReadDir(t *testing.T) {
	t.Parallel()
	f := NewFS()

	tests := []struct {
		name    string
		setup   func(t *testing.T) string
		wantErr bool
		want    []string
	}{
		{
			name: "mixed files and subdirs",
			setup: func(t *testing.T) string {
				td := t.TempDir()
				require.NoError(t, os.WriteFile(filepath.Join(td, "a.txt"), []byte("a"), 0600))
				require.NoError(t, os.MkdirAll(filepath.Join(td, "sub"), 0755))
				return td
			},
			want: []string{"a.txt", "sub"},
		},
		{
			name: "empty directory",
			setup: func(t *testing.T) string {
				return t.TempDir()
			},
			want: []string{},
		},
		{
			name: "non-existent directory",
			setup: func(t *testing.T) string {
				return filepath.Join(t.TempDir(), "nonexistent")
			},
			wantErr: true,
		},
		{
			name: "path is file",
			setup: func(t *testing.T) string {
				fn := filepath.Join(t.TempDir(), "file.txt")
				require.NoError(t, os.WriteFile(fn, []byte("x"), 0600))
				return fn
			},
			wantErr: true,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			dir := tc.setup(t)
			entries, err := f.ReadDir(dir)
			if tc.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			var names []string
			for _, e := range entries {
				names = append(names, e.Name())
			}
			require.ElementsMatch(t, tc.want, names)
		})
	}
}
