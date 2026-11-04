package upstream

import (
	"errors"
	"testing"
)

func TestRegistry(t *testing.T) {
	r := NewRegistry(map[string]Backend{"/a": {Name: "a", URL: "https://a.internal"}})
	r.Set("/b", Backend{Name: "b", URL: "https://b.internal"})
	if got, err := r.Lookup("/b"); err != nil || got.Name != "b" {
		t.Fatalf("Lookup(/b) = %v, %v; want b", got, err)
	}
	r.Delete("/a")
	if _, err := r.Lookup("/a"); !errors.Is(err, ErrNoRoute) {
		t.Fatalf("Lookup(/a) after Delete error = %v, want ErrNoRoute", err)
	}
	if got := r.Len(); got != 1 {
		t.Fatalf("Len() = %d, want 1", got)
	}
}
