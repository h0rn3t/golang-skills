package feed

import (
	"encoding/json"
	"testing"
	"time"
)

// goldenAt is the reference instant every case formats.
var goldenAt = time.Date(2026, 9, 7, 8, 30, 0, 0, time.UTC)

// members decodes a Document to its raw JSON members, so a case can ask what a
// member actually rendered as without depending on field order.
func members(t *testing.T, doc Document, label string) map[string]json.RawMessage {
	t.Helper()
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("json.Marshal(%s) error = %v, want nil", label, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal(%s) error = %v, want nil", label, err)
	}
	return raw
}

// TestBuildEmptyRendersArrays is the trap. A nil slice marshals to null, and
// the documented client walks the members as arrays, so an account with no
// activity has to render as [] rather than null.
func TestBuildEmptyRendersArrays(t *testing.T) {
	cases := []struct {
		name   string
		events []Event
	}{
		{name: "no events"},
		// The other route to a null: every event is filtered out, so the
		// members are built but stay empty.
		{name: "every event dropped", events: []Event{{Kind: "login", At: goldenAt}}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			doc := Build("acct-1", tt.events)
			if doc.Events == nil {
				t.Error("Build().Events is nil, want an empty non-nil slice so JSON renders []")
			}
			if doc.Kinds == nil {
				t.Error("Build().Kinds is nil, want an empty non-nil slice so JSON renders []")
			}

			raw := members(t, doc, "Build "+tt.name)
			for _, member := range []string{"events", "kinds"} {
				if got := string(raw[member]); got != "[]" {
					t.Errorf("marshalled %q = %s, want []", member, got)
				}
			}
			if got := string(raw["account"]); got != `"acct-1"` {
				t.Errorf("marshalled %q = %s, want %q", "account", got, `"acct-1"`)
			}
		})
	}
}

func TestBuildRendersEventsInOrderWithSortedKinds(t *testing.T) {
	events := []Event{
		{ID: "e3", Kind: "logout", At: goldenAt},
		{ID: "", Kind: "dropped", At: goldenAt},
		{ID: "e1", Kind: "login", At: goldenAt.Add(time.Hour)},
		{ID: "e2", Kind: "login", At: goldenAt},
	}

	doc := Build("acct-2", events)

	wantEntries := []Entry{
		{ID: "e3", Kind: "logout", At: "2026-09-07T08:30:00Z"},
		{ID: "e1", Kind: "login", At: "2026-09-07T09:30:00Z"},
		{ID: "e2", Kind: "login", At: "2026-09-07T08:30:00Z"},
	}
	if len(doc.Events) != len(wantEntries) {
		t.Fatalf("Build().Events has length %d, want %d: %+v", len(doc.Events), len(wantEntries), doc.Events)
	}
	for i := range wantEntries {
		if doc.Events[i] != wantEntries[i] {
			t.Errorf("Build().Events[%d] = %+v, want %+v", i, doc.Events[i], wantEntries[i])
		}
	}

	wantKinds := []string{"login", "logout"}
	if len(doc.Kinds) != len(wantKinds) {
		t.Fatalf("Build().Kinds = %v, want %v", doc.Kinds, wantKinds)
	}
	for i := range wantKinds {
		if doc.Kinds[i] != wantKinds[i] {
			t.Errorf("Build().Kinds = %v, want %v", doc.Kinds, wantKinds)
			break
		}
	}
	if doc.Account != "acct-2" {
		t.Errorf("Build().Account = %q, want %q", doc.Account, "acct-2")
	}
}
