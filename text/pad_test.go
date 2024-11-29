package text_test

import (
	"testing"

	"github.com/lucasepe/x/text"
)

func TestPadRight(t *testing.T) {
	got := text.PadRight("foo", 5, '-')
	if got != "foo--" {
		t.Fatal("want", "foo--", "got", got)
	}
}

func TestPadLeft(t *testing.T) {
	got := text.PadLeft("foo", 5, '-')
	if got != "--foo" {
		t.Fatal("want", "--foo", "got", got)
	}
}
