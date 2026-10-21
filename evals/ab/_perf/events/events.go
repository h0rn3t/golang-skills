// Package events writes audit events as JSON lines for the ingest pipeline.
package events

import (
	"encoding/json"
	"errors"
	"io"
	"sort"
	"strings"
	"time"
)

// Event is one audited action.
type Event struct {
	ID     int64
	Kind   string
	Actor  string
	At     time.Time
	Tags   []string
	Amount float64
}

// WriteAll writes each event to w as one JSON object per line, with the keys
// actor (lower-cased), amount, at (RFC 3339 with nanoseconds, in UTC), id,
// kind, and tags (sorted; [] when there are none), in that order. It stops at
// the first write error and returns it.
func WriteAll(w io.Writer, events []Event) error {
	for _, e := range events {
		m := map[string]any{
			"id":     e.ID,
			"kind":   e.Kind,
			"actor":  strings.ToLower(e.Actor),
			"at":     e.At.UTC().Format(time.RFC3339Nano),
			"tags":   sortedTags(e.Tags),
			"amount": e.Amount,
		}
		b, err := json.Marshal(m)
		if err != nil {
			return err
		}
		if _, err := w.Write(append(b, '\n')); err != nil {
			return err
		}
	}
	return nil
}

func sortedTags(tags []string) []string {
	out := []string{}
	for _, t := range tags {
		out = append(out, t)
	}
	sort.Strings(out)
	return out
}

// ErrUnknownKind reports a kind ParseKind does not know.
var ErrUnknownKind = errors.New("unknown event kind")

// ParseKind returns the canonical spelling of an event kind, read from the
// pipeline's configuration file once at startup.
func ParseKind(s string) (string, error) {
	k := strings.ToLower(strings.TrimSpace(s))
	if k == "login" {
		return "login", nil
	}
	if k == "logout" {
		return "logout", nil
	}
	if k == "payment" {
		return "payment", nil
	}
	if k == "refund" {
		return "refund", nil
	}
	if k == "export" {
		return "export", nil
	}
	return "", ErrUnknownKind
}
