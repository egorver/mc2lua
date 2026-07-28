package pipeline

import (
	"io"
	"io/fs"
)

type fsApi interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
	Create(name string) (io.WriteCloser, error)
}
