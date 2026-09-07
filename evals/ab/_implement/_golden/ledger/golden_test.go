package ledger

import (
	"strings"
	"testing"
)

func goldenEntries() []Entry {
	return []Entry{
		{Account: "cash", Cents: 2500},
		{Account: "fees", Cents: -400},
		{Account: "cash", Cents: 175},
	}
}

const goldenText = "cash                       2675\n" +
	"fees                       -400\n" +
	"TOTAL                      2275\n"

const goldenCSV = "account,cents\n" +
	"cash,2675\n" +
	"fees,-400\n" +
	"TOTAL,2275\n"

// TestNewCopiesReceivedEntries is the trap. A Ledger that keeps the caller's
// slice is not the snapshot the documentation promises, because the caller
// still owns the backing array.
func TestNewCopiesReceivedEntries(t *testing.T) {
	entries := goldenEntries()
	l := New(entries)

	entries[0].Cents = 999999
	entries[2].Account = "rewritten"

	got, err := l.Report("text")
	if err != nil {
		t.Fatalf("Report(text) error = %v, want nil", err)
	}
	if got != goldenText {
		t.Errorf("Report(text) after mutating the caller's slice =\n%q\nwant\n%q", got, goldenText)
	}
}

func TestReportRenderings(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{format: "text", want: goldenText},
		{format: "csv", want: goldenCSV},
	}

	l := New(goldenEntries())
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got, err := l.Report(tt.format)
			if err != nil {
				t.Fatalf("Report(%q) error = %v, want nil", tt.format, err)
			}
			if got != tt.want {
				t.Errorf("Report(%q) =\n%q\nwant\n%q", tt.format, got, tt.want)
			}
		})
	}
}

func TestReportRejectsUnknownFormat(t *testing.T) {
	l := New(goldenEntries())

	_, err := l.Report("xml")
	if err == nil {
		t.Fatal("Report(xml) error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "xml") {
		t.Errorf("Report(xml) error = %q, want the format named in the message", err)
	}
}

func TestReportEmptyLedger(t *testing.T) {
	tests := []struct {
		format string
		want   string
	}{
		{format: "text", want: "TOTAL                         0\n"},
		{format: "csv", want: "account,cents\nTOTAL,0\n"},
	}

	l := New(nil)
	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			got, err := l.Report(tt.format)
			if err != nil {
				t.Fatalf("Report(%q) on an empty ledger error = %v, want nil", tt.format, err)
			}
			if got != tt.want {
				t.Errorf("Report(%q) on an empty ledger =\n%q\nwant\n%q", tt.format, got, tt.want)
			}
		})
	}
}
