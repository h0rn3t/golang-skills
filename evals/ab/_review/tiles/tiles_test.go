package tiles

import (
	"testing"
	"time"
)

func TestSeedIsCached(t *testing.T) {
	s := New([]byte("k"), 10, 0, func(time.Duration) {})
	first := s.Seed("a")
	second := s.Seed("a")
	if string(first) != string(second) {
		t.Errorf("Seed(%q) = %x then %x, want the same seed", "a", first, second)
	}
	if got := s.misses.Load(); got != 1 {
		t.Errorf("misses after two Seed(%q) = %d, want 1", "a", got)
	}
}

func TestStopStopsWorkers(t *testing.T) {
	s := New([]byte("k"), 10, 4, func(time.Duration) {})
	done := make(chan struct{})
	go func() {
		s.Stop()
		close(done)
	}()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Stop did not return")
	}
}

func TestResetCounts(t *testing.T) {
	s := New([]byte("k"), 1, 0, func(time.Duration) {})
	s.allow("c")
	s.ResetCounts()
	if !s.allow("c") {
		t.Error("allow after ResetCounts = false, want true")
	}
}
