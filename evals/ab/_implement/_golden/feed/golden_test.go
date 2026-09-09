package feed

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

// goldenAt is the reference instant every case formats.
var goldenAt = time.Date(2026, 9, 7, 8, 30, 0, 0, time.UTC)

// goldenMembers decodes the document to its raw JSON members, so a case can ask
// what a member actually rendered as without depending on the order the
// implementation happens to emit them in.
func goldenMembers(t *testing.T, account string, events []Event) map[string]json.RawMessage {
	t.Helper()
	data, err := Render(account, events)
	if err != nil {
		t.Fatalf("Render(%q) error = %v, want nil", account, err)
	}
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		t.Fatalf("json.Unmarshal(Render(%q)) error = %v, document was %s", account, err, data)
	}
	return raw
}

// TestRenderEmptyKeepsEveryMemberTyped is the trap. A nil slice marshals to
// null and so does a nil map, and the documented client walks these members as
// an array and an object in every response.
func TestRenderEmptyKeepsEveryMemberTyped(t *testing.T) {
	cases := []struct {
		name   string
		events []Event
	}{
		{name: "no events"},
		// The other route to a null: every event is filtered out, so the
		// members are reached but stay empty.
		{name: "every event dropped", events: []Event{{Kind: "login", At: goldenAt}}},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			raw := goldenMembers(t, "acct-1", tt.events)

			want := map[string]string{"events": "[]", "kinds": "[]", "counts": "{}", "account": `"acct-1"`}
			for member, wantJSON := range want {
				if got := string(raw[member]); got != wantJSON {
					t.Errorf("member %q rendered as %s, want %s", member, got, wantJSON)
				}
			}
		})
	}
}

func TestRenderDocument(t *testing.T) {
	events := []Event{
		{ID: "e3", Kind: "logout", At: goldenAt},
		{ID: "", Kind: "dropped", At: goldenAt},
		{ID: "e1", Kind: "login", At: goldenAt.Add(time.Hour)},
		{ID: "e2", Kind: "login", At: goldenAt},
	}

	raw := goldenMembers(t, "acct-2", events)

	wantMembers := map[string]string{
		"account": `"acct-2"`,
		"events": `[{"id":"e3","kind":"logout","at":"2026-09-07T08:30:00Z"},` +
			`{"id":"e1","kind":"login","at":"2026-09-07T09:30:00Z"},` +
			`{"id":"e2","kind":"login","at":"2026-09-07T08:30:00Z"}]`,
		"kinds":  `["login","logout"]`,
		"counts": `{"login":2,"logout":1}`,
	}
	for member, wantJSON := range wantMembers {
		got := string(raw[member])
		if member == "events" || member == "counts" {
			// Order inside events is specified; key order inside counts is not,
			// so compare those two through a decode rather than byte for byte.
			goldenAssertJSONEqual(t, member, got, wantJSON)
			continue
		}
		if got != wantJSON {
			t.Errorf("member %q rendered as %s, want %s", member, got, wantJSON)
		}
	}
}

func goldenAssertJSONEqual(t *testing.T, member, got, want string) {
	t.Helper()
	var gotValue, wantValue any
	if err := json.Unmarshal([]byte(got), &gotValue); err != nil {
		t.Errorf("member %q is not valid JSON: %s", member, got)
		return
	}
	if err := json.Unmarshal([]byte(want), &wantValue); err != nil {
		t.Fatalf("the test's own expectation for %q is not valid JSON: %s", member, want)
	}
	gotNorm, _ := json.Marshal(gotValue)
	wantNorm, _ := json.Marshal(wantValue)
	if string(gotNorm) != string(wantNorm) {
		t.Errorf("member %q rendered as %s, want %s", member, gotNorm, wantNorm)
	}
}

func TestRenderRejectsEmptyAccount(t *testing.T) {
	_, err := Render("", nil)
	if err == nil {
		t.Fatal("Render(empty account) error = nil, want non-nil")
	}
	if !strings.Contains(err.Error(), "account") {
		t.Errorf("Render(empty account) error = %q, want the field named in the message", err)
	}
}
