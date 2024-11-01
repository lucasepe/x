package config

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strings"
)

func Parse(in io.Reader) (Spec, error) {
	c := &specImpl{
		categoryMap: make(map[string]*category),
	}

	currentSection := rootSection

	ln := 0
	reader := bufio.NewReader(in)
	for {
		line, _, err := reader.ReadLine()
		if err == io.EOF {
			break
		}

		ln = ln + 1

		line = bytes.TrimSpace(line)
		size := len(line)
		if size == 0 {
			continue
		}

		if line[0] == '#' || line[0] == ';' {
			continue
		}

		// parse section
		if line[0] == '[' {
			if line[size-1] == ']' {
				currentSection = string(line[1 : size-1])
				continue
			}
		}

		// parse item
		err = parseOne(c, currentSection, string(line))
		if err != nil {
			return nil, fmt.Errorf("ln %d: %w", ln, err)
		}
	}

	return c, nil
}

func parseOne(c *specImpl, sectionName string, line string) error {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return fmt.Errorf("invalid line format, should be 'key: value'")
	}

	name, value := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])

	item := &param{name, value}

	secs, ok := c.categoryMap[sectionName]
	if !ok {
		secs = &category{sectionName, make(map[string]*param)}
	}

	if _, ok := secs.params[name]; !ok {
		secs.params = make(map[string]*param)
	}

	secs.params[name] = item
	c.categoryMap[sectionName] = secs

	return nil
}
