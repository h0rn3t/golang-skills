// Package statement renders account statements for download.
package statement

import (
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"
)

// Entry is one posting on an account. Amount is in minor units (cents).
type Entry struct {
	ID     int64
	Booked time.Time
	Memo   string
	Amount int64
}

// Write streams the statement to w as CSV: the header line
// "id,booked,memo,amount", one line per entry (booked as YYYY-MM-DD, memo in
// double quotes with each inner quote doubled, amount as units.cents with a
// leading minus when negative), and a closing "balance,,," line with the sum
// of the amounts. Each line is written to w as soon as it is formatted, so a
// client downloading the statement receives entries while later ones are
// still being produced. Write stops at the first write error: an error on an
// entry's line is returned wrapped as "entry <ID>: <err>", one on the
// balance line as "balance: <err>", and one on the header unwrapped.
func Write(w io.Writer, entries []Entry) error {
	if _, err := fmt.Fprintf(w, "id,booked,memo,amount\n"); err != nil {
		return err
	}
	var total int64
	for _, e := range entries {
		memo := strings.ReplaceAll(e.Memo, `"`, `""`)
		if _, err := fmt.Fprintf(w, "%d,%s,\"%s\",%s\n", e.ID, e.Booked.Format(time.DateOnly), memo, formatCents(e.Amount)); err != nil {
			return fmt.Errorf("entry %d: %w", e.ID, err)
		}
		total += e.Amount
	}
	if _, err := fmt.Fprintf(w, "balance,,,%s\n", formatCents(total)); err != nil {
		return fmt.Errorf("balance: %w", err)
	}
	return nil
}

func formatCents(c int64) string {
	sign := ""
	if c < 0 {
		sign = "-"
		c = -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

// ErrBadAmount reports an amount ParseAmount cannot read.
var ErrBadAmount = errors.New("bad amount")

// ParseAmount reads an amount written as units.cents, with an optional
// leading minus, into cents. The importer calls it once per uploaded file
// for the opening balance.
func ParseAmount(s string) (int64, error) {
	s = strings.TrimSpace(s)
	negative := strings.HasPrefix(s, "-")
	s = strings.TrimPrefix(s, "-")
	parts := strings.Split(s, ".")
	if len(parts) != 2 || len(parts[1]) != 2 {
		return 0, ErrBadAmount
	}
	units, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return 0, ErrBadAmount
	}
	cents, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil || cents < 0 {
		return 0, ErrBadAmount
	}
	amount := units*100 + cents
	if negative {
		amount = -amount
	}
	return amount, nil
}
