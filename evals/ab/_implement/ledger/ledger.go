// Package ledger reports on posted accounting entries.
package ledger

// Entry is one posted amount, in minor units.
type Entry struct {
	Account string
	Cents   int64
}

// Ledger is a snapshot over the entries it was built from.
//
// A Ledger is immutable: once New has returned, nothing any caller does can
// change what that Ledger reports.
type Ledger struct {
	// The unexported fields are the implementation's own.
}

// New returns a Ledger over entries.
func New(entries []Entry) *Ledger {
	panic("not implemented")
}

// Report renders the per-account balance report.
//
// One line per account that appears in the entries, sorted by account name,
// carrying the sum of that account's amounts, followed by a TOTAL line
// carrying the sum of every amount. An empty Ledger reports the TOTAL line
// alone.
//
// format selects the rendering and is either "text" or "csv". Any other value
// is an error whose message names the format that was asked for.
//
// "text" is one account per line as "%-20s %10d\n", TOTAL included:
//
//	cash                       2675
//	fees                       -400
//	TOTAL                      2275
//
// "csv" is the same rows under a header line, as "%s,%d\n":
//
//	account,cents
//	cash,2675
//	fees,-400
//	TOTAL,2275
//
// The finance team reads the text form in a terminal and loads the csv form
// into a spreadsheet. A third rendering has never been asked for.
func (l *Ledger) Report(format string) (string, error) {
	panic("not implemented")
}
