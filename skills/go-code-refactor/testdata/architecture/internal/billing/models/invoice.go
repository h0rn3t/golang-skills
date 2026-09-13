// Package models holds the billing module's business data.
package models

// Invoice is what billing issues for an order.
type Invoice struct {
	OrderID     string
	AmountCents int64
}
