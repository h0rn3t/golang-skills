package tally

import (
	"strconv"
	"sync"
	"testing"
)

func TestGoldenCounter(t *testing.T) {
	c := New()
	var wg sync.WaitGroup
	wg.Go(func() { // the metrics endpoint reads while requests count
		for range 50 {
			_ = c.Snapshot()
			_ = c.Get("k0")
		}
	})
	for g := range 16 {
		wg.Go(func() {
			for i := range 1000 {
				c.Inc("k" + strconv.Itoa(i%8))
				if i%100 == 0 {
					c.Add("big", int64(g))
				}
			}
		})
	}
	wg.Wait()
	for i := range 8 {
		key := "k" + strconv.Itoa(i)
		if got := c.Get(key); got != 2000 {
			t.Errorf("Get(%q) = %d, want 2000", key, got)
		}
	}
	if got := c.Get("big"); got != 10*120 {
		t.Errorf("Get(%q) = %d, want %d", "big", got, 10*120)
	}
	if got := c.Get("never"); got != 0 {
		t.Errorf("Get(%q) = %d, want 0", "never", got)
	}

	snap := c.Snapshot()
	if len(snap) != 9 || snap["k3"] != 2000 || snap["big"] != 1200 {
		t.Errorf("Snapshot() = %v, want 9 keys with k3=2000 and big=1200", snap)
	}
	snap["k3"] = -1
	delete(snap, "k4")
	if c.Get("k3") != 2000 || c.Get("k4") != 2000 {
		t.Errorf("changing the Snapshot changed the Counter: k3=%d k4=%d", c.Get("k3"), c.Get("k4"))
	}
}

func TestGoldenString(t *testing.T) {
	c := New()
	if got := c.String(); got != "" {
		t.Errorf("New().String() = %q, want empty", got)
	}
	c.Add("b", 2)
	c.Inc("a")
	c.Add("c", -3)
	if got, want := c.String(), "a=1 b=2 c=-3"; got != want {
		t.Errorf("String() = %q, want %q", got, want)
	}
	if snap := New().Snapshot(); snap == nil || len(snap) != 0 {
		t.Errorf("New().Snapshot() = %#v, want an empty non-nil map", snap)
	}
}
