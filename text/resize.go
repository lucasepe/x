package text

import (
	"bytes"

	"github.com/mattn/go-runewidth"
)

// Resize resizes the string with the given length. It ellipses with '...' when the string's length exceeds
// the desired length or pads spaces to the right of the string when length is smaller than desired
func Resize(s string, length uint, rightAlign bool) string {
	slen := runewidth.StringWidth(s)
	n := int(length)
	if slen == n {
		return s
	}
	// Pads only when length of the string smaller than len needed
	if rightAlign {
		s = PadLeft(s, n, ' ')
	} else {
		s = PadRight(s, n, ' ')
	}
	if slen > n {
		var buf bytes.Buffer
		w := 0
		for _, r := range s {
			buf.WriteRune(r)
			rw := runewidth.RuneWidth(r)
			if w+rw >= n-3 {
				break
			}
			w += rw
		}
		buf.WriteString("...")
		s = buf.String()
	}
	return s
}
