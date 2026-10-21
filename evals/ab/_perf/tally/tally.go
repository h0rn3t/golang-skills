// Package tally counts events per key for the service's metrics endpoint.
package tally

import (
	"fmt"
	"maps"
	"sort"
	"strings"
	"sync"
)

// Counter counts per key. It is safe for concurrent use.
type Counter struct {
	mu     sync.Mutex
	counts map[string]int64
}

// New returns an empty Counter.
func New() *Counter {
	return &Counter{counts: map[string]int64{}}
}

// Inc adds one to key's count.
func (c *Counter) Inc(key string) {
	c.Add(key, 1)
}

// Add adds n to key's count.
func (c *Counter) Add(key string, n int64) {
	c.mu.Lock()
	c.counts[key] += n
	c.mu.Unlock()
}

// Get returns key's count, 0 for a key never counted.
func (c *Counter) Get(key string) int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.counts[key]
}

// Snapshot returns a copy of every count; changing it does not change the
// Counter.
func (c *Counter) Snapshot() map[string]int64 {
	c.mu.Lock()
	defer c.mu.Unlock()
	return maps.Clone(c.counts)
}

// String renders every count as key=count, sorted by key, for the debug page.
func (c *Counter) String() string {
	snap := c.Snapshot()
	keys := []string{}
	for k := range snap {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := []string{}
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s=%d", k, snap[k]))
	}
	return strings.Join(parts, " ")
}
