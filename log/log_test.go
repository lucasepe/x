package log_test

import (
	"bytes"
	"fmt"
	"os"

	"github.com/lucasepe/x/log"
)

func ExampleNew() {

	os.Setenv("PRETTY", "true")
	os.Setenv("TIMESTAMP", "false")

	out := bytes.Buffer{}

	l := log.New(&out)
	l.I("User logged in",
		log.String("user", "mario"),
		log.Int("age", 42),
		log.Bool("admin", true),
		log.Float64("score", 99.5),
		log.StringSlice("groups", []string{"dev", "ops"}),
	)

	l.E("Errore di rete", log.String("host", "api.local"), log.Int("retries", 3))

	fmt.Print(out.String())

	// Output:
	// [I] User logged in
	//     ├── user:   mario
	//     ├── age:    42
	//     ├── admin:  true
	//     ├── score:  99.5
	//     └── groups: [dev,ops]
	// [E] Errore di rete
	//     ├── host:    api.local
	//     └── retries: 3
}
