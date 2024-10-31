package env_test

import (
	"fmt"
	"log"
	"strings"

	"github.com/lucasepe/x/env"
)

func ExampleLoad() {
	src := `
VAR_A=Hello World
VAR_B=true
# VAR_C=Ghost Me!
VAR_D=1234
VAR_E=one,two,three,four
`

	err := env.Load(strings.NewReader(src))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(env.Str("VAR_A"))
	fmt.Println(env.True("VAR_B"))
	fmt.Println(env.Int("VAR_C", -1))
	fmt.Println(env.Int("VAR_D", -1))
	fmt.Println(env.Strs("VAR_E", ","))

	// Output:
	// Hello World
	// true
	// -1
	// 1234
	// [one two three four]
}
