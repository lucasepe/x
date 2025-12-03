package text

import "unicode/utf8"

// Clean normalizes raw text input by removing or replacing common invisible or
// problematic Unicode characters that often appear in files, terminals, or
// piped input.
//
// The following transformations are applied:
//
//   - Strip UTF-8 BOM (EF BB BF) at start
//   - Replace NBSP (C2 A0) with ASCII space
//   - Remove zero-width space (E2 80 8B)
//   - Replace tabs with tabSize spaces
//   - Remove all other control chars < 0x20 except '\n'
//
// Clean is useful when handling user-provided text that may include invisible
// characters that cause rendering issues, unexpected spacing, or incorrect
// layout calculations.
//
// The returned slice is newly allocated. The input slice is never modified.
func Clean(in []byte, tabSize int) []byte {
	// Pre-allocate result with a similar capacity
	out := make([]byte, 0, len(in))

	i := 0
	n := len(in)

	// Handle optional BOM only at the start
	if n >= 3 && in[0] == 0xEF && in[1] == 0xBB && in[2] == 0xBF {
		i = 3
	}

	for i < n {
		c := in[i]

		// ASCII range
		if c < 0x80 {
			// Control chars
			if c < 32 {
				if c == '\n' {
					out = append(out, '\n')
				} else if c == '\t' {
					// Insert tabSize spaces
					for s := 0; s < tabSize; s++ {
						out = append(out, ' ')
					}
				}
				// Skip other control chars
				i++
				continue
			}

			// Normal ASCII
			out = append(out, c)
			i++
			continue
		}

		// UTF-8 multi-byte characters
		// Check NBSP = C2 A0
		if i+1 < n && c == 0xC2 && in[i+1] == 0xA0 {
			out = append(out, ' ')
			i += 2
			continue
		}

		// Check Zero-Width Space = E2 80 8B
		if i+2 < n &&
			c == 0xE2 &&
			in[i+1] == 0x80 &&
			in[i+2] == 0x8B {
			i += 3
			continue
		}

		// Default: copy the multi-byte UTF-8 rune
		// (Find UTF-8 rune length)
		_, size := utf8.DecodeRune(in[i:])
		out = append(out, in[i:i+size]...)
		i += size
	}

	return out
}
