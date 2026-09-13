// Package order is the order module's root contract: the small set of types
// and errors another module may depend on. Implementation lives in the
// module's layer packages, which this package never imports.
package order

import "errors"

// ErrNotFound reports an order id the module does not know.
var ErrNotFound = errors.New("order: not found")

// Summary is the cross-module view of an order.
type Summary struct {
	ID         string
	TotalCents int64
}
