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
	// error is a transport failure. Every call is a network round trip.
	Get(sku string) (string, error)
}

// Resolve resolves skus in order and returns the product name for each SKU it
// resolved, keyed by SKU.
//
// A SKU the source does not know is left out of the result rather than being
// fatal. Any other failure abandons the walk and is returned.
//
// The same SKU may appear in skus more than once, and a round trip is the
// expensive part of this function, so a repeated SKU costs one.
//
// A failure is reported to two audiences at once and has to serve both. The
// operator reads it in a log line and needs to see which SKU failed and what
// went wrong underneath. The caller does not read it: the caller inspects it,
// decides whether the failure is worth a retry, and must be able to reach
// every reason the Source can report.
func Resolve(src Source, skus []string) (map[string]string, error) {
	panic("not implemented")
}
