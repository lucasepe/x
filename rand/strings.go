package rand

// Seed the random number generator in one of many possible ways.
// func Seed() {
// 	random = rand.New(rand.NewSource(time.Now().UTC().UnixNano() + 1337))
// }

import (
	"math/rand"
	"strings"
)

const (
	None  = 0
	Lower = 1 << 0
	Upper = 1 << 1
	Digit = 1 << 2
	Punct = 1 << 3

	LowerUpper      = Lower | Upper
	LowerDigit      = Lower | Digit
	UpperDigit      = Upper | Digit
	LowerUpperDigit = LowerUpper | Digit
	All             = LowerUpperDigit | Punct
)

const (
	lower = "abcdefghijklmnopqrstuvwxyz"
	upper = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	digit = "0123456789"
	punct = "~!@#$%^&*()_+-=[]{}<>"
)

// Str generates a random string of a given length.
func Str(src rand.Source, size int, set int) string {
	sb := strings.Builder{}
	if set&Lower > 0 {
		sb.WriteString(lower)
	}
	if set&Upper > 0 {
		sb.WriteString(upper)
	}
	if set&Digit > 0 {
		sb.WriteString(digit)
	}
	if set&Punct > 0 {
		sb.WriteString(punct)
	}
	all := sb.String()
	tot := len(all)

	rng := rand.New(src)

	buf := make([]byte, size)
	for i := 0; i < size; i++ {
		buf[i] = all[rng.Intn(tot)]
	}

	return string(buf)
}
