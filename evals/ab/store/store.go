// Package store reads records from a local cache, falling back to a remote.
package store

import (
	"errors"
	"fmt"
	"strings"
)

// ErrNotFound reports a key that is in neither the cache nor the remote.
var ErrNotFound = errors.New("record not found")

// Manager looks records up in a cache, then in a remote map.
type Manager struct {
	cache  map[string]string
	remote map[string]string
	hits   int
	misses int
}

// NewManager returns a Manager over the given cache and remote maps.
func NewManager(cache, remote map[string]string) *Manager {
	m := new(Manager)
	m.cache = cache
	m.remote = remote
	return m
}

// Get returns the record for key, consulting the cache before the remote.
// Keys are matched case-insensitively after trimming surrounding spaces.
func (m *Manager) Get(key string) (string, error) {
	if key != "" {
		k := strings.ToLower(strings.TrimSpace(key))
		if k != "" {
			if m.cache != nil {
				v, ok := m.cache[k]
				if ok {
					m.hits = m.hits + 1
					return v, nil
				}
			}
			if m.remote != nil {
				v, ok := m.remote[k]
				if ok {
					m.hits = m.hits + 1
					if m.cache != nil {
						m.cache[k] = v
					}
					return v, nil
				}
			}
			m.misses = m.misses + 1
			return "", fmt.Errorf("get %q: %w", key, ErrNotFound)
		}
		m.misses = m.misses + 1
		return "", fmt.Errorf("get %q: %w", key, ErrNotFound)
	}
	m.misses = m.misses + 1
	return "", fmt.Errorf("get %q: %w", key, ErrNotFound)
}

// Stats returns the hit and miss counters.
func (m *Manager) Stats() (hits, misses int) {
	return m.hits, m.misses
}
