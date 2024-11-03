package rand

import (
	"hash/fnv"
	"math/rand"
)

// SimplePRNG returns a deterministic random number source
// based on a given string. The resulting rand.Source produces
// a reproducible sequence of pseudo-random numbers each time
// it's initialized with the same string.
// Useful for test cases and scenarios where repeatable
// random sequences are required.
func SimplePRNG(s string) rand.Source {
	h := fnv.New64a()
	h.Write([]byte(s))

	seed := int64(h.Sum64())

	return rand.NewSource(seed)
}
