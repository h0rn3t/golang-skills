// Package fetch is the client for the partner's order API.
package fetch

import (
	"context"
	"errors"
	"net/http"
	"time"
)

// ErrRejected is what the partner reports when it refuses a request for a
// reason of the caller's making: a 4xx other than 429.
var ErrRejected = errors.New("fetch: request rejected")

// ErrUnavailable is what the partner reports when it could not serve a
// request through no fault of the caller: a 5xx, a 429, or no response at
// all.
var ErrUnavailable = errors.New("fetch: partner unavailable")

// Client talks to the partner API at BaseURL through HTTP.
//
// The partner is a small shop with a fragile backend. It falls over for a
// few seconds several times an hour, and its status page asks integrators
// to try again after a short pause and to wait longer each time before
// asking again while it recovers, honoring any Retry-After it sends.
type Client struct {
	HTTP    *http.Client
	BaseURL string
	// Attempts is the most times one call may send its request to the
	// partner, counting the first one. Zero means DefaultAttempts.
	Attempts int
	// Pause is the least wait before the second send; no later wait is
	// shorter than the one before it. Zero means DefaultPause.
	Pause time.Duration
}

// DefaultAttempts and DefaultPause are what a zero Client uses.
const (
	DefaultAttempts = 4
	DefaultPause    = 200 * time.Millisecond
)

// Order is one order as the partner returns it.
type Order struct {
	ID     string `json:"id"`
	SKU    string `json:"sku"`
	Amount int64  `json:"amount"`
}

// GetOrder returns the order with id, from GET BaseURL/orders/{id}.
//
// A 404 is ErrRejected. Any response the partner could not give at the time
// — a 5xx, a 429, or a transport failure — is retried within the Client's
// Attempts and Pause, and is ErrUnavailable when the attempts run out; the
// returned error names the id and carries the status or transport failure.
// A cancelled ctx ends the call between sends with ctx.Err().
func (c *Client) GetOrder(ctx context.Context, id string) (Order, error) {
	panic("not implemented")
}
