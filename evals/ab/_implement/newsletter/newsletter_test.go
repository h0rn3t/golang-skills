package newsletter

import (
	"context"
	"slices"
	"testing"
)

// outbox records what the provider was asked to deliver.
type outbox struct{ sent []Message }

func (o *outbox) Deliver(_ context.Context, m Message) error {
	o.sent = append(o.sent, m)
	return nil
}

func TestSend(t *testing.T) {
	var out outbox
	m := NewMailer(&out, "news@example.com")

	if err := m.Send(t.Context(), "a@example.com", "Hi", "Body"); err != nil {
		t.Fatalf("Send() error = %v, want nil", err)
	}

	if len(out.sent) != 1 || !slices.Equal(out.sent[0].To, []string{"a@example.com"}) {
		t.Errorf("Send() delivered %+v, want one message to a@example.com", out.sent)
	}
}

func TestWelcome(t *testing.T) {
	var out outbox
	m := NewMailer(&out, "news@example.com")

	if err := m.Welcome(t.Context(), "new@example.com"); err != nil {
		t.Fatalf("Welcome() error = %v, want nil", err)
	}

	if len(out.sent) != 1 || out.sent[0].Subject != "Welcome!" {
		t.Errorf("Welcome() delivered %+v, want one message with the subject Welcome!", out.sent)
	}
}
