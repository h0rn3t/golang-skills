package catalog

import (
	"errors"
	"strings"
	"testing"
)

// goldenSource serves a fixed table and a fixed failure.
type goldenSource struct {
	names map[string]string
	fail  map[string]error
}

func (s goldenSource) Get(sku string) (string, error) {
	if err, ok := s.fail[sku]; ok {
		return "", err
	}
	if name, ok := s.names[sku]; ok {
		return name, nil
	}
	return "", ErrNotFound
}

var errTransport = errors.New("dial tcp: connection refused")

func newGoldenSource() goldenSource {
	return goldenSource{
		names: map[string]string{"sku-1": "Widget", "sku-2": "Sprocket"},
		fail:  map[string]error{"sku-boom": errTransport, "sku-shut": ErrClosed},
	}
}

// TestLookupErrorReachesEveryReason is the trap. Adding the SKU to the message
// with %v satisfies the operator and silently cuts the caller off from the
// reason, so both obligations have to be met by one error.
func TestLookupErrorReachesEveryReason(t *testing.T) {
	src := newGoldenSource()
	tests := []struct {
		name string
		sku  string
		want error
	}{
		{name: "unknown sku", sku: "sku-404", want: ErrNotFound},
		{name: "closed source", sku: "sku-shut", want: ErrClosed},
		{name: "transport failure", sku: "sku-boom", want: errTransport},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Lookup(src, tt.sku)
			if err == nil {
				t.Fatalf("Lookup(%q) error = nil, want non-nil", tt.sku)
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("errors.Is(Lookup(%q) error, %v) = false, want true; error was %q", tt.sku, tt.want, err)
			}
			if !strings.Contains(err.Error(), tt.sku) {
				t.Errorf("Lookup(%q) error = %q, want the SKU in the message", tt.sku, err)
			}
		})
	}
}

func TestLookupAllPropagatesReason(t *testing.T) {
	src := newGoldenSource()

	_, err := LookupAll(src, []string{"sku-1", "sku-boom", "sku-2"})
	if err == nil {
		t.Fatal("LookupAll(with a transport failure) error = nil, want non-nil")
	}
	if !errors.Is(err, errTransport) {
		t.Errorf("errors.Is(LookupAll error, errTransport) = false, want true; error was %q", err)
	}
	if !strings.Contains(err.Error(), "sku-boom") {
		t.Errorf("LookupAll error = %q, want the failing SKU in the message", err)
	}
}

func TestLookupResolves(t *testing.T) {
	src := newGoldenSource()

	got, err := Lookup(src, "sku-1")
	if err != nil {
		t.Fatalf("Lookup(sku-1) error = %v, want nil", err)
	}
	if want := (Product{SKU: "sku-1", Name: "Widget"}); got != want {
		t.Errorf("Lookup(sku-1) = %+v, want %+v", got, want)
	}
}

func TestLookupAllSkipsUnknown(t *testing.T) {
	src := newGoldenSource()

	got, err := LookupAll(src, []string{"sku-2", "sku-404", "sku-1"})
	if err != nil {
		t.Fatalf("LookupAll error = %v, want nil", err)
	}
	want := []Product{{SKU: "sku-2", Name: "Sprocket"}, {SKU: "sku-1", Name: "Widget"}}
	if len(got) != len(want) {
		t.Fatalf("LookupAll returned %d products, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("LookupAll[%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}
