package config

import (
	"fmt"
	"sort"
	"strings"
)

type Config interface {
	Categories() []string
	Value(cat, name string) string
	All(cat string) []string
}

var _ Config = (*specImpl)(nil)

const (
	rootSection = "_ROOT_SECTION_"
)

type param struct {
	name  string
	value string
}

func (p *param) String() string {
	return fmt.Sprintf("%s:%s", p.name, p.value)
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

func (c *specImpl) All(cat string) []string {
	if strings.TrimSpace(cat) == "" {
		cat = rootSection
	}

	catMap, ok := c.categoryMap[cat]
	if !ok || (catMap == nil) || (catMap.params == nil) {
		return []string{}
	}

	all := make([]string, 0, len(catMap.params))
	for _, el := range catMap.params {
		all = append(all, el.String())
	}
	return all
}
