package conv

import (
	"strconv"
	"strings"
	"time"
)

// Strs returns a slice of strings by splitting the
// value using the specified separator 'sep'.
// If the value is empty, the provided default values are used.
func Strs(value, sep string, defaultValues ...string) []string {
	if value == "" {
		if len(defaultValues) > 0 {
			return defaultValues
		}

		return []string{}
	}

	return strings.Split(value, sep)
}

// Bool returns the value as bool or the provided fallback value.
func Bool(value string, fallback bool) bool {
	value = strings.ToUpper(value)
	switch value {
	case "1",
		"T", "TRUE",
		"Y", "YES":
		return true
	}
	return fallback
}

// Int returns the value as int or the provided default value.
func Int(value string, defaultValue int) int {
	if value == "" {
		return defaultValue
	}

	i, err := strconv.Atoi(value)
	if err != nil {
		return defaultValue
	}
	return i
}

// Int64 returns the value as int64 or the provided default value.
func Int64(value string, defaultValue int64) int64 {
	if value == "" {
		return defaultValue
	}

	i64, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return i64
}

// Int32 returns the value as int32 or the provided default value.
func Int32(value string, defaultValue int32) int32 {
	i32, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return defaultValue
	}
	return int32(i32)
}

// UInt64 returns the value as uint64 or the provided default value.
func UInt64(value string, defaultValue uint64) uint64 {
	ui64, err := strconv.ParseUint(value, 10, 64)
	if err != nil {
		return defaultValue
	}
	return ui64
}

// UInt32 returns the value as uint32 or the provided default value.
func UInt32(value string, defaultValue uint32) uint32 {
	ui32, err := strconv.ParseUint(value, 10, 32)
	if err != nil {
		return defaultValue
	}
	return uint32(ui32)
}

// Float64 returns the value as float64 or the provided default value.
func Float64(value string, defaultValue float64) float64 {
	f64, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return defaultValue
	}
	return f64
}

// Float32 returns the value as float64 or the provided default value.
func Float32(value string, defaultValue float32) float32 {
	f32, err := strconv.ParseFloat(value, 32)
	if err != nil {
		return defaultValue
	}
	return float32(f32)
}

func Duration(value string, defaultValue time.Duration) time.Duration {
	res, err := time.ParseDuration(value)
	if err != nil {
		return defaultValue
	}
	return res
}

func RGBA(value string) (r, g, b, a uint8) {
	value = strings.TrimPrefix(value, "#")

	a = 255

	switch len(value) {
	case 3: // Short format (RGB)
		r = parseHex(string(value[0]), 0)
		g = parseHex(string(value[1]), 0)
		b = parseHex(string(value[2]), 0)
		// expand from 4-bit to 8-bit
		r |= r << 4
		g |= g << 4
		b |= b << 4
	case 6: // Long format no alpha (RRGGBB)
		r = parseHex(string(value[0:2]), 0)
		g = parseHex(string(value[2:4]), 0)
		b = parseHex(string(value[4:6]), 0)
	case 8: // Long format with alpha (RRGGBBAA)
		r = parseHex(string(value[0:2]), 0)
		g = parseHex(string(value[2:4]), 0)
		b = parseHex(string(value[4:6]), 0)
		a = parseHex(string(value[6:8]), 255)
	}

	return
}

func parseHex(value string, fallback uint8) uint8 {
	value = strings.TrimPrefix(value, "0x")

	size := len(value) * 4 // len(s)*4 since each hex use 4 bit
	got, err := strconv.ParseUint(value, 16, size)
	if err != nil {
		return fallback
	}

	return uint8(got)
}
