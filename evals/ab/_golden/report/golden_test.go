package report

import (
	"fmt"
	"strings"
	"testing"
)

// TestGoldenRender pins the exact rendered bytes for both formats, the
// totals line, and the two error texts. Column widths are part of the
// observable output, so the expectations spell the format strings out.
func TestGoldenRender(t *testing.T) {
	rows := []Row{
		{Account: "acme", Units: 3, Cents: 12345},
		{Account: "globex", Units: 0, Cents: 7},
	}

	var wantText strings.Builder
	wantText.WriteString(fmt.Sprintf("%-20s %8s %12s\n", "ACCOUNT", "UNITS", "AMOUNT"))
	wantText.WriteString(fmt.Sprintf("%-20s %8d %9d.%02d\n", "acme", 3, 123, 45))
	wantText.WriteString(fmt.Sprintf("%-20s %8d %9d.%02d\n", "globex", 0, 0, 7))
	wantText.WriteString(fmt.Sprintf("%-20s %8s %9d.%02d\n", "TOTAL", "", 123, 52))

	got, err := Render(rows, false, true)
	if err != nil {
		t.Fatalf("text render: %v", err)
	}
	if got != wantText.String() {
		t.Fatalf("text render =\n%q\nwant\n%q", got, wantText.String())
	}

	wantCSV := "account,units,amount\n" +
		fmt.Sprintf("%s,%d,%d.%02d\n", "acme", 3, 123, 45) +
		fmt.Sprintf("%s,%d,%d.%02d\n", "globex", 0, 0, 7) +
		fmt.Sprintf("TOTAL,,%d.%02d\n", 123, 52)

	got, err = Render(rows, true, true)
	if err != nil {
		t.Fatalf("csv render: %v", err)
	}
	if got != wantCSV {
		t.Fatalf("csv render =\n%q\nwant\n%q", got, wantCSV)
	}

	got, err = Render(rows, true, false)
	if err != nil {
		t.Fatalf("csv render without total: %v", err)
	}
	if strings.Contains(got, "TOTAL") {
		t.Fatalf("includeTotal=false must omit the total line, got %q", got)
	}

	got, err = Render(nil, false, false)
	if err != nil {
		t.Fatalf("empty render: %v", err)
	}
	if got != fmt.Sprintf("%-20s %8s %12s\n", "ACCOUNT", "UNITS", "AMOUNT") {
		t.Fatalf("empty render = %q, want headers only", got)
	}

	if _, err := Render([]Row{{Account: "", Units: 1}}, false, false); err == nil || err.Error() != "render: row with empty account" {
		t.Fatalf("empty account error = %v, want %q", err, "render: row with empty account")
	}

	if _, err := Render([]Row{{Account: "acme", Units: -1}}, false, false); err == nil || err.Error() != "render acme: negative units -1" {
		t.Fatalf("negative units error = %v, want %q", err, "render acme: negative units -1")
	}
}
