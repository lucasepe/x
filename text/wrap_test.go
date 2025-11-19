package text_test

import (
	"bytes"
	"testing"

	"github.com/lucasepe/x/text"
)

const (
	input = "The quick brown fox jumps over the lazy dog."

	defaultPenalty = 1e5
)

var (
	sp = []byte{' '}
)

func TestWrap(t *testing.T) {
	exp := [][]string{
		{"The", "quick", "brown", "fox"},
		{"jumps", "over", "the", "lazy", "dog."},
	}
	words := bytes.Split([]byte(input), sp)
	got := text.WrapWords(words, 1, 24, defaultPenalty)
	if len(exp) != len(got) {
		t.Fail()
	}
	for i := range exp {
		if len(exp[i]) != len(got[i]) {
			t.Fail()
		}
		for j := range exp[i] {
			if exp[i][j] != string(got[i][j]) {
				t.Fatal(i, exp[i][j], got[i][j])
			}
		}
	}
}

func TestWrapNarrow(t *testing.T) {
	exp := "The\nquick\nbrown\nfox\njumps\nover\nthe\nlazy\ndog."
	if text.Wrap(input, 5) != exp {
		t.Fail()
	}
}

func TestWrapOneLine(t *testing.T) {
	exp := "The quick brown fox jumps over the lazy dog."
	if text.Wrap(input, 500) != exp {
		t.Fail()
	}
}

func TestWrapBug1(t *testing.T) {
	cases := []struct {
		limit int
		text  string
		want  string
	}{
		{5, "aaaaa", "aaaaa"},
		{5, "a aaaaa", "a\naaaaa"},
		{4, "overlong overlong foo", "over\nlong\nover\nlong\nfoo"},
	}

	for _, test := range cases {
		got := text.Wrap(test.text, test.limit)
		if got != test.want {
			t.Errorf("Wrap(%q, %d) = %q want %q", test.text, test.limit, got, test.want)
		}
	}
}
