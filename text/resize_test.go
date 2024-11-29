package text_test

import (
	"testing"

	xtext "github.com/lucasepe/x/text"
)

func TestResize(t *testing.T) {
	s := "foo"
	got := xtext.Resize(s, 5, false)
	if len(got) != 5 {
		t.Fatal("want", 5, "got", len(got))
	}
	s = "foobar"
	got = xtext.Resize(s, 5, false)

	if got != "fo..." {
		t.Fatal("want", "fo...", "got", got)
	}
}

func TestAlign(t *testing.T) {
	s := "foo"
	got := xtext.Resize(s, 5, false)
	if got != "foo  " {
		t.Fatal("want", "foo  ", "got", got)
	}
	got = xtext.Resize(s, 5, true)
	if got != "  foo" {
		t.Fatal("want", "  foo", "got", got)
	}
}
