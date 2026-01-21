package tabular

import (
	"fmt"
	"io"
	"strings"
)

// Output contains the built table header, separator, and format string.
type Output struct {
	Header    string
	Separator string
	Format    string
}

// Table represents a table definition with multiple columns.
type Table struct {
	columns map[string]Column
	order   []string
	Gap     int // number of spaces between columns
}

// Column represents a single column in the table.
type Column struct {
	Title     string
	Width     int
	RightJust bool
}

// All is a special value to indicate "use all columns in order"
const All = "*"

// New creates a new table definition with default gap of 1 space.
func New() *Table {
	return &Table{
		columns: make(map[string]Column),
		Gap:     1,
	}
}

// AddColumn adds a left-aligned column.
func (t *Table) AddColumn(key, title string, width int) {
	t.add(key, Column{
		Title: title,
		Width: width,
	})
}

// AddRightColumn adds a right-aligned column.
func (t *Table) AddRightColumn(key, title string, width int) {
	t.add(key, Column{
		Title:     title,
		Width:     width,
		RightJust: true,
	})
}

func (t *Table) add(key string, col Column) {
	t.columns[key] = col
	t.order = append(t.order, key)
}

// Build creates the table output definition.
// Uses t.Gap spaces between columns.
func (t *Table) Build(cols ...string) Output {
	if len(cols) == 1 && cols[0] == All {
		cols = t.order
	}

	var (
		header    strings.Builder
		separator strings.Builder
		format    strings.Builder
	)

	space := strings.Repeat(" ", t.Gap)

	for i, key := range cols {
		col, ok := t.columns[key]
		if !ok {
			continue
		}

		if i > 0 {
			header.WriteString(space)
			separator.WriteString(space)
			format.WriteString(space)
		}

		f := col.format()

		fmt.Fprintf(&header, f, col.Title)
		fmt.Fprintf(&separator, f, dashLine(col.Width))
		format.WriteString(f)
	}

	format.WriteByte('\n')

	return Output{
		Header:    header.String(),
		Separator: separator.String(),
		Format:    format.String(),
	}
}

// Print writes the header and separator to the writer and returns the format string.
func (t *Table) Print(out io.Writer, cols ...string) string {
	nfo := t.Build(cols...)
	if strings.TrimSpace(nfo.Header) != "" {
		fmt.Fprintln(out, nfo.Header)
	}
	fmt.Fprintln(out, nfo.Separator)

	return nfo.Format
}

// format returns the printf format string for a column
func (c Column) format() string {
	if c.RightJust {
		return fmt.Sprintf("%%%dv", c.Width)
	}
	return fmt.Sprintf("%%-%dv", c.Width)
}

// dashLine returns a string of '-' repeated n times
func dashLine(n int) string {
	return strings.Repeat("-", n)
}
