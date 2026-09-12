package fetch

import "context"

// PlaceOrder creates an order for sku and amount with POST BaseURL/orders,
// sending {"sku": sku, "amount": amount} as JSON and returning the order the
// partner created from its 201 response.
//
// The partner ships whatever it records. It records an order the moment the
// POST reaches it, before it answers, and it has no way to tell two POSTs
// with the same body apart: it has never heard of an idempotency key. An
// order the customer did not place is a refund, a support ticket, and a
// chargeback fee.
//
// A 4xx other than 429 is ErrRejected. A 5xx, a 429, or a transport failure
// is ErrUnavailable; the returned error names the sku and carries the status
// or transport failure. A cancelled ctx ends the call with ctx.Err().
func (c *Client) PlaceOrder(ctx context.Context, sku string, amount int64) (Order, error) {
	panic("not implemented")
}
