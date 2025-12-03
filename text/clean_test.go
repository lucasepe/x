package text_test

import (
	"bytes"
	"testing"

	"github.com/lucasepe/x/text"
)

func TestClean(t *testing.T) {
	tests := []struct {
		name string
		in   []byte
		out  []byte
	}{
		{
			name: "removes BOM",
			in:   []byte{0xEF, 0xBB, 0xBF, 'H', 'i'},
			out:  []byte("Hi"),
		},
		{
			name: "replaces NBSP with space",
			in:   []byte("Hello\xc2\xa0World"),
			out:  []byte("Hello World"),
		},
		{
			name: "removes zero-width space",
			in:   []byte("A\xe2\x80\x8bB"),
			out:  []byte("AB"),
		},
		{
			name: "removes ASCII control chars except \\n",
			in:   []byte("A\x01B\x02\nC\tD\x07"),
			out:  []byte("AB\nC   D"),
		},
		{
			name: "leaves normal UTF-8 untouched",
			in:   []byte("Caffè ☕"),
			out:  []byte("Caffè ☕"),
		},
		{
			name: "empty input produces empty output",
			in:   []byte{},
			out:  []byte{},
		},
		{
			name: "no changes",
			in:   []byte("Hello World"),
			out:  []byte("Hello World"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := text.Clean(tt.in, 3)
			if !bytes.Equal(got, tt.out) {
				t.Fatalf("Clean(%q) = %q, want %q", tt.in, got, tt.out)
			}
		})
	}
}
