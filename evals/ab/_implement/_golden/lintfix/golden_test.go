package lintfix

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func goldenFile(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "sync.json")
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestLoadConfigReadsTheFile(t *testing.T) {
	path := goldenFile(t, `{"endpoint":"https://partner.example/catalog","retries":3,"base":250000000}`)
	got, err := LoadConfig(path)
	if err != nil {
		t.Fatalf("LoadConfig(valid) error = %v, want nil", err)
	}
	if want := (Config{Endpoint: "https://partner.example/catalog", Retries: 3, Base: 250 * time.Millisecond}); got != want {
		t.Errorf("LoadConfig(valid) = %+v, want %+v", got, want)
	}
}

// TestLoadConfigRejectsInvalidDocument is the real finding: the unchecked
// Unmarshal turned a broken settings file into a zero Config.
func TestLoadConfigRejectsInvalidDocument(t *testing.T) {
	for _, body := range []string{`{"endpoint":`, `not json`, `{"retries":"three"}`} {
		if _, err := LoadConfig(goldenFile(t, body)); err == nil {
			t.Errorf("LoadConfig(%q) error = nil, want an error", body)
		}
	}
	if _, err := LoadConfig(filepath.Join(t.TempDir(), "missing.json")); err == nil {
		t.Error("LoadConfig(missing file) error = nil, want an error")
	}
}

// TestBackoffKeepsItsJitter holds the false positive: the jitter stays in its
// documented range and stays random, whatever silenced or replaced G404.
func TestBackoffKeepsItsJitter(t *testing.T) {
	const base = 100 * time.Millisecond
	for n := range 4 {
		d := base << n
		seen := map[time.Duration]bool{}
		for range 64 {
			got := Backoff(base, n)
			if got < d || got >= 2*d {
				t.Fatalf("Backoff(%v, %d) = %v, want in [%v, %v)", base, n, got, d, 2*d)
			}
			seen[got] = true
		}
		if len(seen) < 2 {
			t.Errorf("Backoff(%v, %d) returned one value over 64 calls, want jitter", base, n)
		}
	}
}
