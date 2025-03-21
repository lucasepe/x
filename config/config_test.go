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
	fav: true
	`

	conf, err := config.Parse(strings.NewReader(data))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(conf.Categories())
	fmt.Println(conf.All("https://github.com/lucasepe/jviz"))
	fmt.Println(conf.Value("", "outdir"))

	// Output:
	// [https://github.com/lucasepe/file2go https://github.com/lucasepe/jviz]
	// [comment:Easily visualize and share JSON and YAML structures. fav:true]
	// /Users/lucasepe/Downloads/repoes-backups
}
