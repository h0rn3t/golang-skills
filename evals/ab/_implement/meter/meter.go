// Package meter limits how many requests each client address may make per
// window.
package meter

import "net/netip"

// Meter counts the requests of each client address in the current window and
// refuses a client once it reaches its limit. A Meter is safe for concurrent
// use.
//
// The service runs one Meter per process, and a ticker in main calls Reset
// once a minute. Most minutes see a few thousand addresses; a scan can bring
// half a million in one minute. The pods run under a tight memory limit.
type Meter struct {
	// The unexported fields are the implementation's own.
}

// New returns a Meter that allows each client limit requests per window.
// limit is at least 1. New starts no goroutine.
func New(limit int) *Meter {
	panic("not implemented")
}

// Allow records a request from addr and reports whether it is within the
// client's limit for the current window. A refused request still counts.
func (m *Meter) Allow(addr netip.Addr) bool {
	panic("not implemented")
}

// Count returns how many requests addr has made in the current window.
func (m *Meter) Count(addr netip.Addr) int {
	panic("not implemented")
}

// Reset starts a new window: every client's count is back to zero.
func (m *Meter) Reset() {
	panic("not implemented")
}
