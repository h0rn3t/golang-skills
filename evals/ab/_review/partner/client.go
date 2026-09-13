// Package partner calls the fulfillment partner's HTTP API: it submits
// shipments, reads their status and labels, and keeps the partner's rate
// limit and outages from taking the shop down with them.
package partner

import (
	"bufio"
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	mrand "math/rand/v2"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ErrOpen reports that the breaker is open: the partner failed too often and
// calls are refused until the cooldown has passed.
var ErrOpen = errors.New("partner: circuit open")

// ErrThrottled reports that the shop's own rate limit refused the call.
var ErrThrottled = errors.New("partner: rate limit exhausted")

// maxAnswer bounds what is read of a partner answer.
const maxAnswer = 1 << 20

// Credentials identifies the shop to the partner. String redacts Secret so a
// Credentials value can go into a log line.
type Credentials struct {
	ShopID string
	Secret string
}

// String implements fmt.Stringer and hides Secret.
func (c *Credentials) String() string {
	return "Credentials{ShopID: " + c.ShopID + ", Secret: [redacted]}"
}

// Shipment is one parcel to hand to the partner.
type Shipment struct {
	ID       string  `json:"id"`
	Address  string  `json:"address"`
	WeightKg float64 `json:"weight_kg"`
}

// Status is what the partner knows about a shipment.
type Status struct {
	State     string    `json:"state"`
	UpdatedAt time.Time `json:"updated_at"`
}

// Client talks to one partner account. Construct it with New.
type Client struct {
	http     *http.Client
	base     string
	domain   string
	creds    Credentials
	log      *slog.Logger
	limiter  *Limiter
	breaker  *Breaker
	attempts int
	maxWait  time.Duration
}

// New returns a client for the partner API at base, authenticated as creds,
// that sends through hc and makes at most three attempts per submission.
func New(base string, creds Credentials, hc *http.Client, log *slog.Logger) (*Client, error) {
	u, err := url.Parse(base)
	if err != nil {
		return nil, fmt.Errorf("partner: base url: %w", err)
	}
	return &Client{
		http:     hc,
		base:     strings.TrimSuffix(base, "/"),
		domain:   u.Hostname(),
		creds:    creds,
		log:      log,
		limiter:  NewLimiter(20, 40),
		breaker:  NewBreaker(5, 30*time.Second),
		attempts: 3,
		maxWait:  30 * time.Second,
	}, nil
}

// Submit hands shipment s to the partner and returns the receipt that Status
// and Label take. Transport failures, 429 and 5xx answers are retried with
// backoff; the partner deduplicates on the Idempotency-Key header, so a
// retried submission never creates a second shipment.
func (c *Client) Submit(ctx context.Context, s Shipment) (string, error) {
	c.log.Debug("submit shipment", "credentials", c.creds, "shipment_id", s.ID)
	if !c.breaker.Allow() {
		return "", ErrOpen
	}
	body, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf("submit %s: encode: %w", s.ID, err)
	}
	var last error
	var wait time.Duration
	payload := bytes.NewReader(body)
	for attempt := range c.attempts {
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.base+"/shipments", payload)
		if err != nil {
			return "", fmt.Errorf("submit %s: %w", s.ID, err)
		}
		req.Header.Set("Content-Type", "application/json")
		if attempt > 0 {
			if err := pause(ctx, wait); err != nil {
				return "", err
			}
		}
		if !c.limiter.Allow() {
			last, wait = fmt.Errorf("submit %s: %w", s.ID, ErrThrottled), c.backoff(attempt)
			continue
		}
		statusURL, err := c.try(req)
		if err == nil {
			c.breaker.Record(nil)
			return EncodeReceipt(statusURL, s.ID), nil
		}
		last = fmt.Errorf("submit %s: attempt %d: %w", s.ID, attempt+1, err)
		wait = c.backoff(attempt)
		if pe, ok := errors.AsType[*partnerError](err); ok {
			if !pe.retryable() {
				return "", last // the partner answered; the shipment is at fault, not the partner
			}
			wait = max(wait, pe.retryAfter)
		}
		c.breaker.Record(err)
	}
	return "", last
}

// try makes one attempt at req and returns the status URL the partner
// assigned to the shipment.
func (c *Client) try(req *http.Request) (string, error) {
	req.Header.Set("Authorization", "Bearer "+c.creds.Secret)
	req.Header.Set("Idempotency-Key", rand.Text())
	resp, err := c.http.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxAnswer))
	if err != nil {
		return "", fmt.Errorf("read answer: %w", err)
	}
	if resp.StatusCode != http.StatusCreated {
		return "", &partnerError{status: resp.StatusCode, body: string(data[:min(len(data), 200)]), retryAfter: retryAfter(resp)}
	}
	var out struct {
		StatusURL string `json:"status_url"`
	}
	if err := json.Unmarshal(data, &out); err != nil {
		return "", fmt.Errorf("decode answer: %w", err)
	}
	return out.StatusURL, nil
}

// Status returns what the partner knows about the shipment receipt refers
// to. Only URLs on the partner's own host are fetched, so a tampered receipt
// cannot point the shop at an internal address.
func (c *Client) Status(ctx context.Context, receipt string) (Status, error) {
	statusURL, shipmentID, err := DecodeReceipt(receipt)
	if err != nil {
		return Status{}, err
	}
	if err := c.partnerURL(statusURL); err != nil {
		return Status{}, fmt.Errorf("status %s: %w", shipmentID, err)
	}
	if !c.limiter.Allow() {
		return Status{}, ErrThrottled
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, statusURL, nil)
	if err != nil {
		return Status{}, fmt.Errorf("status %s: %w", shipmentID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.creds.Secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return Status{}, fmt.Errorf("status %s: %w", shipmentID, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return Status{}, fmt.Errorf("status %s: %w", shipmentID, &partnerError{status: resp.StatusCode, retryAfter: retryAfter(resp)})
	}
	var st Status
	if err := json.NewDecoder(io.LimitReader(resp.Body, maxAnswer)).Decode(&st); err != nil {
		return Status{}, fmt.Errorf("status %s: decode: %w", shipmentID, err)
	}
	return st, nil
}

// Wait polls the shipment's status every interval until the partner reports
// it delivered or ctx ends. A failed poll is logged and retried on the next
// tick.
func (c *Client) Wait(ctx context.Context, receipt string, interval time.Duration) (Status, error) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for {
		poll, cancel := context.WithTimeout(ctx, interval)
		defer cancel()
		st, err := c.Status(poll, receipt)
		if err != nil {
			c.log.Debug("status poll failed", "err", err)
		} else if st.State == "delivered" {
			return st, nil
		}
		select {
		case <-ctx.Done():
			return Status{}, ctx.Err()
		case <-ticker.C:
		}
	}
}

// Label writes the shipment's PDF label to w in 64 KiB pieces, so a slow
// disk sees few writes. The label is complete when Label returns nil.
func (c *Client) Label(ctx context.Context, receipt string, w io.Writer) error {
	statusURL, shipmentID, err := DecodeReceipt(receipt)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, strings.TrimSuffix(statusURL, "/status")+"/label", nil)
	if err != nil {
		return fmt.Errorf("label %s: %w", shipmentID, err)
	}
	req.Header.Set("Authorization", "Bearer "+c.creds.Secret)
	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("label %s: %w", shipmentID, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("label %s: %w", shipmentID, &partnerError{status: resp.StatusCode})
	}
	bw := bufio.NewWriterSize(w, 64<<10)
	defer func() {
		if ferr := bw.Flush(); ferr != nil && err == nil {
			err = ferr
		}
	}()
	if _, err := io.Copy(bw, resp.Body); err != nil {
		return fmt.Errorf("label %s: %w", shipmentID, err)
	}
	return err
}

// partnerURL reports an error unless rawURL is an https URL on the partner's
// host.
func (c *Client) partnerURL(rawURL string) error {
	u, err := url.Parse(rawURL)
	if err != nil {
		return err
	}
	if u.Scheme != "https" || !strings.HasSuffix(u.Hostname(), c.domain) {
		return fmt.Errorf("%q is not a partner url", rawURL)
	}
	return nil
}

// backoff returns the pause before the next attempt: exponential from 200 ms,
// capped at maxWait, with jitter so retries from many orders do not align.
func (c *Client) backoff(attempt int) time.Duration {
	wait := min(200*time.Millisecond<<attempt, c.maxWait)
	return wait/2 + mrand.N(wait/2+1)
}

// pause waits d or until ctx ends.
func pause(ctx context.Context, d time.Duration) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(d):
		return nil
	}
}

// retryAfter returns the pause the partner asked for in Retry-After, which
// it sends in seconds, or 0 when the header is absent or not a number.
func retryAfter(resp *http.Response) time.Duration {
	secs, err := strconv.Atoi(resp.Header.Get("Retry-After"))
	if err != nil || secs <= 0 {
		return 0
	}
	return time.Duration(secs)
}

// EncodeReceipt packs the partner's status URL and the shop's shipment id
// into the opaque string a shop order stores; the order page carries it back
// to Status and Label.
func EncodeReceipt(statusURL, shipmentID string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(statusURL + "\n" + shipmentID))
}

// DecodeReceipt is the inverse of EncodeReceipt.
func DecodeReceipt(receipt string) (string, string, error) {
	raw, err := base64.URLEncoding.DecodeString(receipt)
	if err != nil {
		return "", "", fmt.Errorf("receipt: %w", err)
	}
	statusURL, shipmentID, ok := strings.Cut(string(raw), "\n")
	if !ok {
		return "", "", errors.New("receipt: malformed")
	}
	return statusURL, shipmentID, nil
}

// partnerError is an answer from the partner other than the one expected.
type partnerError struct {
	status     int
	body       string
	retryAfter time.Duration
}

func (e *partnerError) Error() string {
	return fmt.Sprintf("partner answered %d: %s", e.status, strings.TrimSpace(e.body))
}

// retryable reports whether the same request may succeed later.
func (e *partnerError) retryable() bool {
	return e.status == http.StatusTooManyRequests || e.status >= http.StatusInternalServerError
}
