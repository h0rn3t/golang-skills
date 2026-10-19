package gatekeeper

import (
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newTestServer(t *testing.T) (http.Handler, *Store, *Sessions) {
	t.Helper()
	store := NewStore()
	sessions := NewSessions(func(context.Context, string) (bool, error) { return false, nil })
	return NewServer(store, sessions, slog.New(slog.DiscardHandler)).Routes(), store, sessions
}

func do(t *testing.T, h http.Handler, method, path, session, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(t.Context(), method, path, strings.NewReader(body))
	if session != "" {
		req.Header.Set("Cookie", "__Host-session="+session)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func TestDocuments(t *testing.T) {
	h, store, sessions := newTestServer(t)
	store.AddMember("o1", "ann")
	ann := sessions.Start("ann")
	bob := sessions.Start("bob")

	rec := do(t, h, http.MethodPost, "/orgs/o1/docs", ann, `{"org_id":"o1","title":"plan","body":"line one\nmore"}`)
	if rec.Code != http.StatusCreated {
		t.Fatalf("POST /orgs/o1/docs = %d, want %d", rec.Code, http.StatusCreated)
	}

	for _, tt := range []struct {
		name, session string
		want          int
	}{
		{name: "member", session: ann, want: http.StatusOK},
		{name: "no session", want: http.StatusUnauthorized},
		{name: "unknown session", session: "nope", want: http.StatusUnauthorized},
		{name: "non-member", session: bob, want: http.StatusForbidden},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := do(t, h, http.MethodGet, "/orgs/o1/docs/1", tt.session, "").Code; got != tt.want {
				t.Errorf("GET /orgs/o1/docs/1 as %s = %d, want %d", tt.name, got, tt.want)
			}
		})
	}
}

func TestInviteOnce(t *testing.T) {
	h, store, sessions := newTestServer(t)
	store.AddInvite(Invite{Code: "c1", OrgID: "o1", Expires: time.Now().Add(time.Hour)})
	ann := sessions.Start("ann")

	if got := do(t, h, http.MethodPost, "/invites/c1", ann, "").Code; got != http.StatusNoContent {
		t.Fatalf("first POST /invites/c1 = %d, want %d", got, http.StatusNoContent)
	}
	if !store.IsMember("o1", "ann") {
		t.Error("IsMember(o1, ann) = false after redeeming, want true")
	}
	if got := do(t, h, http.MethodPost, "/invites/c1", ann, "").Code; got != http.StatusNotFound {
		t.Errorf("second POST /invites/c1 = %d, want %d", got, http.StatusNotFound)
	}
}
