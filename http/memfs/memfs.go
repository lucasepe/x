package memfs

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

func New(files map[string][]byte) (http.FileSystem, error) {
	root, err := buildTree(files)
	if err != nil {
		return nil, err
	}

	fs := &memFs{}
	fs.root.Store(root)
	return fs, nil
}

var _ http.FileSystem = (*memFs)(nil)

// memFs implements the http.FileSystem interface
type memFs struct {
	root atomic.Value
}

func (f *memFs) Open(path string) (http.File, error) {
	path = filepath.Clean(path)

	if path == "/" || path == "" {
		return f.root.Load().(*File), nil
	}

	parts := strings.Split(path, "/")
	if len(parts[0]) == 0 {
		parts = parts[1:]
	}

	parent := f.root.Load().(*File)
	for _, part := range parts {
		found := false

		for _, child := range parent.children {
			if child.name == part {
				parent = child
				found = true
				break
			}
		}

		if !found {
			return nil, os.ErrNotExist
		}
	}

	return parent, nil
}

func buildTree(fs map[string][]byte) (*File, error) {
	type item struct {
		path  string
		bytes []byte
		isDir bool
	}

	items := []*item{}
	for k, v := range fs {
		isDir := strings.HasSuffix(k, "/")
		if isDir && v != nil {
			return nil, fmt.Errorf("dir path cannot have file content %v", k)
		}
		if !isDir && v == nil {
			return nil, fmt.Errorf("file path cannot have nil content %v", k)
		}
		path := filepath.Clean(k)

		items = append(items, &item{
			path:  path,
			bytes: v,
			isDir: isDir,
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		return items[i].path < items[j].path
	})

	ts := time.Now()
	lastPath := ""
	root := &File{
		name:     "/",
		path:     "/",
		modified: ts,
	}
	for _, item := range items {
		if item.path == lastPath {
			return nil, fmt.Errorf("duplicated path %v", lastPath)
		}
		lastPath = item.path

		parts := strings.Split(item.path, "/")
		if len(parts[0]) == 0 {
			parts = parts[1:]
		}

		parent := root
		idx := 0
		for ; idx < len(parts); idx++ {
			found := false
			for _, child := range parent.children {
				if child.name == parts[idx] {
					parent = child
					found = true
					break
				}
			}

			if !found {
				break
			}
		}
		if idx >= len(parts) {
			return nil, fmt.Errorf("unexpected path error")
		}
		if parent.bytes != nil {
			return nil, fmt.Errorf("node cannot be dir and data file at the same time: %v", parent.name)
		}

		for ; idx < len(parts); idx++ {
			child := &File{
				name:     parts[idx],
				path:     "/" + strings.Join(parts[:idx+1], "/"),
				modified: ts,
			}

			parent.children = append(parent.children, child)
			parent = child
		}
		parent.bytes = item.bytes
	}

	return root, nil
}
