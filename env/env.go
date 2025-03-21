// Package env provides convenience functions for retrieving data from environment variables
package env

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/lucasepe/x/text/conv"
)

// Str does the same as os.Getenv, but allows the user to provide a default value (optional).
// Only the first optional argument is used, the rest is discarded.
func Str(name, defaultValue string) string {
	// Retrieve the environment variable as a (possibly empty) string
	value := os.Getenv(name)

	// If empty and a default value was provided, return that
	if value == "" {
		return defaultValue
	}

	// If not, return the value of the environment variable
	return value
}

// Strs returns a slice of strings by splitting the value associated with
// the 'name' using the specified separator 'sep'. If the value is not found,
// the provided default value is used.
func Strs(name, sep string, defaults ...string) []string {
	value := os.Getenv(name)
	return conv.Strs(value, sep, defaults...)
}

// True returns the bool value of the given environment variable name.
// Returns false if it is not declared or empty.
func True(name string) bool {
	value := os.Getenv(name)
	return conv.Bool(value, false)
}

// Equal returns true if the given environment variable is the given string value.
// The whitespace of both values are trimmed before the comparison.
func Equal(name, value string) bool {
	got := strings.TrimSpace(os.Getenv(name))
	return got == strings.TrimSpace(value)
}

// Int returns the number stored in the environment variable, or the provided default value.
func Int(name string, defaultValue int) int {
	return conv.Int(os.Getenv(name), defaultValue)
}

// Int64 returns the number stored in the environment variable, or the provided default value.
func Int64(name string, defaultValue int64) int64 {
	return conv.Int64(os.Getenv(name), defaultValue)
}

// Int32 returns the number stored in the environment variable, or the provided default value.
func Int32(name string, defaultValue int32) int32 {
	return conv.Int32(os.Getenv(name), defaultValue)
}

// UInt64 returns the number stored in the environment variable, or the provided default value.
func UInt64(name string, defaultValue uint64) uint64 {
	return conv.UInt64(os.Getenv(name), defaultValue)
}

// UInt32 returns the number stored in the environment variable, or the provided default value.
func UInt32(name string, defaultValue uint32) uint32 {
	return conv.UInt32(os.Getenv(name), defaultValue)
}

// Float64 returns the number stored in the environment variable, or the provided default value.
func Float64(name string, defaultValue float64) float64 {
	return conv.Float64(os.Getenv(name), defaultValue)
}

// Float32 returns the number stored in the environment variable, or the provided default value.
func Float32(name string, defaultValue float32) float32 {
	return conv.Float32(os.Getenv(name), defaultValue)
}

func Duration(key string, defaultValue time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	return conv.Duration(val, defaultValue)
}

// Contains checks if the given environment variable contains the given string
func Contains(name string, value string) bool {
	return strings.Contains(os.Getenv(name), value)
}

// HomeDir returns the path to the home directory of the user, if available.
// If not available, the username is to construct a path starting with /home/.
// If a username is not available, then "/tmp" is returned.
// The returned string is what the home directory should have been named, if it would have existed.
// No checks are made for if the directory exists.
func HomeDir() string {
	if homeDir, err := userHomeDir(); err == nil { // success, use the home directory
		return homeDir
	}
	userName := os.Getenv("USER")
	switch userName {
	case "root":
		// If the user name is "root", use /root
		return "/root"
	case "":
		// If the user name is not available, use either $HOME or /tmp
		return Str("HOME", "/tmp")
	default:
		// Use $HOME if it's available, and a constructed home directory path if not
		return Str("HOME", "/home/"+userName)
	}
}

// ExpandUser replaces a leading ~ or $HOME with the path
// to the home directory of the current user
func ExpandUser(path string) string {
	// this is a simpler alternative to using os.UserHomeDir (which requires Go 1.12 or later)
	if strings.HasPrefix(path, "~") {
		// Expand ~ to the home directory
		path = strings.Replace(path, "~", HomeDir(), 1)
	} else if strings.HasPrefix(path, "$HOME") {
		// Expand a leading $HOME variable to the home directory
		path = strings.Replace(path, "$HOME", HomeDir(), 1)
	}
	return path
}

// Keys returns the all the environment variable names as a sorted string slice
func Keys() []string {
	var keys []string
	for _, keyAndValue := range os.Environ() {
		pair := strings.SplitN(keyAndValue, "=", 2)
		keys = append(keys, pair[0])
	}
	sort.Strings(keys)
	return keys
}

// Map returns the current environment variables as a map from name to value
func Map() map[string]string {
	m := make(map[string]string)
	for _, keyAndValue := range os.Environ() {
		pair := strings.SplitN(keyAndValue, "=", 2)
		m[pair[0]] = pair[1]
	}
	return m
}

// Unset will clear an environment variable by calling os.Setenv(name, "").
// The cache entry will also be cleared if useCaching is true.
func Unset(name string) error {
	return os.Setenv(name, "")
}

func Set(name string, value any) {
	switch v := value.(type) {
	case string:
		os.Setenv(name, v)
	case []byte:
		os.Setenv(name, string(v))
	case fmt.Stringer:
		os.Setenv(name, v.String())
	default:
		os.Setenv(name, fmt.Sprintf("%v", v))
	}
}
