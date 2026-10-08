// Package profile serves user profiles from memory in front of the profile
// service.
package profile

import (
	"context"
	"time"
)

// Profile is what the profile service returns for one user.
type Profile struct {
	Tenant      string
	UserID      string
	DisplayName string
}

// Fetch asks the profile service for userID. The service answers for the
// tenant ctx carries (see WithTenant): user IDs are numbered per tenant, so
// user 7 of one tenant and user 7 of another are different people.
type Fetch func(ctx context.Context, userID string) (Profile, error)

// WithTenant returns a copy of ctx that carries tenant.
func WithTenant(ctx context.Context, tenant string) context.Context {
	panic("not implemented")
}

// TenantOf returns the tenant ctx carries, or "" when it carries none.
func TenantOf(ctx context.Context) string {
	panic("not implemented")
}

// Cache keeps profiles in memory so most page views never reach the profile
// service.
//
// A call to the service is a network round trip of about 300ms, and the
// service grants this process a small request budget. The last incident was a
// deploy: the cache started empty, the home page asks for the signed-in
// user's profile from a dozen widgets at once, and every one of those
// requests went to the service. Profiles of different users never wait on
// each other: a slow answer from the service holds up only the requests for
// that same profile.
type Cache struct {
	// The unexported fields are the implementation's own.
}

// New returns a Cache that asks fetch for a profile it does not hold and keeps
// each answer for ttl. New starts no goroutine. A Cache is safe for concurrent
// use.
func New(fetch Fetch, ttl time.Duration) *Cache {
	panic("not implemented")
}

// Get returns the profile of userID in the tenant ctx carries.
//
// An answer is served from memory until ttl has passed since it was fetched.
// A failed fetch is returned to every caller waiting on it and is not kept:
// the next Get asks the service again.
func (c *Cache) Get(ctx context.Context, userID string) (Profile, error) {
	panic("not implemented")
}
