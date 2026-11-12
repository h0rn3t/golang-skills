package newsletter

import (
	"context"
	"errors"
	"slices"
	"testing"
)

// goldenOutbox records what the provider was asked to deliver, and fails
// every delivery with err when it is set.
type goldenOutbox struct {
	sent []Message
	err  error
}

func (o *goldenOutbox) Deliver(_ context.Context, m Message) error {
	o.sent = append(o.sent, m)
	return o.err
}

var goldenSubscribers = []string{"ann@example.com", "bo@example.com", "cy@example.com"}

// goldenDigest sends one digest to goldenSubscribers and returns what reached
// the provider.
func goldenDigest(t *testing.T) []Message {
	t.Helper()
	var out goldenOutbox
	if err := NewMailer(&out, "news@example.com").Digest(t.Context(), goldenSubscribers, "This week"); err != nil {
		t.Fatalf("Digest(3 subscribers) error = %v, want nil", err)
	}
	return out.sent
}

// TestGoldenDigestIsOneMessage is the request: the provider bills per message,
// and every subscriber still has to get the digest.
func TestGoldenDigestIsOneMessage(t *testing.T) {
	sent := goldenDigest(t)
	if len(sent) != 1 {
		t.Fatalf("Digest(3 subscribers) delivered %d messages, want 1", len(sent))
	}
	m := sent[0]
	if m.From != "news@example.com" || m.Subject != "Weekly digest" || m.Body != "This week" {
		t.Errorf("Digest() message = %+v, want it from news@example.com with the subject Weekly digest and the text", m)
	}
	reached := slices.Concat(m.To, m.Cc, m.Bcc)
	for _, s := range goldenSubscribers {
		if !slices.Contains(reached, s) {
			t.Errorf("Digest() message does not reach %s: %+v", s, m)
		}
	}
}

// TestGoldenDigestHidesSubscribers is what the request does not say: To and Cc
// are shown to every recipient, so a subscriber list there hands each
// subscriber the addresses of all the others.
func TestGoldenDigestHidesSubscribers(t *testing.T) {
	for _, m := range goldenDigest(t) {
		var shown []string
		for _, address := range slices.Concat(m.To, m.Cc) {
			if slices.Contains(goldenSubscribers, address) {
				shown = append(shown, address)
			}
		}
		if len(shown) > 1 {
			t.Errorf("Digest() shows %v to every recipient of one message; want the subscribers in Bcc", shown)
		}
	}
}

func TestGoldenDigestWithoutSubscribersSendsNothing(t *testing.T) {
	var out goldenOutbox
	if err := NewMailer(&out, "news@example.com").Digest(t.Context(), nil, "This week"); err != nil {
		t.Fatalf("Digest(no subscribers) error = %v, want nil", err)
	}
	if len(out.sent) != 0 {
		t.Errorf("Digest(no subscribers) delivered %+v, want nothing", out.sent)
	}
}

func TestGoldenDigestReturnsTheProviderError(t *testing.T) {
	errDown := errors.New("provider down")
	out := goldenOutbox{err: errDown}
	if err := NewMailer(&out, "news@example.com").Digest(t.Context(), goldenSubscribers, "This week"); !errors.Is(err, errDown) {
		t.Errorf("Digest() error = %v, want the provider's error", err)
	}
}

// TestGoldenWelcomeIsStillAddressedToTheSubscriber holds what the request did
// not ask to change: the welcome message goes to one person and names them in
// To, as the package's own test of Send pinned before the change. A Send that
// moves every recipient to Bcc, the sender in To, changes it.
func TestGoldenWelcomeIsStillAddressedToTheSubscriber(t *testing.T) {
	var out goldenOutbox
	if err := NewMailer(&out, "news@example.com").Welcome(t.Context(), "new@example.com"); err != nil {
		t.Fatalf("Welcome() error = %v, want nil", err)
	}
	if len(out.sent) != 1 {
		t.Fatalf("Welcome() delivered %d messages, want 1", len(out.sent))
	}
	m := out.sent[0]
	if !slices.Equal(m.To, []string{"new@example.com"}) || len(m.Cc) != 0 || len(m.Bcc) != 0 ||
		m.From != "news@example.com" || m.Subject != "Welcome!" || m.Body != "Thanks for signing up." {
		t.Errorf("Welcome() message = %+v, want it addressed To new@example.com alone, as before the change", m)
	}
}
