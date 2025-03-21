package config

import (
	"sort"
	"strings"
)

type Config interface {
	Categories() []string
	Value(cat, name string) string
}

var _ Config = (*specImpl)(nil)

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

// Get returns the string value associated with the 'name'
// in category 'cat'.
func (c *specImpl) Value(cat, name string) string {
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
