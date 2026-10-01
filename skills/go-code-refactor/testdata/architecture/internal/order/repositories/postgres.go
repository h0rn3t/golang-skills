// Package repositories owns the order module's persistence: SQL, row mapping,
// and the transaction that keeps an order and its audit line atomic.
package repositories

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"example.com/shop/internal/order"
	"example.com/shop/internal/order/models"
)

// OrderStore stores orders in PostgreSQL.
type OrderStore struct{ db *sql.DB }

// NewOrderStore wraps a database handle.
func NewOrderStore(db *sql.DB) *OrderStore { return &OrderStore{db: db} }

// CreateWithAudit inserts the order and its audit line in one transaction.
// Both statements run on the transaction, never on the pool; a failed second
// write rolls the first back, and a commit failure is returned, not hidden.
func (s *OrderStore) CreateWithAudit(ctx context.Context, o models.Order, audit string) error {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer tx.Rollback() // no-op after Commit; a panic before Commit still releases the connection
	if _, err := tx.ExecContext(ctx, `INSERT INTO orders (id, customer, total_cents) VALUES ($1, $2, $3)`, o.ID, o.Customer, o.Total()); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO order_audit (order_id, note) VALUES ($1, $2)`, o.ID, audit); err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Summary loads one order's id and the total stored with it; a missing row is
// order.ErrNotFound, the module's contract, not sql.ErrNoRows.
func (s *OrderStore) Summary(ctx context.Context, id string) (order.Summary, error) {
	var sum order.Summary
	err := s.db.QueryRowContext(ctx, `SELECT id, total_cents FROM orders WHERE id = $1`, id).Scan(&sum.ID, &sum.TotalCents)
	if errors.Is(err, sql.ErrNoRows) {
		return order.Summary{}, order.ErrNotFound
	}
	if err != nil {
		return order.Summary{}, fmt.Errorf("find order %q: %w", id, err)
	}
	return sum, nil
}
