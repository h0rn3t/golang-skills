package partner

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// serve starts a TLS partner stub running h and returns it with a client
// that trusts it.
func serve(t *testing.T, h http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewTLSServer(h)
	t.Cleanup(srv.Close)
	c, err := New(srv.URL, Credentials{ShopID: "shop-1", Secret: "s3cr3t"}, srv.Client(), slog.New(slog.DiscardHandler))
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}
	return srv, c
}

func TestSubmitReturnsReceipt(t *testing.T) {
	srv, c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/shipments" || r.Header.Get("Idempotency-Key") == "" {
			http.Error(w, "unexpected request", http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]string{"status_url": "https://" + r.Host + "/shipments/p-1/status"})
	})
	receipt, err := c.Submit(t.Context(), Shipment{ID: "ord-01", Address: "1 Main St", WeightKg: 1.5})
	if err != nil {
		t.Fatalf("Submit() error = %v", err)
	}
	statusURL, shipmentID, err := DecodeReceipt(receipt)
	if err != nil {
		t.Fatalf("DecodeReceipt() error = %v", err)
	}
	if want := srv.URL + "/shipments/p-1/status"; statusURL != want || shipmentID != "ord-01" {
		t.Errorf("receipt = (%q, %q), want (%q, %q)", statusURL, shipmentID, want, "ord-01")
	}
}

func TestSubmitDoesNotRetryClientErrors(t *testing.T) {
	var calls atomic.Int32
	_, c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		http.Error(w, "address unknown", http.StatusUnprocessableEntity)
	})
	if _, err := c.Submit(t.Context(), Shipment{ID: "ord-02"}); err == nil {
		t.Fatal("Submit() = nil error, want the partner's 422")
	}
	if got := calls.Load(); got != 1 {
		t.Errorf("partner called %d times, want 1", got)
	}
}

func TestStatusReadsState(t *testing.T) {
	srv, c := serve(t, func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(Status{State: "in_transit", UpdatedAt: time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)})
	})
	st, err := c.Status(t.Context(), EncodeReceipt(srv.URL+"/shipments/p-1/status", "ord-01"))
	if err != nil {
		t.Fatalf("Status() error = %v", err)
	}
	if st.State != "in_transit" {
		t.Errorf("Status().State = %q, want %q", st.State, "in_transit")
	}
}

func TestStatusRejectsForeignHost(t *testing.T) {
	var calls atomic.Int32
	_, c := serve(t, func(w http.ResponseWriter, r *http.Request) { calls.Add(1) })
	if _, err := c.Status(t.Context(), EncodeReceipt("https://evil.example/status", "ord-01")); err == nil {
		t.Fatal("Status() = nil error, want a rejected host")
	}
	if calls.Load() != 0 {
		t.Error("the foreign url was fetched")
	}
}

func TestLimiterRefills(t *testing.T) {
	clock := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	l := NewLimiter(2, 2)
	l.now = func() time.Time { return clock }
	l.last = clock
	for i := range 2 {
		if !l.Allow() {
			t.Fatalf("Allow() #%d = false, want true from a full bucket", i+1)
		}
	}
	if l.Allow() {
		t.Fatal("Allow() = true, want false from an empty bucket")
	}
	clock = clock.Add(time.Second)
	if !l.Allow() {
		t.Fatal("Allow() = false, want true after a second of refill")
	}
}

func TestBreakerOpensThenProbes(t *testing.T) {
	clock := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	b := NewBreaker(2, time.Minute)
	b.now = func() time.Time { return clock }
	down := errors.New("down")
	b.Record(down)
	b.Record(down)
	if b.Allow() {
		t.Fatal("Allow() = true, want false after two failures")
	}
	clock = clock.Add(time.Minute)
	if !b.Allow() {
		t.Fatal("Allow() = false, want a probe after the cooldown")
	}
}
