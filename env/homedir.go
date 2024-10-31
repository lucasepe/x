//go:build !windows

package env

import (
	"fmt"
	"os"
)

// userHomeDir is the same as os.UserHomeDir, except that "getenv" is called instead of "os.Getenv"
// (for caching). Also, the two switch statements are combined into just one.
func userHomeDir() (string, error) {
	value := os.Getenv("HOME")
	if value == "" {
		return "", fmt.Errorf("$HOME is not defined")
	}
	return value, nil
}
