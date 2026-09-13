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
func (s *OrderStore) CreateWithAudit(ctx context.Context, o models.Order, audit string) (err error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback() // the first error is the one to report
		}
	}()
	if _, err = tx.ExecContext(ctx, `INSERT INTO orders (id, customer, total_cents) VALUES ($1, $2, $3)`, o.ID, o.Customer, o.Total()); err != nil {
		return fmt.Errorf("insert order: %w", err)
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO order_audit (order_id, note) VALUES ($1, $2)`, o.ID, audit); err != nil {
		return fmt.Errorf("insert audit: %w", err)
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit: %w", err)
	}
	return nil
}

// Find loads one order; a missing row is order.ErrNotFound, the module's
// contract, not sql.ErrNoRows.
func (s *OrderStore) Find(ctx context.Context, id string) (models.Order, error) {
	var o models.Order
	err := s.db.QueryRowContext(ctx, `SELECT id, customer FROM orders WHERE id = $1`, id).Scan(&o.ID, &o.Customer)
	if errors.Is(err, sql.ErrNoRows) {
		return models.Order{}, order.ErrNotFound
	}
	if err != nil {
		return models.Order{}, fmt.Errorf("find order %q: %w", id, err)
	}
	return o, nil
}
