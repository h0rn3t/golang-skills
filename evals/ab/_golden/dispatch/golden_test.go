package dispatch

import (
	"context"
	"errors"
	"strings"
	"testing"
)

type fakeStore struct {
	puts    map[string]string
	deletes []string
	err     error
}

func (f *fakeStore) Put(ctx context.Context, key, value string) error {
	if f.err != nil {
		return f.err
	}
	if f.puts == nil {
		f.puts = map[string]string{}
	}
	f.puts[key] = value
	return nil
}

func (f *fakeStore) Delete(ctx context.Context, key string) error {
	if f.err != nil {
		return f.err
	}
	f.deletes = append(f.deletes, key)
	return nil
}

// TestGoldenDispatch pins the observable behavior the refactor must preserve:
// key format, which store call each kind makes, and the error text.
func TestGoldenDispatch(t *testing.T) {
	ctx := context.Background()

	s := &fakeStore{}
	if err := Dispatch(ctx, s, Event{Kind: "created", ID: "AB-1", Payload: "p"}); err != nil {
		t.Fatalf("created: %v", err)
	}
	if got := s.puts["event:created:ab-1"]; got != "p" {
		t.Fatalf("created key/value = %q, want %q at event:created:ab-1", got, "p")
	}

	s = &fakeStore{}
	if err := Dispatch(ctx, s, Event{Kind: "updated", ID: "AB-2", Payload: "q"}); err != nil {
		t.Fatalf("updated: %v", err)
	}
	if got := s.puts["event:updated:ab-2"]; got != "q" {
		t.Fatalf("updated key/value = %q, want %q at event:updated:ab-2", got, "q")
	}

	s = &fakeStore{}
	if err := Dispatch(ctx, s, Event{Kind: "deleted", ID: "AB-3"}); err != nil {
		t.Fatalf("deleted: %v", err)
	}
	if len(s.deletes) != 1 || s.deletes[0] != "event:deleted:ab-3" {
		t.Fatalf("deleted calls = %v, want [event:deleted:ab-3]", s.deletes)
	}
	if len(s.puts) != 0 {
		t.Fatalf("deleted must not Put, got %v", s.puts)
	}

	err := Dispatch(ctx, &fakeStore{}, Event{Kind: "created"})
	if err == nil || err.Error() != "dispatch created: empty id" {
		t.Fatalf("empty id error = %v, want %q", err, "dispatch created: empty id")
	}

	err = Dispatch(ctx, &fakeStore{}, Event{Kind: "nope", ID: "x"})
	if err == nil || err.Error() != `dispatch: unknown kind "nope"` {
		t.Fatalf("unknown kind error = %v, want %q", err, `dispatch: unknown kind "nope"`)
	}

	sentinel := errors.New("boom")
	err = Dispatch(ctx, &fakeStore{err: sentinel}, Event{Kind: "created", ID: "AB-4"})
	if !errors.Is(err, sentinel) {
		t.Fatalf("store error must be wrapped with %%w, got %v", err)
	}
	if !strings.HasPrefix(err.Error(), "dispatch created AB-4: ") {
		t.Fatalf("wrap prefix = %q, want %q", err.Error(), "dispatch created AB-4: ")
	}

	s = &fakeStore{}
	if err := DispatchAll(ctx, s, []Event{{Kind: "created", ID: "a", Payload: "1"}, {Kind: "created", ID: "b", Payload: "2"}}); err != nil {
		t.Fatalf("DispatchAll: %v", err)
	}
	if len(s.puts) != 2 {
		t.Fatalf("DispatchAll puts = %d, want 2", len(s.puts))
	}
	if err := DispatchAll(ctx, &fakeStore{}, []Event{{Kind: "created", ID: "a"}, {Kind: "bad"}}); err == nil {
		t.Fatal("DispatchAll must stop at the first failure")
	}
}
