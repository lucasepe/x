package rand_test

import (
	"fmt"

	xrand "github.com/lucasepe/x/rand"
)

func Example_rand_Str() {
	src := xrand.SimplePRNG("awesome 'x' package")

	str := xrand.Str(src, 14, xrand.All)
	fmt.Println(str)

	// Output:
	// yFc&g+m5oA^5]-
}
