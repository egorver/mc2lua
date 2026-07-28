package runtime

import (
	"io"
	"io/fs"
	"os"
)

type FS struct{}

func NewFS() *FS {
	return &FS{}
}

func (FS) Create(path string) (io.WriteCloser, error)    { return os.Create(path) }
func (FS) ReadDir(dirname string) ([]fs.DirEntry, error) { return os.ReadDir(dirname) }
func (FS) ReadFile(name string) ([]byte, error)          { return os.ReadFile(name) }
