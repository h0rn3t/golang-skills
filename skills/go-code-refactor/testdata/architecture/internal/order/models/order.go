// Package models holds the order module's business data and invariants.
package models

import "errors"

// ErrEmptyOrder is returned when an order has no lines.
var ErrEmptyOrder = errors.New("order: no lines")

// Line is one priced item on an order.
type Line struct {
	SKU        string
	Quantity   int
	PriceCents int64
}

// Order is an order as the business sees it.
type Order struct {
	ID       string
	Customer string
	Lines    []Line
}

// Total is the order's price in cents.
func (o Order) Total() int64 {
	var total int64
	for _, l := range o.Lines {
		total += int64(l.Quantity) * l.PriceCents
	}
	return total
}

// Validate applies the invariants an order must satisfy before it is stored.
func (o Order) Validate() error {
	if len(o.Lines) == 0 {
		return ErrEmptyOrder
	}
	return nil
}
