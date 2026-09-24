// Package probe holds one lint scenario per file.
package probe

import (
	"context"
	"database/sql"
	"io"
	"net/http"
)

// CountRows counts the rows of t.
func CountRows(ctx context.Context, db *sql.DB) (int, error) {
	rows, err := db.QueryContext(ctx, "SELECT id FROM t")
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	n := 0
	for rows.Next() {
		n++
	}
	return n, rows.Err()
}

// Touch updates t inside a transaction.
func Touch(ctx context.Context, db *sql.DB) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, "UPDATE t SET n = n + 1"); err != nil {
		return err
	}
	return tx.Commit()
}

// Fetch returns the body served at url.
func Fetch(ctx context.Context, c *http.Client, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
