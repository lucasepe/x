package conv_test

import (
	"testing"

	"github.com/lucasepe/x/text/conv"
)

func TestRGBA(t *testing.T) {
	tests := []struct {
		input    string
		expected [4]uint8 // r, g, b, a
	}{
		{"#000", [4]uint8{0, 0, 0, 255}},
		{"#FFF", [4]uint8{255, 255, 255, 255}},
		{"#123", [4]uint8{17, 34, 51, 255}}, // 0x11, 0x22, 0x33
		{"#112233", [4]uint8{17, 34, 51, 255}},
		{"#AABBCC", [4]uint8{170, 187, 204, 255}},
		{"#AABBCCDD", [4]uint8{170, 187, 204, 221}},
		{"#abcdef", [4]uint8{171, 205, 239, 255}},
		{"#12345678", [4]uint8{18, 52, 86, 120}}, // Alpha = 0x78 = 120
		{"#GGGGGG", [4]uint8{0, 0, 0, 255}},      // not hex chars
		{"#12", [4]uint8{0, 0, 0, 255}},          // invalid length
		{"", [4]uint8{0, 0, 0, 255}},             // empty string
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			r, g, b, a := conv.RGBA(tt.input)
			got := [4]uint8{r, g, b, a}

			if got != tt.expected {
				t.Fatalf("got %v, expected %v", got, tt.expected)
			}
		})
	}
}
