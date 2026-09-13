package books

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// maxPage is the most entries one statement page carries.
const maxPage = 500

// entriesQuery pages through an account's entries by id.
const entriesQuery = `SELECT id, account_id, debit_cents, credit_cents, COALESCE(memo, ''), posted_at
	FROM entries
	WHERE account_id = $1 AND id >= $2
	ORDER BY id
	LIMIT $3`

// Store reads and writes accounts and entries in the ledger database; the
// tables are in schema.sql.
type Store struct {
	db *sql.DB
}

// NewStore returns a store over db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Account returns the account with id, or an error wrapping ErrNoAccount.
func (s *Store) Account(ctx context.Context, id int64) (Account, error) {
	var a Account
	err := s.db.QueryRowContext(ctx, `SELECT id, owner, currency, balance_cents, frozen FROM accounts WHERE id = $1`, id).
		Scan(&a.ID, &a.Owner, &a.Currency, &a.Balance, &a.Frozen)
	if errors.Is(err, sql.ErrNoRows) {
		return Account{}, fmt.Errorf("account %d: %w", id, ErrNoAccount)
	}
	if err != nil {
		return Account{}, fmt.Errorf("account %d: %w", id, err)
	}
	return a, nil
}

// Entry returns the entry with id, or an error wrapping ErrNoEntry.
func (s *Store) Entry(ctx context.Context, id int64) (Entry, error) {
	var e Entry
	err := s.db.QueryRowContext(ctx, `SELECT id, account_id, debit_cents, credit_cents, memo, posted_at FROM entries WHERE id = $1`, id).
		Scan(&e.ID, &e.AccountID, &e.Debit, &e.Credit, &e.Memo, &e.PostedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return Entry{}, fmt.Errorf("entry %d: %w", id, ErrNoEntry)
	}
	if err != nil {
		return Entry{}, fmt.Errorf("entry %d: %w", id, err)
	}
	return e, nil
}

// Entries returns up to limit entries of account posted after the entry with
// id after, oldest first, and whether more follow. An after of 0 starts at
// the account's first entry.
func (s *Store) Entries(ctx context.Context, account, after int64, limit int) ([]Entry, bool, error) {
	if limit <= 0 || limit > maxPage {
		return nil, false, fmt.Errorf("entries: limit %d is not between 1 and %d", limit, maxPage)
	}
	fetch := limit + 1 // one row more than asked says whether a page follows
	rows, err := s.db.QueryContext(ctx, entriesQuery, account, after, fetch)
	if err != nil {
		return nil, false, fmt.Errorf("entries: %w", err)
	}
	defer func() { _ = rows.Close() }()
	out := make([]Entry, 0, fetch)
	for rows.Next() {
		var e Entry
		if err := rows.Scan(&e.ID, &e.AccountID, &e.Credit, &e.Debit, &e.Memo, &e.PostedAt); err != nil {
			return nil, false, fmt.Errorf("entries: %w", err)
		}
		out = append(out, e)
	}
	if err := rows.Err(); err != nil {
		return nil, false, fmt.Errorf("entries: %w", err)
	}
	more := len(out) > limit
	if more {
		out = out[:limit]
	}
	return out, more, nil
}

// Transfer moves amount cents from account from to account to in one
// transaction and returns the id of the debit entry. It refuses a frozen
// source account and fails with ErrInsufficient when the source balance is
// too small. Accounts of one customer share a currency, so none is checked.
func (s *Store) Transfer(ctx context.Context, from, to, amount int64, memo string) (int64, error) {
	if amount <= 0 {
		return 0, fmt.Errorf("transfer: amount %d is not positive", amount)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("transfer: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // a no-op after Commit
	var balance int64
	var frozen bool
	err = tx.QueryRowContext(ctx, `SELECT balance_cents, frozen FROM accounts WHERE id = $1 FOR UPDATE`, from).Scan(&balance, &frozen)
	if err != nil {
		return 0, fmt.Errorf("transfer from %d: %w", from, noAccount(err))
	}
	if frozen {
		return 0, fmt.Errorf("transfer from %d: account is frozen", from)
	}
	if balance < amount {
		return 0, fmt.Errorf("transfer from %d: %w", from, ErrInsufficient)
	}
	var toBalance int64
	err = tx.QueryRowContext(ctx, `SELECT balance_cents FROM accounts WHERE id = $1`, to).Scan(&toBalance)
	if err != nil {
		return 0, fmt.Errorf("transfer to %d: %w", to, noAccount(err))
	}
	if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = balance_cents - $1 WHERE id = $2`, amount, from); err != nil {
		return 0, fmt.Errorf("transfer: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `UPDATE accounts SET balance_cents = $1 WHERE id = $2`, toBalance+amount, to); err != nil {
		return 0, fmt.Errorf("transfer: %w", err)
	}
	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO entries (account_id, debit_cents, memo) VALUES ($1, $2, NULLIF($3, '')) RETURNING id`, from, amount, memo).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("transfer: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO entries (account_id, credit_cents, memo) VALUES ($1, $2, NULLIF($3, ''))`, to, amount, memo); err != nil {
		return 0, fmt.Errorf("transfer: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("transfer: %w", err)
	}
	return id, nil
}

// noAccount maps a missing row to ErrNoAccount and passes other errors on.
func noAccount(err error) error {
	if errors.Is(err, sql.ErrNoRows) {
		return ErrNoAccount
	}
	return err
}
