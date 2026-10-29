// Package refunds records card payments and the refunds issued against them.
package refunds

import (
	"context"
	"errors"
	"sync"
)

// ErrNotFound is returned for a payment ID the store does not hold.
var ErrNotFound = errors.New("payment not found")

// Payment is one captured card payment. Amounts are in minor units.
type Payment struct {
	ID       string
	Amount   int64
	Refunded int64
	Currency string
}

// Store keeps payments in memory. It is safe for concurrent use.
type Store struct {
	mu       sync.Mutex
	payments map[string]Payment
}

// NewStore returns an empty store.
func NewStore() *Store {
	return &Store{payments: make(map[string]Payment)}
}

// Get returns the payment with the given ID.
func (s *Store) Get(_ context.Context, id string) (Payment, error) {
	p, ok := s.payments[id]
	if !ok {
		return Payment{}, ErrNotFound
	}
	return p, nil
}

// Put stores p, replacing any payment with the same ID.
func (s *Store) Put(_ context.Context, p Payment) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.payments[p.ID] = p
	return nil
}
