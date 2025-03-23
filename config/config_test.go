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
	fmt.Println(conf.Value("", "outdir"))
	all := conf.All("https://github.com/lucasepe/jviz")
	for k, v := range all {
		fmt.Println(k, "->", v)
	}

	all = conf.All("pippo")
	if len(all) == 0 {
		fmt.Println("no values for category 'pippo'")
	}

	// Output:
	// [https://github.com/lucasepe/file2go https://github.com/lucasepe/jviz]
	// /Users/lucasepe/Downloads/repoes-backups
	// comment -> Easily visualize and share JSON and YAML structures.
	// fav -> true
	// no values for category 'pippo'
}
