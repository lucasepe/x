// Package env provides convenience functions for retrieving data from environment variables
package env

import (
	"fmt"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Str does the same as os.Getenv, but allows the user to provide a default value (optional).
// Only the first optional argument is used, the rest is discarded.
func Str(name string, defaultValue ...string) string {
	// Retrieve the environment variable as a (possibly empty) string
	value := os.Getenv(name)

	// If empty and a default value was provided, return that
	if value == "" && len(defaultValue) > 0 {
		return defaultValue[0]
	}

	// If not, return the value of the environment variable
	return value
}

// Strs returns a slice of strings by splitting the value associated with
// the 'name' using the specified separator 'sep'. If the value is not found,
// the provided default value is used.
func Strs(name, sep string, defaultValue ...string) []string {
	value := Str(name, defaultValue...)
	return strings.Split(value, sep)
}

// True returns the bool value of the given environment variable name.
// Returns false if it is not declared or empty.
func True(name string) bool {
	val := strings.ToUpper(strings.TrimSpace(Str(name)))
	switch val {
	case "1",
		"ENABLE", "ENABLED",
		"POSITIVE",
		"T", "TRUE",
		"Y", "YES":
		return true
	}
	return false
}

// Equal returns true if the given environment variable is the given string value.
// The whitespace of both values are trimmed before the comparison.
func Equal(name, value string) bool {
	return strings.TrimSpace(Str(name)) == strings.TrimSpace(value)
}

// Int returns the number stored in the environment variable, or the provided default value.
func Int(name string, defaultValue int) int {
	i, err := strconv.Atoi(Str(name))
	if err != nil {
		return defaultValue
	}
	return i
}

// Int64 returns the number stored in the environment variable, or the provided default value.
func Int64(name string, defaultValue int64) int64 {
	i64, err := strconv.ParseInt(Str(name), 10, 64)
	if err != nil {
		return defaultValue
	}
	return i64
}

// Int32 returns the number stored in the environment variable, or the provided default value.
func Int32(name string, defaultValue int32) int32 {
	i32, err := strconv.ParseInt(Str(name), 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(i32)
}

// Int16 returns the number stored in the environment variable, or the provided default value.
func Int16(name string, defaultValue int16) int16 {
	i16, err := strconv.ParseInt(Str(name), 10, 16)
	if err != nil {
		return defaultValue
	}
	return int16(i16)
}

// Int8 returns the number stored in the environment variable, or the provided default value.
func Int8(name string, defaultValue int8) int8 {
	i8, err := strconv.ParseInt(Str(name), 10, 8)
	if err != nil {
		return defaultValue
	}
	return int8(i8)
}

// UInt64 returns the number stored in the environment variable, or the provided default value.
func UInt64(name string, defaultValue uint64) uint64 {
	ui64, err := strconv.ParseUint(Str(name), 10, 64)
	if err != nil {
		return defaultValue
	}
	return ui64
}

// UInt32 returns the number stored in the environment variable, or the provided default value.
func UInt32(name string, defaultValue uint32) uint32 {
	ui32, err := strconv.ParseUint(Str(name), 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint32(ui32)
}

// UInt16 returns the number stored in the environment variable, or the provided default value.
func UInt16(name string, defaultValue uint16) uint16 {
	ui16, err := strconv.ParseUint(Str(name), 10, 16)
	if err != nil {
		return defaultValue
	}
	return uint16(ui16)
}

// UInt8 returns the number stored in the environment variable, or the provided default value.
func UInt8(name string, defaultValue uint8) uint8 {
	ui8, err := strconv.ParseUint(Str(name), 10, 8)
	if err != nil {
		return defaultValue
	}
	return uint8(ui8)
}

// Float64 returns the number stored in the environment variable, or the provided default value.
func Float64(name string, defaultValue float64) float64 {
	f64, err := strconv.ParseFloat(Str(name), 64)
	if err != nil {
		return defaultValue
	}
	return f64
}

// Float32 returns the number stored in the environment variable, or the provided default value.
func Float32(name string, defaultValue float32) float32 {
	f32, err := strconv.ParseFloat(Str(name), 32)
	if err != nil {
		return defaultValue
	}
	return float32(f32)
}

func Duration(key string, defaultValue time.Duration) time.Duration {
	val, ok := os.LookupEnv(key)
	if !ok {
		return defaultValue
	}

	res, err := time.ParseDuration(strings.TrimSpace(val))
	if err != nil {
		return defaultValue
	}
	return res
}

// Contains checks if the given environment variable contains the given string
func Contains(name string, value string) bool {
	return strings.Contains(Str(name), value)
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
	userName := Str("USER")
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
