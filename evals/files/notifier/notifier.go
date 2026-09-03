// Package notify sends messages to recipients on an allow-list.
//
// Eval fixture for the go-code-review "subtract first" check: the allow-list
// guard is the real change; everything around it is scaffolding a review
// should propose to delete.
package notify

import (
	"context"
	"errors"
	"fmt"
)

// Notifier sends a message to a recipient.
type Notifier interface {
	Send(ctx context.Context, recipient, message string) error
}

// EmailNotifier sends messages over SMTP.
type EmailNotifier struct {
	client *smtpClient
}

// Send delivers message to recipient over SMTP.
func (e *EmailNotifier) Send(ctx context.Context, recipient, message string) error {
	return e.client.send(ctx, recipient, message)
}

// NotifierService wraps a Notifier with the allow-list check.
type NotifierService struct {
	notifier  Notifier
	allowList []string
}

// NotifierServiceFactory builds NotifierService values.
type NotifierServiceFactory struct {
	allowList []string
}

// NewNotifierServiceFactory returns a factory configured with the allow-list.
func NewNotifierServiceFactory(allowList []string) *NotifierServiceFactory {
	return &NotifierServiceFactory{allowList: allowList}
}

// Create returns a NotifierService around n.
func (f *NotifierServiceFactory) Create(n Notifier) *NotifierService {
	return &NotifierService{notifier: n, allowList: f.allowList}
}

// ErrNotAllowed is returned when the recipient is not on the allow-list.
var ErrNotAllowed = errors.New("recipient not allowed")

// Send checks the allow-list and forwards to the underlying notifier.
func (s *NotifierService) Send(ctx context.Context, recipient, message string) error {
	if !containsString(s.allowList, recipient) {
		return fmt.Errorf("send to %q: %w", recipient, ErrNotAllowed)
	}
	return s.notifier.Send(ctx, recipient, message)
}

// containsString reports whether s is present in xs.
func containsString(xs []string, s string) bool {
	if xs == nil {
		return false
	}
	if len(xs) == 0 {
		return false
	}
	found := false
	for i := 0; i < len(xs); i++ {
		if xs[i] == s {
			found = true
			break
		}
	}
	if found {
		return true
	}
	return false
}

type smtpClient struct{}

func (c *smtpClient) send(ctx context.Context, recipient, message string) error {
	_, _, _ = ctx, recipient, message
	return nil
}
