package runtime

import (
	"bytes"
	"io"
	"io/fs"
	"path/filepath"
	"sync"
	"time"
)

type FSMock struct {
	mu             sync.Mutex
	files          map[string]*mockFile
	ReadFileErrors map[string]error
	CreateErrors   map[string]error
	ReadDirErrors  map[string]error
	Calls          []string
}

type mockFile struct {
	name    string
	data    *bytes.Buffer
	mode    fs.FileMode
	modTime time.Time
}

func NewFSMock() *FSMock {
	return &FSMock{
		files:          make(map[string]*mockFile),
		ReadFileErrors: make(map[string]error),
		CreateErrors:   make(map[string]error),
		ReadDirErrors:  make(map[string]error),
	}
}

func (m *FSMock) AddFile(path string, content []byte, mode fs.FileMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	mf := &mockFile{name: path, data: bytes.NewBuffer(content), mode: mode, modTime: time.Now()}
	m.files[filepath.Clean(path)] = mf
}

func (m *FSMock) AddDir(path string, perm fs.FileMode) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.files[filepath.Clean(path)] = &mockFile{
		name:    path,
		data:    bytes.NewBuffer(nil),
		mode:    perm | fs.ModeDir,
		modTime: time.Now(),
	}
}

func (m *FSMock) Record(call string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Calls = append(m.Calls, call)
}

func (m *FSMock) ReadFile(path string) ([]byte, error) {
	m.Record("ReadFile:" + path)
	cpath := filepath.Clean(path)
	if e, ok := m.ReadFileErrors[cpath]; ok && e != nil {
		return nil, e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	if f, ok := m.files[cpath]; ok {
		return f.data.Bytes(), nil
	}
	return nil, fs.ErrNotExist
}

func (m *FSMock) Create(path string) (io.WriteCloser, error) {
	m.Record("Create:" + path)
	cpath := filepath.Clean(path)
	if e, ok := m.CreateErrors[cpath]; ok && e != nil {
		return nil, e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	mf := &mockFile{name: path, data: bytes.NewBuffer(nil), mode: 0644, modTime: time.Now()}
	m.files[cpath] = mf
	return &mockFSWriter{buf: mf.data}, nil
}

func (m *FSMock) ReadDir(dirname string) ([]fs.DirEntry, error) {
	m.Record("ReadDir:" + dirname)
	if e, ok := m.ReadDirErrors[dirname]; ok && e != nil {
		return nil, e
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	var entries []fs.DirEntry
	prefix := filepath.Clean(dirname)
	if prefix == "." || prefix == string(filepath.Separator) {
		prefix = ""
	}
	for p, f := range m.files {
		if prefix == "" {
			entries = append(entries, dirEntry{f.name, f.mode})
			continue
		}
		if filepath.Dir(p) == prefix {
			entries = append(entries, dirEntry{f.name, f.mode})
		}
	}
	return entries, nil
}

type mockFSWriter struct {
	buf *bytes.Buffer
}

func (w *mockFSWriter) Write(p []byte) (int, error) { return w.buf.Write(p) }
func (w *mockFSWriter) Close() error                { return nil }

type dirEntry struct {
	name string
	mode fs.FileMode
}

func (d dirEntry) Name() string               { return filepath.Base(d.name) }
func (d dirEntry) IsDir() bool                { return d.mode.IsDir() }
func (d dirEntry) Type() fs.FileMode          { return d.mode }
func (d dirEntry) Info() (fs.FileInfo, error) { return fileInfo{d.name, 0, d.mode, time.Now()}, nil }

type fileInfo struct {
	name string
	size int64
	mode fs.FileMode
	mod  time.Time
}

func (fi fileInfo) Name() string       { return filepath.Base(fi.name) }
func (fi fileInfo) Size() int64        { return fi.size }
func (fi fileInfo) Mode() fs.FileMode  { return fi.mode }
func (fi fileInfo) ModTime() time.Time { return fi.mod }
func (fi fileInfo) IsDir() bool        { return fi.mode.IsDir() }
func (fi fileInfo) Sys() any           { return nil }

var _ fs.FileInfo = fileInfo{}

func filepathHasPrefix(p, prefix string) bool {
	p = filepath.Clean(p)
	prefix = filepath.Clean(prefix)
	if p == prefix {
		return true
	}
	if len(p) <= len(prefix) {
		return false
	}
	if p[:len(prefix)] != prefix {
		return false
	}
	return p[len(prefix)] == filepath.Separator
}
