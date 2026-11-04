// Package upstream routes requests to backend services. Routes change at
// run time from two sources: the config watcher and the admin API.
package upstream

import (
	"errors"
	"maps"
	"sync/atomic"
)

// ErrNoRoute reports a path prefix with no backend.
var ErrNoRoute = errors.New("upstream: no route")

// Backend is one service a prefix is routed to.
type Backend struct {
	Name string
	URL  string
}

// Registry maps path prefixes to backends. Reads are lock-free and every
// method is safe for concurrent use, including Set and Delete from the
// config watcher and the admin API at the same time.
type Registry struct {
	routes atomic.Pointer[map[string]Backend]
}

// NewRegistry returns a registry holding routes.
func NewRegistry(routes map[string]Backend) *Registry {
	r := &Registry{}
	m := maps.Clone(routes)
	if m == nil {
		m = map[string]Backend{}
	}
	r.routes.Store(&m)
	return r
}

// Lookup returns the backend for prefix.
func (r *Registry) Lookup(prefix string) (Backend, error) {
	b, ok := (*r.routes.Load())[prefix]
	if !ok {
		return Backend{}, ErrNoRoute
	}
	return b, nil
}

// Set routes prefix to b, replacing any earlier backend.
func (r *Registry) Set(prefix string, b Backend) {
	next := maps.Clone(*r.routes.Load())
	next[prefix] = b
	r.routes.Store(&next)
}

// Delete removes the route for prefix.
func (r *Registry) Delete(prefix string) {
	next := maps.Clone(*r.routes.Load())
	delete(next, prefix)
	r.routes.Store(&next)
}

// Len reports how many prefixes are routed.
func (r *Registry) Len() int {
	return len(*r.routes.Load())
}
