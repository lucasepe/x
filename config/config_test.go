package config_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/lucasepe/x/config"
)

func Example_config_Parse() {
	data := `
	outdir: /Users/lucasepe/Downloads/repoes-backups
	
	[https://github.com/lucasepe/file2go]
	comment: Convert any file to Go source.
	
	[https://github.com/lucasepe/jviz]
	comment: Easily visualize and share JSON and YAML structures.
	`

	conf, err := config.Parse(strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("download files to: %s\n", conf.Value("", "outdir"))

	for _, name := range conf.Categories() {
		fmt.Println(name)
		fmt.Printf(" > %s\n", conf.Value(name, "comment"))
	}

	// Output:
	// download files to: /Users/lucasepe/Downloads/repoes-backups
	// https://github.com/lucasepe/file2go
	//  > Convert any file to Go source.
	// https://github.com/lucasepe/jviz
	//  > Easily visualize and share JSON and YAML structures.
}
