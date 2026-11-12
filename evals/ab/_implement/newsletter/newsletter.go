// Package newsletter sends the service's email through the mail provider: the
// welcome message on signup and the weekly digest.
package newsletter

import "context"

// Message is one email as the provider's API takes it. The provider bills per
// message.
type Message struct {
	From        string
	To, Cc, Bcc []string
	Subject     string
	Body        string
}

// Provider is the mail provider's client.
type Provider interface {
	Deliver(ctx context.Context, m Message) error
}

// Mailer sends the service's email from one sender address.
type Mailer struct {
	provider Provider
	from     string
}

// NewMailer returns a Mailer that sends from the address from.
func NewMailer(p Provider, from string) *Mailer {
	return &Mailer{provider: p, from: from}
}

// Send emails one recipient.
func (m *Mailer) Send(ctx context.Context, to, subject, body string) error {
	return m.provider.Deliver(ctx, Message{From: m.from, To: []string{to}, Subject: subject, Body: body})
}

// Welcome greets a new subscriber.
func (m *Mailer) Welcome(ctx context.Context, address string) error {
	return m.Send(ctx, address, "Welcome!", "Thanks for signing up.")
}

// Digest emails the weekly digest to every subscriber. It sends nothing when
// there are none.
func (m *Mailer) Digest(ctx context.Context, subscribers []string, text string) error {
	for _, address := range subscribers {
		if err := m.Send(ctx, address, "Weekly digest", text); err != nil {
			return err
		}
	}
	return nil
}
