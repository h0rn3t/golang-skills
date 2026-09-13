// Package books keeps account balances and the entries that move them, and
// prepares the figures a customer sees on a statement.
package books

import (
	"errors"
	"fmt"
	"slices"
	"strconv"
	"strings"
	"time"
)

// ErrNoAccount reports an account id the store does not hold.
var ErrNoAccount = errors.New("no such account")

// ErrNoEntry reports an entry id the store does not hold.
var ErrNoEntry = errors.New("no such entry")

// ErrInsufficient reports a debit larger than the account's balance.
var ErrInsufficient = errors.New("insufficient funds")

// Account is one customer account. Balance is in the minor unit of Currency.
type Account struct {
	ID       int64  `json:"id"`
	Owner    string `json:"owner"`
	Currency string `json:"currency"`
	Balance  int64  `json:"balance_cents"`
	Frozen   bool   `json:"frozen"`
}

// Equal reports whether a and b describe the same account state. The
// reconciler uses it to decide whether a row changed since the last sync.
func (a Account) Equal(b Account) bool {
	return a.ID == b.ID && a.Owner == b.Owner && a.Currency == b.Currency && a.Balance == b.Balance
}

// Entry is one posting on an account. Memo is the counterparty's free text,
// if any; SettledAt is filled in by the settlement import, zero until then.
type Entry struct {
	ID        int64     `json:"id"`
	AccountID int64     `json:"account_id"`
	Debit     int64     `json:"debit_cents"`
	Credit    int64     `json:"credit_cents"`
	Memo      string    `json:"memo,omitempty"`
	PostedAt  time.Time `json:"posted_at"`
	SettledAt time.Time `json:"settled_at,omitzero"`
}

// Amount returns the signed movement of the entry: credits are positive,
// debits negative.
func (e Entry) Amount() int64 { return e.Credit - e.Debit }

// Settled reports whether the counterparty has confirmed the entry.
func (e Entry) Settled() bool { return !e.SettledAt.IsZero() }

// Statement is what a customer sees for one period: the account as it stood
// at the end and every entry posted in between.
type Statement struct {
	Account Account
	Entries []Entry
	// Totals is the net movement by memo; the renderer fills it in.
	Totals map[string]int64
}

// Clone returns a copy the renderer may sort and annotate without touching
// the statement the store handed out.
func (s Statement) Clone() Statement {
	s.Entries = slices.Clone(s.Entries)
	return s
}

// MaxMemo is the longest memo a counterparty may attach, in characters.
const MaxMemo = 140

// ValidateMemo reports an error when memo does not fit a statement line.
func ValidateMemo(memo string) error {
	if len(memo) > MaxMemo {
		return fmt.Errorf("memo is %d characters, the limit is %d", len(memo), MaxMemo)
	}
	return nil
}

// Fee returns the fee on amount at bps basis points, rounded half up to the
// cent. A refund is a negative amount and carries a negative fee of the same
// magnitude as the fee on the charge it reverses.
func Fee(amount, bps int64) int64 {
	return (amount*bps + 5_000) / 10_000
}

// ParseAccountIDs parses the comma-separated account ids of a statement
// request. An empty list is valid and selects every account of the customer.
func ParseAccountIDs(s string) ([]int64, error) {
	parts := strings.Split(s, ",")
	ids := make([]int64, 0, len(parts))
	for _, p := range parts {
		id, err := strconv.ParseInt(strings.TrimSpace(p), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("account id %q: %w", p, err)
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// Duplicates returns the entries of batch that repeat one already in
// existing: same account, same movement, posted at the same instant. The
// importer calls it with the store's rows before writing a replayed batch.
func Duplicates(existing, batch []Entry) []Entry {
	type key struct {
		account, amount int64
		at              time.Time
	}
	seen := make(map[key]bool, len(existing))
	for _, e := range existing {
		seen[key{e.AccountID, e.Amount(), e.PostedAt}] = true
	}
	var dups []Entry
	for _, e := range batch {
		if seen[key{e.AccountID, e.Amount(), e.PostedAt}] {
			dups = append(dups, e)
		}
	}
	return dups
}

// NextBillingDate returns the day the account is billed after t: the same
// day of the following month, or that month's last day when it is shorter.
func NextBillingDate(t time.Time) time.Time {
	return t.AddDate(0, 1, 0)
}

// StartOfDay returns midnight at the start of t's calendar day in loc, the
// boundary a statement period is cut at.
func StartOfDay(t time.Time, loc *time.Location) time.Time {
	return t.In(loc).Truncate(24 * time.Hour)
}
