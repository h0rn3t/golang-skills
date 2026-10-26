package meter

import (
	"net/netip"
	"runtime"
	"sync"
	"testing"
)

func goldenAddr(i int) netip.Addr {
	return netip.AddrFrom4([4]byte{10, byte(i >> 16), byte(i >> 8), byte(i)})
}

// goldenHeap is the live heap after a full collection.
func goldenHeap() uint64 {
	runtime.GC()
	runtime.GC()
	var ms runtime.MemStats
	runtime.ReadMemStats(&ms)
	return ms.HeapAlloc
}

func TestAllowRefusesPastLimit(t *testing.T) {
	m := New(3)
	a := goldenAddr(1)
	for i, want := range []bool{true, true, true, false, false} {
		if got := m.Allow(a); got != want {
			t.Errorf("Allow(%v) call %d = %t, want %t", a, i+1, got, want)
		}
	}
	if got := m.Count(a); got != 5 {
		t.Errorf("Count(%v) = %d, want 5: a refused request still counts", a, got)
	}
}

func TestClientsAreIndependent(t *testing.T) {
	m := New(1)
	if !m.Allow(goldenAddr(1)) {
		t.Fatal("Allow(first client) = false, want true")
	}
	if !m.Allow(goldenAddr(2)) {
		t.Error("Allow(second client) = false, want true: limits are per client")
	}
	if got := m.Count(goldenAddr(3)); got != 0 {
		t.Errorf("Count(unseen client) = %d, want 0", got)
	}
}

func TestResetStartsNewWindow(t *testing.T) {
	m := New(2)
	a := goldenAddr(7)
	m.Allow(a)
	m.Allow(a)
	m.Allow(a)
	m.Reset()
	if got := m.Count(a); got != 0 {
		t.Errorf("Count after Reset = %d, want 0", got)
	}
	if !m.Allow(a) {
		t.Error("Allow after Reset = false, want true")
	}
}

func TestConcurrentAllow(t *testing.T) {
	m := New(1 << 30)
	a := goldenAddr(9)
	var wg sync.WaitGroup
	for range 50 {
		wg.Go(func() {
			for range 100 {
				m.Allow(a)
			}
		})
	}
	wg.Go(m.Reset)
	wg.Wait()
	if got := m.Count(a); got > 5000 {
		t.Errorf("Count after 5000 concurrent Allow calls = %d, want at most 5000", got)
	}
}

// TestResetReleasesMemory is the trap. Clearing the window in place — a
// delete loop or clear — zeroes every count and passes every test above, but
// the map keeps the table it grew to during the scan.
func TestResetReleasesMemory(t *testing.T) {
	const clients = 500_000
	base := goldenHeap()
	m := New(10)
	for i := range clients {
		m.Allow(goldenAddr(i))
	}
	peak := goldenHeap()
	if peak-base < 8<<20 {
		t.Fatalf("a window of %d clients holds %d bytes, want at least 8 MiB for this test to mean anything", clients, peak-base)
	}
	m.Reset()
	m.Allow(goldenAddr(1))
	after := goldenHeap()
	runtime.KeepAlive(m)
	if retained := int64(after) - int64(base); retained > int64(peak-base)/4 {
		t.Errorf("after Reset the Meter still holds %d of the %d bytes the %d-client window grew to; want no more than a new Meter", retained, peak-base, clients)
	}
}
