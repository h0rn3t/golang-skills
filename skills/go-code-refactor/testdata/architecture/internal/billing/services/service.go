// Package services holds billing's use cases. It consumes an order summary
// through its own interface and its own named type; internal/app adapts the
// order module's root contract to it.
package services

import (
	"context"
	"errors"
	"fmt"

	"example.com/shop/internal/billing/models"
)

// ErrOrderUnknown is billing's view of an order id the order module rejects.
var ErrOrderUnknown = errors.New("billing: unknown order")

// OrderSummary is the minimal order data billing relies on.
type OrderSummary struct {
	OrderID    string
	TotalCents int64
}

// Orders is the cross-module dependency, declared where it is consumed.
type Orders interface {
	Summary(ctx context.Context, id string) (OrderSummary, error)
}

// Service is billing's use cases.
type Service struct{ orders Orders }

// New wires the use cases over an order reader.
func New(orders Orders) *Service { return &Service{orders: orders} }

// Invoice prices an invoice for an order.
func (s *Service) Invoice(ctx context.Context, orderID string) (models.Invoice, error) {
	sum, err := s.orders.Summary(ctx, orderID)
	if err != nil {
		return models.Invoice{}, fmt.Errorf("invoice %q: %w", orderID, err)
	}
	return models.Invoice{OrderID: sum.OrderID, AmountCents: sum.TotalCents}, nil
}
