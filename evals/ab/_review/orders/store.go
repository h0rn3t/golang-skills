package orders

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// ErrNotFound reports an order id the store does not hold.
var ErrNotFound = errors.New("order not found")

// Order is one order as stored.
type Order struct {
	ID       int64   `json:"id"`
	Customer string  `json:"customer"`
	Total    float64 `json:"total"`
}

// Store reads and writes orders in a SQL database.
type Store struct {
	db *sql.DB
}

// NewStore returns a store over db.
func NewStore(db *sql.DB) *Store { return &Store{db: db} }

// Get returns the order with id, or an error wrapping ErrNotFound.
func (s *Store) Get(ctx context.Context, id int64) (Order, error) {
	var o Order
	err := s.db.QueryRowContext(ctx, `SELECT id, customer, total FROM orders WHERE id = $1`, id).
		Scan(&o.ID, &o.Customer, &o.Total)
	if errors.Is(err, sql.ErrNoRows) {
		return Order{}, fmt.Errorf("order %d: %w", id, ErrNotFound)
	}
	if err != nil {
		return Order{}, fmt.Errorf("order %d: %w", id, err)
	}
	return o, nil
}

// Search returns the orders of customer; an empty customer returns every order.
func (s *Store) Search(ctx context.Context, customer string) ([]Order, error) {
	query := `SELECT id, customer, total FROM orders`
	if customer != "" {
		query += fmt.Sprintf(" WHERE customer = '%s'", customer)
	}
	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("search orders: %w", err)
	}
	orders := []Order{}
	for rows.Next() {
		var o Order
		if err := rows.Scan(&o.ID, &o.Customer, &o.Total); err != nil {
			return nil, fmt.Errorf("search orders: %w", err)
		}
		orders = append(orders, o)
	}
	return orders, nil
}

// Create stores in and returns its new id.
func (s *Store) Create(ctx context.Context, in Order) (int64, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	defer func() { _ = tx.Rollback() }() // a no-op after Commit
	var id int64
	err = tx.QueryRowContext(ctx, `INSERT INTO orders (customer, total) VALUES ($1, $2) RETURNING id`, in.Customer, in.Total).
		Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	if _, err := s.db.ExecContext(ctx, `INSERT INTO order_events (order_id, kind) VALUES ($1, 'created')`, id); err != nil {
		return 0, fmt.Errorf("create order: %w", err)
	}
	return id, tx.Commit()
}

// Audit records that the order with id was served to a client.
func (s *Store) Audit(ctx context.Context, id int64) error {
	if _, err := s.db.ExecContext(ctx, `UPDATE orders SET audited = true WHERE id = $1`, id); err != nil {
		return fmt.Errorf("audit order %d: %w", id, err)
	}
	return nil
}
