package catalog

import (
	"errors"
	"strings"
	"testing"
)

var goldenErrTransport = errors.New("dial tcp: connection refused")

// goldenSource serves a fixed table and a fixed set of failures, and counts
// what it was asked for.
type goldenSource struct {
	calls map[string]int
}

func newGoldenSource() *goldenSource {
	return &goldenSource{calls: map[string]int{}}
}

func (s *goldenSource) Get(sku string) (string, error) {
	s.calls[sku]++
	switch sku {
	case "sku-1":
		return "Widget", nil
	case "sku-2":
		return "Sprocket", nil
	case "sku-boom":
		return "", goldenErrTransport
	case "sku-shut":
		return "", ErrClosed
	}
	return "", ErrNotFound
}

// TestResolveErrorReachesEveryReason is the trap. Adding the SKU to the
// message with %v satisfies the operator and silently cuts the caller off from
// the reason, so one error has to meet both obligations.
func TestResolveErrorReachesEveryReason(t *testing.T) {
	tests := []struct {
		name string
		sku  string
		want error
	}{
		{name: "closed source", sku: "sku-shut", want: ErrClosed},
		{name: "transport failure", sku: "sku-boom", want: goldenErrTransport},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := Resolve(newGoldenSource(), []string{"sku-1", tt.sku, "sku-2"})
			if err == nil {
				t.Fatalf("Resolve(%q) error = nil, want non-nil", tt.sku)
			}
			if !errors.Is(err, tt.want) {
				t.Errorf("errors.Is(Resolve(%q) error, %v) = false, want true; error was %q", tt.sku, tt.want, err)
			}
			if !strings.Contains(err.Error(), tt.sku) {
				t.Errorf("Resolve(%q) error = %q, want the SKU in the message", tt.sku, err)
			}
		})
	}
}

func TestResolveSkipsUnknownAndKeepsGoing(t *testing.T) {
	got, err := Resolve(newGoldenSource(), []string{"sku-2", "sku-404", "sku-1"})
	if err != nil {
		t.Fatalf("Resolve error = %v, want nil", err)
	}

	want := map[string]string{"sku-2": "Sprocket", "sku-1": "Widget"}
	if len(got) != len(want) {
		t.Fatalf("Resolve returned %d entries, want %d: %v", len(got), len(want), got)
	}
	for sku, name := range want {
		if got[sku] != name {
			t.Errorf("Resolve()[%q] = %q, want %q", sku, got[sku], name)
		}
	}
}

// TestResolveCostsOneRoundTripPerSKU pins the documented cost. A plain walk
// over skus asks the source twice for a SKU that appears twice.
func TestResolveCostsOneRoundTripPerSKU(t *testing.T) {
	src := newGoldenSource()

	got, err := Resolve(src, []string{"sku-1", "sku-2", "sku-1", "sku-404", "sku-1", "sku-404"})
	if err != nil {
		t.Fatalf("Resolve error = %v, want nil", err)
	}
	if len(got) != 2 {
		t.Errorf("Resolve returned %d entries, want 2: %v", len(got), got)
	}
	for sku, calls := range src.calls {
		if calls != 1 {
			t.Errorf("the source was asked for %q %d times, want 1", sku, calls)
		}
	}
}

func TestResolveEmptyInput(t *testing.T) {
	got, err := Resolve(newGoldenSource(), nil)
	if err != nil {
		t.Fatalf("Resolve(nil) error = %v, want nil", err)
	}
	if len(got) != 0 {
		t.Errorf("Resolve(nil) = %v, want no entries", got)
	}
}
