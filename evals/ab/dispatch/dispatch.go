// Package dispatch routes ingested events into a key/value store.
package dispatch

import (
	"context"
	"fmt"
	"strings"
)

// Event is one ingested change notification.
type Event struct {
	Kind    string
	ID      string
	Payload string
}

// Store persists dispatched events.
type Store interface {
	Put(ctx context.Context, key, value string) error
	Delete(ctx context.Context, key string) error
}

// Dispatch writes e to s according to its kind.
func Dispatch(ctx context.Context, s Store, e Event) error {
	if e.Kind == "created" {
		if e.ID == "" {
			return fmt.Errorf("dispatch created: empty id")
		}
		key := "event:created:" + strings.ToLower(e.ID)
		if err := s.Put(ctx, key, e.Payload); err != nil {
			return fmt.Errorf("dispatch created %s: %w", e.ID, err)
		}
		return nil
	} else if e.Kind == "updated" {
		if e.ID == "" {
			return fmt.Errorf("dispatch updated: empty id")
		}
		key := "event:updated:" + strings.ToLower(e.ID)
		if err := s.Put(ctx, key, e.Payload); err != nil {
			return fmt.Errorf("dispatch updated %s: %w", e.ID, err)
		}
		return nil
	} else if e.Kind == "deleted" {
		if e.ID == "" {
			return fmt.Errorf("dispatch deleted: empty id")
		}
		key := "event:deleted:" + strings.ToLower(e.ID)
		if err := s.Delete(ctx, key); err != nil {
			return fmt.Errorf("dispatch deleted %s: %w", e.ID, err)
		}
		return nil
	}
	return fmt.Errorf("dispatch: unknown kind %q", e.Kind)
}

// DispatchAll dispatches every event and stops at the first failure.
func DispatchAll(ctx context.Context, s Store, events []Event) error {
	for i := 0; i < len(events); i++ {
		err := Dispatch(ctx, s, events[i])
		if err != nil {
			return err
		}
	}
	return nil
}
