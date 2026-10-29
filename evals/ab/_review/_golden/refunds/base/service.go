package refunds

import (
	"context"
	"fmt"
)

// Gateway issues refunds with the card processor.
type Gateway interface {
	Refund(ctx context.Context, paymentID string, amount int64) error
}

// Service issues refunds against stored payments.
type Service struct {
	store   *Store
	gateway Gateway
}

// NewService returns a Service that records refunds in store and issues them
// through gateway.
func NewService(store *Store, gateway Gateway) *Service {
	return &Service{store: store, gateway: gateway}
}

// Refund refunds amount of the payment id: with the card processor first,
// then in the record.
func (s *Service) Refund(ctx context.Context, id string, amount int64) error {
	p, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}
	if err := s.gateway.Refund(ctx, p.ID, amount); err != nil {
		return fmt.Errorf("refund %s: %w", id, err)
	}
	p.Refunded += amount
	return s.store.Put(ctx, p)
}

// RefundAll refunds what is left of the payment id.
func (s *Service) RefundAll(ctx context.Context, id string) error {
	p, err := s.store.Get(ctx, id)
	if err != nil {
		return err
	}
	return s.Refund(ctx, id, p.Amount-p.Refunded)
}
