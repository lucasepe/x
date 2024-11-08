package zipfs

import (
	"io/fs"
	"time"
)

var (
	_ fs.FileInfo = (*zipInfo)(nil)
)

type zipInfo struct {
	name string
}

// Name return zip name
func (d *zipInfo) Name() string {
	return d.name
}

// IsDir always returns true for zip file
func (d *zipInfo) IsDir() bool {
	return true
}

// Type returns the FileMode type
func (d *zipInfo) Type() fs.FileMode {
	return fs.ModeDir.Perm().Type()
}

// ModTime sumulatate the time modification
func (d *zipInfo) ModTime() time.Time {
	return time.Now()
}

// Mode returns fs.FileMode
func (d *zipInfo) Mode() fs.FileMode {
	return 0755
}

// Sys returns nil
func (d *zipInfo) Sys() any {
	return nil
}

// Size of zip file dir
func (d *zipInfo) Size() int64 {
	return 0
}
