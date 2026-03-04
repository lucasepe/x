package section_test

import (
	"fmt"
	"strings"

	"github.com/lucasepe/x/section"
)

func ExampleParse() {
	src := `
title = demo

[server]
host = 127.0.0.1
port = 8080

[feature]
enabled = true
`

	cfg, err := section.Parse(strings.NewReader(src))
	if err != nil {
		fmt.Println("error:", err)
		return
	}

	fmt.Println("units:", cfg.Units())
	fmt.Println("has root:", cfg.Has(""))
	fmt.Println("root:", cfg.Content(""))
	fmt.Println("server:", cfg.Content("server"))
	fmt.Println("feature:", cfg.Content("feature"))

	// Output:
	// units: [server feature]
	// has root: true
	// root: [title = demo]
	// server: [host = 127.0.0.1 port = 8080]
	// feature: [enabled = true]
}
