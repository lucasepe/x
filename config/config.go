package config

import (
	"sort"
	"strconv"
	"strings"
)

type Spec interface {
	Categories() []string
	Str(cat, name string) string
	Strs(cat, name, sep string) []string
	True(cat, name string) bool
	Int(cat, name string, defaultValue int) int
	Int64(cat, name string, defaultValue int64) int64
	Float64(cat, name string, defaultValue float64) float64
}

var _ Spec = (*specImpl)(nil)

const (
	rootSection = "_ROOT_SECTION_"
)

type param struct {
	name  string
	value string
}

type category struct {
	name   string
	params map[string]*param
}

type specImpl struct {
	categoryMap map[string]*category
}

// Strs returns a slice of strings by splitting the value
// associated with the 'name' in category 'cat' using the
// specified separator 'sep'.
func (c *specImpl) Strs(cat, name, sep string) []string {
	value := c.Str(cat, name)
	return strings.Split(value, sep)
}

// Str returns a string valu associated with the 'name'
// in category 'cat'.
func (c *specImpl) Str(cat, name string) string {
	if strings.TrimSpace(cat) == "" {
		cat = rootSection
	}

	catMap, ok := c.categoryMap[cat]
	if !ok || (catMap == nil) || (catMap.params == nil) {
		return ""
	}

	item, ok := catMap.params[name]
	if !ok {
		return ""
	}

	return item.value
}

// True returns the bool value of the given variable
// name in the specified category.
// Returns false if it is not declared or empty.
func (c *specImpl) True(cat, name string) bool {
	val := strings.ToUpper(strings.TrimSpace(c.Str(cat, name)))
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

// Int returns the number stored in the specified category
// with the specified name, or the provided default value.
func (c *specImpl) Int(cat, name string, defaultValue int) int {
	i, err := strconv.Atoi(c.Str(cat, name))
	if err != nil {
		return defaultValue
	}
	return i
}

// Int64 returns the number stored in the specified category
// with the specified name, or the provided default value.
func (c *specImpl) Int64(cat, name string, defaultValue int64) int64 {
	i64, err := strconv.ParseInt(c.Str(cat, name), 10, 64)
	if err != nil {
		return defaultValue
	}
	return i64
}

// Float64 returns the number stored iin the specified category
// with the specified name, or the provided default value.
func (c *specImpl) Float64(cat, name string, defaultValue float64) float64 {
	f64, err := strconv.ParseFloat(c.Str(cat, name), 64)
	if err != nil {
		return defaultValue
	}
	return f64
}

func (c *specImpl) Categories() []string {
	all := []string{}

	for name := range c.categoryMap {
		if name == rootSection {
			continue
		}

		all = append(all, name)
	}

	sort.Strings(all)

	return all
}
