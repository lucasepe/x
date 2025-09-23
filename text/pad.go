package text

import (
	"bytes"

	"github.com/mattn/go-runewidth"
)

// PadRight returns a new string of a specified visual width in which
// the end of the current string is padded with spaces or with a specified Unicode character.
func PadRight(str string, length int, pad rune) string {
	w := runewidth.StringWidth(str)
	if w >= length {
		return str
	}
	var buf bytes.Buffer
	buf.WriteString(str)
	for i := 0; i < length-w; i++ {
		buf.WriteRune(pad)
	}
	return buf.String()
}

// PadLeft returns a new string of a specified visual width in which the
// beginning of the current string is padded with spaces or with a specified Unicode character.
func PadLeft(str string, length int, pad rune) string {
	w := runewidth.StringWidth(str)
	if w >= length {
		return str
	}
	var buf bytes.Buffer
	for i := 0; i < length-w; i++ {
		buf.WriteRune(pad)
	}
	buf.WriteString(str)
	return buf.String()
}
