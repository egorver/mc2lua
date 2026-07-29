package minecraft

import "io/fs"

type fsApi interface {
	ReadFile(name string) ([]byte, error)
	ReadDir(name string) ([]fs.DirEntry, error)
}
