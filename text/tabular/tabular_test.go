package tabular_test

import (
	"io"
	"testing"

	"github.com/lucasepe/x/text/tabular"
)

func TestFormat(t *testing.T) {
	tab := tabular.New()
	tab.AddRightColumn("id", "ID", 6)
	tab.AddColumn("env", "Environment", 14)
	tab.AddColumn("cls", "Cluster", 10)
	tab.AddColumn("svc", "Service", 25)
	tab.AddColumn("hst", "Database Host", 25)
	tab.AddRightColumn("pct", "%CPU", 5)

	// Test Partial Printing
	want := "%6v %-14v %-10v\n"
	if got := tab.Print(io.Discard, "id", "env", "cls"); got != want {
		t.Errorf("ERROR: tab.Print() failed\n   want: %q\n    got: %q", want, got)
	}

	tWant := tabular.Output{
		Header:    "    ID Environment    Cluster    Service                   Database Host              %CPU",
		Separator: "------ -------------- ---------- ------------------------- ------------------------- -----",
		Format:    "%6v %-14v %-10v %-25v %-25v %5v\n",
	}

	// Test Printing All
	want = tWant.Format
	if got := tab.Print(io.Discard, tabular.All); got != want {
		t.Errorf("ERROR: tab.Print(All) failed\n   want: %q\n    got: %q", want, got)
	}

	// Test Parsing
	if tGot := tab.Build("id", "env", "cls", "svc", "hst", "pct"); tGot != tWant {
		if tGot.Header != tWant.Header {
			t.Errorf("ERROR: tab.Parse() failed\n   want: %q\n    got: %q", tWant.Header, tGot.Header)
		}
		if tGot.Separator != tWant.Separator {
			t.Errorf("ERROR: tab.Parse() failed\n   want: %q\n    got: %q", tWant.Separator, tGot.Separator)
		}
		if tGot.Format != tWant.Format {
			t.Errorf("ERROR: tab.Parse() failed\n   want: %q\n    got: %q", tWant.Format, tGot.Format)
		}
	}
}
