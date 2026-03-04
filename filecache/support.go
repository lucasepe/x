package filecache

import (
	"os"
	"path/filepath"
)

func NewCacheDir(name string) (string, error) {
	dir, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}

	cacheDir := filepath.Join(dir, name)
	err = os.MkdirAll(cacheDir, 0o755)

	return cacheDir, err
}
