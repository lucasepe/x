package memfs

import (
	"fmt"
	"io"
	"os"
	"time"
)

// File implements the http.File interface
type File struct {
	name     string
	path     string
	at       int64
	bytes    []byte
	children []*File
	modified time.Time
}

func (f *File) Close() error {
	return nil
}

func (f *File) Stat() (os.FileInfo, error) {
	return &FileInfo{f}, nil
}

func (f *File) Readdir(count int) ([]os.FileInfo, error) {
	if f.bytes != nil {
		return nil, fmt.Errorf("reading dir on a regular file %v", f.name)
	}

	fis := []os.FileInfo{}
	for _, file := range f.children {
		fi, err := file.Stat()
		if err != nil {
			return nil, err
		}

		fis = append(fis, fi)
		if count > 0 && len(fis) >= count {
			return fis, nil
		}
	}

	return fis, nil
}

func (f *File) Read(b []byte) (int, error) {
	if f.bytes == nil {
		return 0, fmt.Errorf("reading data on a dir %v", f.name)
	}

	cnt := 0
	for f.at < int64(len(f.bytes)) && cnt < len(b) {
		b[cnt] = f.bytes[f.at]
		cnt++
		f.at++
	}

	if cnt == 0 {
		return 0, io.EOF
	}

	return cnt, nil
}

func (f *File) Seek(offset int64, whence int) (int64, error) {
	if f.bytes == nil {
		return 0, fmt.Errorf("seeking on a dir %v", f.name)
	}

	switch whence {
	case 0:
		f.at = offset
	case 1:
		f.at += offset
	case 2:
		f.at = int64(len(f.bytes)) + offset
	}

	return f.at, nil
}
