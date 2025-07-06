package cl

import (
	"bytes"
	"flag"
	"strings"
	"testing"
)

func TestPrintFlags(t *testing.T) {
	tests := []struct {
		name     string
		setup    func(fs *flag.FlagSet)
		expected string
	}{
		{
			name: "single flag, single-line usage",
			setup: func(fs *flag.FlagSet) {
				fs.String("name", "default", "The name of the user")
			},
			expected: strings.Join([]string{
				"  -name  The name of the user",
				"          ↳ (default: default)",
				"",
			}, "\n"),
		},
		{
			name: "single flag, multi-line usage",
			setup: func(fs *flag.FlagSet) {
				fs.String("name", "default", "The name of the user\nUse full name")
			},
			expected: strings.Join([]string{
				"  -name  The name of the user",
				"         Use full name",
				"          ↳ (default: default)",
				"",
			}, "\n"),
		},
		{
			name: "multiple flags, aligned output",
			setup: func(fs *flag.FlagSet) {
				fs.String("n", "", "Short flag")
				fs.String("longflag", "", "Longer flag description")
			},
			expected: strings.Join([]string{
				"  -longflag  Longer flag description",
				"",
				"  -n         Short flag",
				"",
				"",
			}, "\n"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fs := flag.NewFlagSet("test", flag.ContinueOnError)
			tt.setup(fs)

			var buf bytes.Buffer
			printFlags(fs, &buf)

			got := buf.String()
			if got != tt.expected {
				t.Errorf("Unexpected output:\nGot:\n%s\nExpected:\n%s", got, tt.expected)
			}
		})
	}
}
