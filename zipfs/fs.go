package zipfs

import (
	"archive/zip"
	"fmt"
	"io/fs"
)

var _ fs.FS = (*FS)(nil)

type FS struct {
	*zip.Reader
}

// NewFS creates a new zipFS reader
func NewFS(r *zip.Reader) *FS {
	return &FS{
		Reader: r,
	}
}

// Open implements fs.FS Open method
func (fsys *FS) Open(name string) (fs.File, error) {

	if !fs.ValidPath(name) {
		return nil, fmt.Errorf("path `%s` invalid", name)
	}

	f, err := fsys.Reader.Open(name)

	if err != nil {
		return nil, fmt.Errorf("can not open `%s`: %w", name, err)
	}

	return f, nil
}

// ReadDir implements fs.ReadDirFS
func (fsys *FS) ReadDir(name string) ([]fs.DirEntry, error) {

	entries := make([]fs.DirEntry, 0)

	for _, f := range fsys.File {

		if f.FileInfo().IsDir() {
			continue
		}

		entries = append(entries, dirEntry{
			FileInfo: f.FileInfo(),
			path:     f.Name,
		})

	}
	return entries, nil
}

// Stat implements fs.StatFS
func (fsys *FS) Stat(name string) (fs.FileInfo, error) {

	if name == "" {
		return &zipInfo{}, nil
	}

	if !fs.ValidPath(name) {
		return nil, fmt.Errorf("path `%s` invalid", name)
	}

	f, err := fsys.Reader.Open(name)

	if err != nil {
		return nil, fmt.Errorf("can not open `%s`: %w", name, err)
	}

	return f.Stat()
}
