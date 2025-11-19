package cl

import (
	"flag"
	"fmt"
	"io"
	"strings"
)

func PrintFlags(fs *flag.FlagSet, w io.Writer) {
	maxNameLen := 0
	fs.VisitAll(func(f *flag.Flag) {
		if len(f.Name) > maxNameLen {
			maxNameLen = len(f.Name)
		}
	})

	indentColumn := 2 + 1 + maxNameLen + 2 // 2 spaces + "-" + name + 2 spaces padding

	fs.VisitAll(func(f *flag.Flag) {
		usage := strings.Split(f.Usage, "\n")
		tot := len(usage)
		for i, line := range usage {
			if i == 0 {
				padding := strings.Repeat(" ", maxNameLen-len(f.Name)+2)
				fmt.Fprintf(w, "  -%s%s%s\n", f.Name, padding, line)
			} else {
				fmt.Fprintf(w, "%s%s\n", strings.Repeat(" ", indentColumn), line)
			}
			if f.DefValue != "" && (i == (tot - 1)) {
				fmt.Fprintf(w, "%s ↳ (default: %s)\n",
					strings.Repeat(" ", indentColumn), f.DefValue)
			}
		}
		fmt.Fprintln(w)
	})
}
