package fetch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"testing/synctest"
	"time"
)

// goldenPartner is the partner API as a RoundTripper: no sockets, so the
// client's waits run on synctest's clock and every send is on the record.
type goldenPartner struct {
	// script is the status each successive request gets; the last entry
	// repeats. 0 is a transport failure.
	script []int
	// retryAfter, when set, goes out with every non-2xx response.
	retryAfter string
	requests   []goldenRequest
}

type goldenRequest struct {
	method, path, body string
	at                 time.Time
}

var goldenErrTransport = errors.New("dial tcp: connection refused")

func (p *goldenPartner) RoundTrip(req *http.Request) (*http.Response, error) {
	body := ""
	if req.Body != nil {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			return nil, err
		}
		body = string(b)
	}
	p.requests = append(p.requests, goldenRequest{method: req.Method, path: req.URL.Path, body: body, at: time.Now()})
	status := p.script[min(len(p.requests), len(p.script))-1]
	if status == 0 {
		return nil, goldenErrTransport
	}
	resp := &http.Response{
		StatusCode: status,
		Header:     http.Header{},
		Body:       io.NopCloser(strings.NewReader(`{"id":"o-7","sku":"sku-1","amount":250}`)),
		Request:    req,
	}
	if status >= 300 {
		resp.Body = io.NopCloser(strings.NewReader(`{"error":"no"}`))
		if p.retryAfter != "" {
			resp.Header.Set("Retry-After", p.retryAfter)
		}
	}
	return resp, nil
}

func goldenClient(p *goldenPartner) *Client {
	return &Client{HTTP: &http.Client{Transport: p}, BaseURL: "https://partner.example"}
}

// TestPlaceOrderIsSentOnce is the trap. GetOrder and PlaceOrder share every
// line but the method, so one retrying helper serves both — and the partner
// ships one order per POST it received, answered or not.
func TestPlaceOrderIsSentOnce(t *testing.T) {
	for _, tt := range []struct {
		name   string
		script []int
		want   error
	}{
		{name: "500", script: []int{500}, want: ErrUnavailable},
		{name: "503 then 201", script: []int{503, 201}, want: ErrUnavailable},
		{name: "429", script: []int{429}, want: ErrUnavailable},
		{name: "transport failure", script: []int{0}, want: ErrUnavailable},
	} {
		t.Run(tt.name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				p := &goldenPartner{script: tt.script}

				_, err := goldenClient(p).PlaceOrder(t.Context(), "sku-1", 250)
				if !errors.Is(err, tt.want) {
					t.Errorf("PlaceOrder() error = %v, want %v", err, tt.want)
				}
				if len(p.requests) != 1 {
					t.Errorf("PlaceOrder() sent %d requests, want 1", len(p.requests))
				}
			})
		})
	}
}

func TestPlaceOrderRequest(t *testing.T) {
	p := &goldenPartner{script: []int{201}}

	got, err := goldenClient(p).PlaceOrder(t.Context(), "sku-1", 250)
	if err != nil {
		t.Fatalf("PlaceOrder() error = %v, want nil", err)
	}
	if want := (Order{ID: "o-7", SKU: "sku-1", Amount: 250}); got != want {
		t.Errorf("PlaceOrder() = %+v, want %+v", got, want)
	}
	if len(p.requests) != 1 {
		t.Fatalf("PlaceOrder() sent %d requests, want 1", len(p.requests))
	}
	req := p.requests[0]
	if req.method != http.MethodPost || req.path != "/orders" {
		t.Errorf("PlaceOrder() sent %s %s, want POST /orders", req.method, req.path)
	}
	var body map[string]any
	if err := json.Unmarshal([]byte(req.body), &body); err != nil {
		t.Fatalf("PlaceOrder() body %q is not JSON: %v", req.body, err)
	}
	if body["sku"] != "sku-1" || body["amount"] != float64(250) {
		t.Errorf("PlaceOrder() body = %s, want sku-1 and 250", req.body)
	}
}

// TestGetOrderRetriesWithGrowingPauses pins the documented recovery pattern:
// the read is resent, the first wait is at least Pause, no wait is shorter
// than the one before it, and the waits grow over the run. Jitter added on
// top of the schedule passes; a constant pause, or one shaved below Pause,
// does not.
func TestGetOrderRetriesWithGrowingPauses(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &goldenPartner{script: []int{503, 0, 503, 200}}

		got, err := goldenClient(p).GetOrder(t.Context(), "o-7")
		if err != nil {
			t.Fatalf("GetOrder() error = %v, want nil after three retries", err)
		}
		if got.ID != "o-7" {
			t.Errorf("GetOrder().ID = %q, want %q", got.ID, "o-7")
		}
		if len(p.requests) != 4 {
			t.Fatalf("GetOrder() sent %d requests, want 4", len(p.requests))
		}
		var waits []time.Duration
		for i := 1; i < len(p.requests); i++ {
			waits = append(waits, p.requests[i].at.Sub(p.requests[i-1].at))
		}
		if waits[0] < DefaultPause {
			t.Errorf("GetOrder() first resend came after %v, want at least Pause (%v)", waits[0], DefaultPause)
		}
		for i := 1; i < len(waits); i++ {
			if waits[i] < waits[i-1] {
				t.Errorf("GetOrder() waits were %v, want none shorter than the one before it", waits)
				break
			}
		}
		if waits[len(waits)-1] <= waits[0] {
			t.Errorf("GetOrder() waits were %v, want them to grow over the run", waits)
		}
	})
}

func TestGetOrderGivesUpWithinAttempts(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &goldenPartner{script: []int{503}}
		c := goldenClient(p)
		c.Attempts = 3

		_, err := c.GetOrder(t.Context(), "o-7")
		if !errors.Is(err, ErrUnavailable) {
			t.Errorf("GetOrder() error = %v, want %v", err, ErrUnavailable)
		}
		if err == nil || !strings.Contains(err.Error(), "o-7") || !strings.Contains(err.Error(), "503") {
			t.Errorf("GetOrder() error = %v, want the id and the status in the message", err)
		}
		if len(p.requests) != 3 {
			t.Errorf("GetOrder() sent %d requests, want Attempts = 3", len(p.requests))
		}
	})
}

func TestGetOrderHonorsRetryAfter(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &goldenPartner{script: []int{429, 200}, retryAfter: "3"}

		if _, err := goldenClient(p).GetOrder(t.Context(), "o-7"); err != nil {
			t.Fatalf("GetOrder() error = %v, want nil", err)
		}
		if len(p.requests) != 2 {
			t.Fatalf("GetOrder() sent %d requests, want 2", len(p.requests))
		}
		if gap := p.requests[1].at.Sub(p.requests[0].at); gap < 3*time.Second {
			t.Errorf("GetOrder() resent after %v with Retry-After: 3, want at least 3s", gap)
		}
	})
}

func TestGetOrderDoesNotRetryRejections(t *testing.T) {
	p := &goldenPartner{script: []int{404}}

	_, err := goldenClient(p).GetOrder(t.Context(), "nope")
	if !errors.Is(err, ErrRejected) {
		t.Errorf("GetOrder(unknown) error = %v, want %v", err, ErrRejected)
	}
	if len(p.requests) != 1 {
		t.Errorf("GetOrder(unknown) sent %d requests, want 1", len(p.requests))
	}
}

func TestGetOrderStopsOnCancel(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		p := &goldenPartner{script: []int{503}}
		ctx, cancel := context.WithCancel(t.Context())
		c := goldenClient(p)
		c.Pause = time.Minute

		done := make(chan error, 1)
		go func() {
			_, err := c.GetOrder(ctx, "o-7")
			done <- err
		}()
		time.Sleep(time.Second)
		cancel()
		synctest.Wait()

		select {
		case err := <-done:
			if !errors.Is(err, context.Canceled) {
				t.Errorf("GetOrder(cancelled) error = %v, want %v", err, context.Canceled)
			}
		default:
			t.Fatal("GetOrder(cancelled) is still waiting out its pause")
		}
		if len(p.requests) != 1 {
			t.Errorf("GetOrder(cancelled) sent %d requests, want 1", len(p.requests))
		}
	})
}
