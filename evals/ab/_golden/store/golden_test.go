package store

import (
	"errors"
	"testing"
)

// TestGoldenGet pins key normalization, cache promotion on a remote hit, the
// hit/miss counters, and the not-found error text.
func TestGoldenGet(t *testing.T) {
	cache := map[string]string{"ab": "cached"}
	remote := map[string]string{"cd": "remote"}
	m := NewManager(cache, remote)

	got, err := m.Get("  AB ")
	if err != nil || got != "cached" {
		t.Fatalf("cache hit = (%q, %v), want (\"cached\", nil)", got, err)
	}

	got, err = m.Get("CD")
	if err != nil || got != "remote" {
		t.Fatalf("remote hit = (%q, %v), want (\"remote\", nil)", got, err)
	}
	if cache["cd"] != "remote" {
		t.Fatalf("remote hit must promote into the cache, cache = %v", cache)
	}

	_, err = m.Get("missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("missing key error = %v, want wrapped ErrNotFound", err)
	}
	if err.Error() != `get "missing": record not found` {
		t.Fatalf("missing key error text = %q", err.Error())
	}

	if _, err := m.Get(""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("empty key error = %v, want wrapped ErrNotFound", err)
	}
	if _, err := m.Get("   "); !errors.Is(err, ErrNotFound) {
		t.Fatalf("blank key error = %v, want wrapped ErrNotFound", err)
	}

	hits, misses := m.Stats()
	if hits != 2 || misses != 3 {
		t.Fatalf("Stats() = (%d, %d), want (2, 3)", hits, misses)
	}

	nilMaps := NewManager(nil, nil)
	if _, err := nilMaps.Get("x"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("nil maps must report not found, got %v", err)
	}
}
