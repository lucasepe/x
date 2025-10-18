// Package cl implements a simple way for a single command to have many
// subcommands, each of which takes arguments and so forth.
package cl

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"github.com/lucasepe/x/text"
	"github.com/lucasepe/x/text/table"
)

// A Task represents a single command.
type Task interface {
	// Name returns the name of the command.
	Name() string

	// Synopsis returns a short string (less than one line) describing the command.
	Synopsis() string

	// Usage returns a long string explaining the command and giving usage
	// information.
	Usage() string

	// SetFlags adds the flags for this command to the specified set.
	SetFlags(*flag.FlagSet)

	// Execute executes the command and returns an ExitStatus.
	Execute(ctx context.Context, f *flag.FlagSet, args ...any) ExitStatus

	Ctx() context.Context
}

// A Tool represents a set of tasks.
type Tool struct {
	tasks    []*TaskGroup
	topFlags *flag.FlagSet // top-level flags
	name     string        // normally path.Base(os.Args[0])

	Header func(io.Writer)
	Footer func(io.Writer)

	Explain          func(io.Writer)             // A function to print a top level usage explanation. Can be overridden.
	ExplainTaskGroup func(io.Writer, *TaskGroup) // A function to print a command group's usage explanation. Can be overridden.
	ExplainTask      func(io.Writer, Task)       // A function to print a command usage explanation. Can be overridden.

	Output io.Writer // Output specifies where the commander should write its output (default: os.Stdout).
	Error  io.Writer // Error specifies where the commander should write its error (default: os.Stderr).
}

// A TaskGroup represents a set of tasks about a common topic.
type TaskGroup struct {
	name  string
	tasks []Task
}

// Name returns the group name
func (g *TaskGroup) Name() string {
	return g.name
}

// An ExitStatus represents a Posix exit status that a subcommand
// expects to be returned to the shell.
type ExitStatus int

const (
	ExitSuccess ExitStatus = iota
	ExitFailure
	ExitUsageError
)

// NewTool returns a new tool with the specified top-level
// flags and command name. The Usage function for the topLevelFlags
// will be set as well.
func NewTool(topLevelFlags *flag.FlagSet, name string) *Tool {
	cdr := &Tool{
		topFlags: topLevelFlags,
		name:     name,
		Output:   os.Stdout,
		Error:    os.Stderr,
	}

	cdr.Explain = cdr.explain
	cdr.ExplainTaskGroup = explainGroup
	cdr.ExplainTask = explainTask
	topLevelFlags.Usage = func() { cdr.Explain(cdr.Error) }
	return cdr
}

// Name returns the commander's name
func (cdr *Tool) Name() string {
	return cdr.name
}

// Register adds a subcommand to the supported subcommands in the
// specified group. (Help output is sorted and arranged by group name.)
// The empty string is an acceptable group name; such subcommands are
// explained first before named groups.
func (cdr *Tool) Register(cmd Task, group string) {
	for _, g := range cdr.tasks {
		if g.name == group {
			g.tasks = append(g.tasks, cmd)
			return
		}
	}
	cdr.tasks = append(cdr.tasks, &TaskGroup{
		name:  group,
		tasks: []Task{cmd},
	})
}

// VisitGroups visits each command group in lexicographical order, calling
// fn for each.
func (cdr *Tool) VisitGroups(fn func(*TaskGroup)) {
	sort.Sort(byGroupName(cdr.tasks))
	for _, g := range cdr.tasks {
		fn(g)
	}
}

// VisitAll visits the top level flags in lexicographical order, calling fn
// for each. It visits all flags, even those not set.
func (cdr *Tool) VisitAll(fn func(*flag.Flag)) {
	if cdr.topFlags != nil {
		cdr.topFlags.VisitAll(fn)
	}
}

// Execute should be called once the top-level-flags on a Commander
// have been initialized. It finds the correct subcommand and executes
// it, and returns an ExitStatus with the result. On a usage error, an
// appropriate message is printed to os.Stderr, and ExitUsageError is
// returned. The additional args are provided as-is to the Execute method
// of the selected Command.
func (cdr *Tool) Execute(ctx context.Context, args ...any) ExitStatus {
	if cdr.topFlags.NArg() < 1 {
		cdr.topFlags.Usage()
		return ExitUsageError
	}

	name := cdr.topFlags.Arg(0)

	for _, group := range cdr.tasks {
		for _, cmd := range group.tasks {
			if name != cmd.Name() {
				continue
			}
			f := flag.NewFlagSet(name, flag.ContinueOnError)
			f.Usage = func() { cdr.ExplainTask(cdr.Error, cmd) }
			cmd.SetFlags(f)

			if err := f.Parse(cdr.topFlags.Args()[1:]); err != nil {
				if err == flag.ErrHelp {
					// For top-level flags, `flags.Parse()` will handle
					// `--help` and `-h` flags by printing usage information
					// and exiting with status 0 (success).
					//
					// For consistency, we return ExitSuccess here so that
					// calling a subcommand with `--help` or `-h` will also be
					// treated as success.
					return ExitSuccess
				}

				return ExitUsageError
			}

			return cmd.Execute(ctx, f, args...)
		}
	}

	// Cannot find this command.
	cdr.topFlags.Usage()
	return ExitUsageError
}

// countFlags returns the number of top-level flags defined, even those not set.
func (cdr *Tool) countTopFlags() int {
	count := 0
	cdr.VisitAll(func(*flag.Flag) {
		count++
	})
	return count
}

// Sorting of a slice of command groups.
type byGroupName []*TaskGroup

// TODO Sort by function rather than implement sortable?
func (p byGroupName) Len() int           { return len(p) }
func (p byGroupName) Less(i, j int) bool { return p[i].name < p[j].name }
func (p byGroupName) Swap(i, j int)      { p[i], p[j] = p[j], p[i] }

// explain prints a brief description of all the subcommands and the
// important top-level flags.
func (cdr *Tool) explain(w io.Writer) {
	if cdr.Header != nil {
		cdr.Header(w)
	}

	fmt.Fprintf(w, "USAGE:\n\n")
	if cdr.topFlags.NFlag() > 0 {
		fmt.Fprintf(w, "  %s <COMMAND>\n\n", cdr.name)
	} else {
		fmt.Fprintf(w, "  %s [flags] <COMMAND>\n\n", cdr.name)
	}

	sort.Sort(byGroupName(cdr.tasks))
	for _, group := range cdr.tasks {
		cdr.ExplainTaskGroup(w, group)
	}

	if cdr.Footer != nil {
		cdr.Footer(w)
	}

	if cdr.topFlags == nil {
		fmt.Fprintln(w, "\nNo top level flags.")
		return
	}
}

// Sorting of the commands within a group.
func (g TaskGroup) Len() int           { return len(g.tasks) }
func (g TaskGroup) Less(i, j int) bool { return g.tasks[i].Name() < g.tasks[j].Name() }
func (g TaskGroup) Swap(i, j int)      { g.tasks[i], g.tasks[j] = g.tasks[j], g.tasks[i] }

// explainGroup explains all the subcommands for a particular group.
func explainGroup(w io.Writer, group *TaskGroup) {
	if len(group.tasks) == 0 {
		return
	}
	if group.name == "" {
		fmt.Fprintf(w, "COMMANDS:\n\n")
	} else {
		fmt.Fprintf(w, "COMMANDS for %s:\n\n", group.name)
	}
	sort.Sort(group)

	tbl := table.New()
	tbl.Separator = "     "
	tbl.Wrap = true
	for _, cmd := range group.tasks {
		tbl.AddRow(cmd.Name(), cmd.Synopsis())
	}
	fmt.Fprintln(w, text.Indent(tbl.String(), "  "))
	fmt.Fprintln(w)
}

// explainCmd prints a brief description of a single command.
func explainTask(w io.Writer, cmd Task) {
	fmt.Fprintf(w, "%s", cmd.Usage())
	subflags := flag.NewFlagSet(cmd.Name(), flag.ExitOnError)
	subflags.SetOutput(w)
	cmd.SetFlags(subflags)
	//subflags.PrintDefaults()

	count := 0
	subflags.VisitAll(func(f *flag.Flag) {
		count++
	})
	if count > 0 {
		fmt.Fprint(w, "FLAGS:\n\n")
		printFlags(subflags, w)
	}
}

func printFlags(fs *flag.FlagSet, w io.Writer) {
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
