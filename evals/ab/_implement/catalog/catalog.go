// Package catalog resolves product SKUs against a backing source.
package catalog

import "errors"

// ErrNotFound is what a Source reports when it holds no product for a SKU.
var ErrNotFound = errors.New("catalog: product not found")

// ErrClosed is what a Source reports once it has been shut down.
var ErrClosed = errors.New("catalog: source closed")

// Source is the backing store a lookup reads from.
type Source interface {
	// Get returns the product name for sku. It reports ErrNotFound when the
	// SKU is unknown and ErrClosed when the store is shut down; any other
	// error is a transport failure.
	Get(sku string) (string, error)
}

// Product is one resolved catalog entry.
type Product struct {
	SKU  string
	Name string
}

// Lookup resolves one SKU.
//
// A failure is reported to two audiences at once and has to serve both. The
// operator reads it in a log line and needs to see which SKU failed and what
// went wrong underneath. The caller does not read it: the caller inspects it,
// decides whether this failure is worth a retry or a 404, and must be able to
// reach every reason the Source can report.
func Lookup(src Source, sku string) (Product, error) {
	panic("not implemented")
}

// LookupAll resolves every SKU in order and returns the products it resolved.
//
// An unknown SKU is skipped rather than fatal. Any other failure stops the
// walk and is reported with the same obligations Lookup has.
func LookupAll(src Source, skus []string) ([]Product, error) {
	panic("not implemented")
}
