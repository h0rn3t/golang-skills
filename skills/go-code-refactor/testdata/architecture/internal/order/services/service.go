// Package services holds the order module's use cases. It declares the
// storage behavior it consumes; internal/order/repositories satisfies it
// structurally and this package never imports it.
package services

import (
	"context"
	"fmt"
	"time"

	"example.com/shop/internal/order"
	"example.com/shop/internal/order/models"
)

// Store is the persistence the use cases need. CreateWithAudit is one atomic
// action by contract: either both rows are stored or neither is.
type Store interface {
	CreateWithAudit(ctx context.Context, o models.Order, audit string) error
	Find(ctx context.Context, id string) (models.Order, error)
}

// Clock is the time source, so tests can pin it.
type Clock interface{ Now() time.Time }

// Service is the order module's use cases.
type Service struct {
	store Store
	clock Clock
}

// New wires the use cases over a store and a clock.
func New(store Store, clock Clock) *Service { return &Service{store: store, clock: clock} }

// Place validates and stores an order with its audit line.
func (s *Service) Place(ctx context.Context, actor string, o models.Order) error {
	if actor == "" {
		return fmt.Errorf("place order %q: %w", o.ID, ErrForbidden)
	}
	if err := o.Validate(); err != nil {
		return fmt.Errorf("place order %q: %w", o.ID, err)
	}
	audit := fmt.Sprintf("%s placed by %s", s.clock.Now().UTC().Format(time.RFC3339), actor)
	if err := s.store.CreateWithAudit(ctx, o, audit); err != nil {
		return fmt.Errorf("place order %q: %w", o.ID, err)
	}
	return nil
}

// Summary returns the cross-module view of an order.
func (s *Service) Summary(ctx context.Context, id string) (order.Summary, error) {
	o, err := s.store.Find(ctx, id)
	if err != nil {
		return order.Summary{}, err
	}
	return order.Summary{ID: o.ID, TotalCents: o.Total()}, nil
}
