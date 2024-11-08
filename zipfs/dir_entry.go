package zipfs

import (
	"io/fs"
)

var (
	_ fs.DirEntry = (*dirEntry)(nil)
)

type dirEntry struct {
	fs.FileInfo
	path string
}

// IsDir check if an entry is directory
func (d dirEntry) IsDir() bool {
	return d.FileInfo.IsDir()
}

// Type implements fs.FileMode
func (d dirEntry) Type() fs.FileMode {
	return d.FileInfo.Mode().Type()
}

// Info returns the entry fs.FileInfo
func (d dirEntry) Info() (fs.FileInfo, error) {
	return d.FileInfo, nil
}

// Name returns the entry path
func (d dirEntry) Name() string {
	return d.path
}
