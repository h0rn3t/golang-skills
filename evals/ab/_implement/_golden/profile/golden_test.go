package profile

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

var errGoldenDown = errors.New("profile service: 503")

// goldenService stands in for the profile service. It answers for the tenant
// in ctx, counts its calls, holds every call until release is closed when
// release is set, and fails while down is set.
type goldenService struct {
	calls   atomic.Int32
	release chan struct{}
	down    atomic.Bool
}

func (s *goldenService) fetch(ctx context.Context, userID string) (Profile, error) {
	s.calls.Add(1)
	if s.release != nil {
		<-s.release
	}
	if s.down.Load() {
		return Profile{}, errGoldenDown
	}
	tenant := TenantOf(ctx)
	return Profile{Tenant: tenant, UserID: userID, DisplayName: tenant + "/" + userID}, nil
}

// TestGetKeepsTenantsApart is the first trap. A key of the user ID alone is
// the shortest cache, and it serves one tenant's user 7 to the other.
func TestGetKeepsTenantsApart(t *testing.T) {
	var svc goldenService
	c := New(svc.fetch, time.Minute)
	for _, tenant := range []string{"acme", "globex", "acme"} {
		got, err := c.Get(WithTenant(t.Context(), tenant), "7")
		if err != nil {
			t.Fatalf("Get(%s, 7) error = %v, want nil", tenant, err)
		}
		if got.Tenant != tenant {
			t.Errorf("Get(%s, 7) = %+v, the profile of tenant %q", tenant, got, got.Tenant)
		}
	}
	if n := svc.calls.Load(); n != 2 {
		t.Errorf("fetches = %d, want 2: one per tenant, the second acme Get from memory", n)
	}
}

// TestGetConcurrentMissesShareOneFetch is the second trap: the deploy
// incident. Checking the map, unlocking, and fetching lets every widget miss.
func TestGetConcurrentMissesShareOneFetch(t *testing.T) {
	svc := goldenService{release: make(chan struct{})}
	c := New(svc.fetch, time.Minute)
	ctx := WithTenant(t.Context(), "acme")
	var wg sync.WaitGroup
	for range 12 {
		wg.Go(func() {
			if got, err := c.Get(ctx, "7"); err != nil || got.UserID != "7" {
				t.Errorf("Get(acme, 7) = %+v, %v; want user 7, nil", got, err)
			}
		})
	}
	time.Sleep(100 * time.Millisecond) // every widget's request has arrived
	close(svc.release)
	wg.Wait()
	if n := svc.calls.Load(); n != 1 {
		t.Errorf("12 concurrent Gets of one cold profile made %d fetches, want 1", n)
	}
}

// TestGetDoesNotWaitOnAnotherProfile is the third trap. Holding the cache's
// lock across the fetch also gives one fetch per profile, and it stalls every
// other profile behind the slow one.
func TestGetDoesNotWaitOnAnotherProfile(t *testing.T) {
	release := make(chan struct{})
	defer close(release)
	c := New(func(ctx context.Context, userID string) (Profile, error) {
		if userID == "slow" {
			<-release
		}
		return Profile{Tenant: TenantOf(ctx), UserID: userID}, nil
	}, time.Minute)
	ctx := WithTenant(t.Context(), "acme")
	if _, err := c.Get(ctx, "cached"); err != nil {
		t.Fatalf("Get(acme, cached) error = %v, want nil", err)
	}
	go func() { _, _ = c.Get(ctx, "slow") }()
	time.Sleep(50 * time.Millisecond) // the slow fetch is in flight
	done := make(chan struct{})
	go func() {
		defer close(done)
		_, _ = c.Get(ctx, "cached")
		_, _ = c.Get(ctx, "cold")
	}()
	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Error("Get of a cached and a cold profile waited on another profile's slow fetch")
	}
}

func TestGetDoesNotKeepFailures(t *testing.T) {
	var svc goldenService
	svc.down.Store(true)
	c := New(svc.fetch, time.Minute)
	ctx := WithTenant(t.Context(), "acme")
	if _, err := c.Get(ctx, "7"); !errors.Is(err, errGoldenDown) {
		t.Fatalf("Get(acme, 7) with the service down: error = %v, want %v", err, errGoldenDown)
	}
	svc.down.Store(false)
	if got, err := c.Get(ctx, "7"); err != nil || got.UserID != "7" {
		t.Errorf("Get(acme, 7) after the service recovered = %+v, %v; want user 7, nil", got, err)
	}
}

func TestGetExpiresAfterTTL(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var svc goldenService
		c := New(svc.fetch, time.Minute)
		ctx := WithTenant(t.Context(), "acme")
		for _, wait := range []time.Duration{0, 0, 59 * time.Second} {
			time.Sleep(wait)
			if _, err := c.Get(ctx, "7"); err != nil {
				t.Fatalf("Get(acme, 7) error = %v, want nil", err)
			}
		}
		if n := svc.calls.Load(); n != 1 {
			t.Errorf("fetches within the ttl = %d, want 1", n)
		}
		time.Sleep(2 * time.Second)
		if _, err := c.Get(ctx, "7"); err != nil {
			t.Fatalf("Get(acme, 7) error = %v, want nil", err)
		}
		if n := svc.calls.Load(); n != 2 {
			t.Errorf("fetches after the ttl = %d, want 2", n)
		}
	})
}
