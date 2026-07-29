package runtime

import (
	"errors"
	"io"
	"io/fs"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"
)

var (
	errCreate   = errors.New("create fail")
	errReadDir  = errors.New("readdir fail")
	errReadFile = errors.New("read fail")
)

func TestFSMock_Create(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *FSMock)
		path    string
		wantErr error
		check   func(t *testing.T, m *FSMock, w io.WriteCloser)
	}{
		{
			name: "ok",
			path: "/new.txt",
			check: func(t *testing.T, m *FSMock, w io.WriteCloser) {
				n, err := w.Write([]byte("created"))
				require.NoError(t, err)
				require.Equal(t, 7, n)
				require.NoError(t, w.Close())

				data, err := m.ReadFile("/new.txt")
				require.NoError(t, err)
				require.Equal(t, "created", string(data))
			},
		},
		{
			name:    "create_error",
			setup:   func(m *FSMock) { m.CreateErrors[filepath.Clean("/bad")] = errCreate },
			path:    "/bad",
			wantErr: errCreate,
		},
		{
			name: "write_then_close_ok",
			path: "/ok.txt",
			check: func(t *testing.T, m *FSMock, w io.WriteCloser) {
				_, err := w.Write([]byte("hello"))
				require.NoError(t, err)
				require.NoError(t, w.Close())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := NewFSMock()
			if tc.setup != nil {
				tc.setup(m)
			}
			w, err := m.Create(tc.path)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if tc.check != nil {
				tc.check(t, m, w)
			}
		})
	}
}

func TestFSMock_ReadDir(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *FSMock)
		dirname string
		wantErr error
		check   func(t *testing.T, entries []fs.DirEntry)
	}{
		{
			name: "flat",
			setup: func(m *FSMock) {
				m.AddFile("/dir/a.txt", []byte("a"), 0644)
				m.AddFile("/dir/b.txt", []byte("b"), 0644)
			},
			dirname: "/dir",
			check: func(t *testing.T, entries []fs.DirEntry) {
				require.Len(t, entries, 2)
			},
		},
		{
			name: "nested",
			setup: func(m *FSMock) {
				m.AddFile("/root/a.txt", []byte("a"), 0644)
				m.AddFile("/root/sub/b.txt", []byte("b"), 0644)
			},
			dirname: "/root",
			check: func(t *testing.T, entries []fs.DirEntry) {
				require.Len(t, entries, 1)
				require.Equal(t, "a.txt", entries[0].Name())
			},
		},
		{
			name:    "root",
			setup:   func(m *FSMock) { m.AddFile("/f.txt", []byte("x"), 0644) },
			dirname: "/",
			check: func(t *testing.T, entries []fs.DirEntry) {
				require.Len(t, entries, 1)
				require.Equal(t, "f.txt", entries[0].Name())
			},
		},
		{
			name:    "dot",
			setup:   func(m *FSMock) { m.AddFile("/a.txt", []byte("x"), 0644) },
			dirname: ".",
			check: func(t *testing.T, entries []fs.DirEntry) {
				require.Len(t, entries, 1)
			},
		},
		{
			name:    "error",
			setup:   func(m *FSMock) { m.ReadDirErrors["/bad"] = errReadDir },
			dirname: "/bad",
			wantErr: errReadDir,
		},
		{
			name:    "dir_entry_info",
			setup:   func(m *FSMock) { m.AddFile("/dir/f.txt", []byte("x"), 0644) },
			dirname: "/dir",
			check: func(t *testing.T, entries []fs.DirEntry) {
				require.Len(t, entries, 1)
				fi, err := entries[0].Info()
				require.NoError(t, err)
				require.Equal(t, "f.txt", fi.Name())
				require.False(t, fi.IsDir())
			},
		},
		{
			name:    "dir_entry_type",
			setup:   func(m *FSMock) { m.AddFile("/dir/sub", nil, fs.ModeDir|0755) },
			dirname: "/dir",
			check: func(t *testing.T, entries []fs.DirEntry) {
				require.Len(t, entries, 1)
				require.True(t, entries[0].IsDir())
				require.True(t, entries[0].Type().IsDir())
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := NewFSMock()
			if tc.setup != nil {
				tc.setup(m)
			}
			entries, err := m.ReadDir(tc.dirname)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			if tc.check != nil {
				tc.check(t, entries)
			}
		})
	}
}

func TestFSMock_ReadFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		setup   func(m *FSMock)
		path    string
		want    string
		wantErr error
	}{
		{
			name:  "ok",
			setup: func(m *FSMock) { m.AddFile("/tmp/test.json", []byte(`{"key": "value"}`), 0644) },
			path:  "/tmp/test.json",
			want:  `{"key": "value"}`,
		},
		{
			name:    "not_found",
			path:    "/nonexistent",
			wantErr: fs.ErrNotExist,
		},
		{
			name:    "error",
			setup:   func(m *FSMock) { m.ReadFileErrors[filepath.Clean("/bad")] = errReadFile },
			path:    "/bad",
			wantErr: errReadFile,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			m := NewFSMock()
			if tc.setup != nil {
				tc.setup(m)
			}
			data, err := m.ReadFile(tc.path)
			if tc.wantErr != nil {
				require.ErrorIs(t, err, tc.wantErr)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tc.want, string(data))
		})
	}
}

func TestFSMock_Create_Overwrite(t *testing.T) {
	t.Parallel()

	m := NewFSMock()
	m.AddFile("/existing.txt", []byte("original"), 0644)

	w, err := m.Create("/existing.txt")
	require.NoError(t, err)
	_, err = w.Write([]byte("overwritten"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	data, err := m.ReadFile("/existing.txt")
	require.NoError(t, err)
	require.Equal(t, "overwritten", string(data))
}

func TestFSMock_CallsRecording(t *testing.T) {
	t.Parallel()

	m := NewFSMock()
	m.AddFile("/f", []byte("x"), 0644)

	_, _ = m.Create("/new")
	_, _ = m.ReadDir("/d")
	_, _ = m.ReadFile("/f")

	require.Equal(t, []string{
		"Create:/new",
		"ReadDir:/d",
		"ReadFile:/f",
	}, m.Calls)
}

func TestFSMock_FilepathHasPrefix(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		p      string
		prefix string
		want   bool
	}{
		{name: "exact match", p: "/a/b/c", prefix: "/a/b/c", want: true},
		{name: "prefix match", p: "/a/b/c/d", prefix: "/a/b/c", want: true},
		{name: "no match", p: "/x/y/z", prefix: "/a/b/c", want: false},
		{name: "shorter path no match", p: "/a", prefix: "/a/b/c", want: false},
		{name: "prefix boundary", p: "/a/b/ccc", prefix: "/a/b/c", want: false},
		{name: "empty path", p: "", prefix: "", want: true},
		{name: "empty prefix", p: "/a", prefix: "", want: false},
		{name: "root prefix", p: "/a", prefix: "/", want: false},
		{name: "root path root prefix", p: "/", prefix: "/", want: true},
		{name: "same prefix", p: "/a/b/c", prefix: "/a/b/c/d", want: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := filepathHasPrefix(tt.p, tt.prefix)
			require.Equal(t, tt.want, got)
		})
	}
}
