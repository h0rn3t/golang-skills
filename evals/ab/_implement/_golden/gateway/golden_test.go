package gateway

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func goldenAccounts() []Account {
	return []Account{
		{ID: "a-3", Name: "Fees", Cents: -400, Active: false},
		{ID: "a-1", Name: "Cash", Cents: 2675, Active: true},
		{ID: "a-2", Name: "Card", Cents: 900, Active: true},
	}
}

// serve drives one request through the server's handler. Only NewServer is
// part of the fixture's API, so everything behind it is the implementation's
// own business.
func serve(t *testing.T, method, target string) *httptest.ResponseRecorder {
	t.Helper()
	srv := NewServer(":8080", goldenAccounts())
	if srv == nil {
		t.Fatal("NewServer returned nil, want a server")
	}
	if srv.Handler == nil {
		t.Fatal("NewServer().Handler = nil, want a handler")
	}
	rec := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rec, httptest.NewRequest(method, target, nil))
	return rec
}

// TestNewServerBoundsSlowClients is the trap. A zero timeout is no timeout, so
// an edge server assembled as &http.Server{Addr: addr, Handler: h} holds a
// stalled or idle connection until the process runs out of them.
func TestNewServerBoundsSlowClients(t *testing.T) {
	srv := NewServer(":8080", goldenAccounts())
	if srv == nil {
		t.Fatal("NewServer returned nil, want a server")
	}

	timeouts := []struct {
		field string
		zero  bool
	}{
		{field: "ReadHeaderTimeout", zero: srv.ReadHeaderTimeout == 0},
		{field: "ReadTimeout", zero: srv.ReadTimeout == 0},
		{field: "WriteTimeout", zero: srv.WriteTimeout == 0},
		{field: "IdleTimeout", zero: srv.IdleTimeout == 0},
	}
	for _, tt := range timeouts {
		if tt.zero {
			t.Errorf("NewServer().%s = 0, want a bound on the connection", tt.field)
		}
	}
	if srv.Addr != ":8080" {
		t.Errorf("NewServer().Addr = %q, want %q", srv.Addr, ":8080")
	}
}

func TestHealth(t *testing.T) {
	rec := serve(t, http.MethodGet, "/healthz")

	if rec.Code != http.StatusOK {
		t.Errorf("GET /healthz = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Body.String(); got != "ok" {
		t.Errorf("GET /healthz body = %q, want %q", got, "ok")
	}
}

func TestListOrdersByID(t *testing.T) {
	rec := serve(t, http.MethodGet, "/accounts")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /accounts = %d, want %d", rec.Code, http.StatusOK)
	}

	var got []Account
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(GET /accounts) error = %v, body was %s", err, rec.Body)
	}
	want := []Account{
		{ID: "a-1", Name: "Cash", Cents: 2675, Active: true},
		{ID: "a-2", Name: "Card", Cents: 900, Active: true},
		{ID: "a-3", Name: "Fees", Cents: -400, Active: false},
	}
	if len(got) != len(want) {
		t.Fatalf("GET /accounts returned %d accounts, want %d: %s", len(got), len(want), rec.Body)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("GET /accounts [%d] = %+v, want %+v", i, got[i], want[i])
		}
	}
}

func TestListFiltersByActive(t *testing.T) {
	tests := []struct {
		target string
		wantID []string
	}{
		{target: "/accounts?active=true", wantID: []string{"a-1", "a-2"}},
		{target: "/accounts?active=false", wantID: []string{"a-3"}},
	}

	for _, tt := range tests {
		t.Run(tt.target, func(t *testing.T) {
			rec := serve(t, http.MethodGet, tt.target)
			if rec.Code != http.StatusOK {
				t.Fatalf("GET %s = %d, want %d", tt.target, rec.Code, http.StatusOK)
			}

			var got []Account
			if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
				t.Fatalf("json.Unmarshal(GET %s) error = %v, body was %s", tt.target, err, rec.Body)
			}
			if len(got) != len(tt.wantID) {
				t.Fatalf("GET %s returned %d accounts, want %d: %s", tt.target, len(got), len(tt.wantID), rec.Body)
			}
			for i, id := range tt.wantID {
				if got[i].ID != id {
					t.Errorf("GET %s [%d].ID = %q, want %q", tt.target, i, got[i].ID, id)
				}
			}
		})
	}
}

func TestListRejectsUnknownActiveValue(t *testing.T) {
	rec := serve(t, http.MethodGet, "/accounts?active=maybe")

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("GET /accounts?active=maybe = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if !strings.Contains(rec.Body.String(), "active") {
		t.Errorf("GET /accounts?active=maybe body = %q, want the parameter named", rec.Body)
	}
}

func TestGetOne(t *testing.T) {
	rec := serve(t, http.MethodGet, "/accounts/a-2")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /accounts/a-2 = %d, want %d", rec.Code, http.StatusOK)
	}

	var got Account
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("json.Unmarshal(GET /accounts/a-2) error = %v, body was %s", err, rec.Body)
	}
	if want := (Account{ID: "a-2", Name: "Card", Cents: 900, Active: true}); got != want {
		t.Errorf("GET /accounts/a-2 = %+v, want %+v", got, want)
	}
}

func TestStatusCodes(t *testing.T) {
	tests := []struct {
		name     string
		method   string
		target   string
		wantCode int
	}{
		{name: "unknown account", method: http.MethodGet, target: "/accounts/nope", wantCode: http.StatusNotFound},
		{name: "unknown path", method: http.MethodGet, target: "/nope", wantCode: http.StatusNotFound},
		{name: "post to list", method: http.MethodPost, target: "/accounts", wantCode: http.StatusMethodNotAllowed},
		{name: "delete one", method: http.MethodDelete, target: "/accounts/a-1", wantCode: http.StatusMethodNotAllowed},
		{name: "post to health", method: http.MethodPost, target: "/healthz", wantCode: http.StatusMethodNotAllowed},
		{name: "head to health", method: http.MethodHead, target: "/healthz", wantCode: http.StatusMethodNotAllowed},
		{name: "head to list", method: http.MethodHead, target: "/accounts", wantCode: http.StatusMethodNotAllowed},
		{name: "head to account", method: http.MethodHead, target: "/accounts/a-1", wantCode: http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := serve(t, tt.method, tt.target)
			if rec.Code != tt.wantCode {
				t.Errorf("%s %s = %d, want %d", tt.method, tt.target, rec.Code, tt.wantCode)
			}
		})
	}
}

func TestEmptyListIsJSONArray(t *testing.T) {
	for _, tt := range []struct {
		name     string
		accounts []Account
		target   string
	}{
		{name: "nil", target: "/accounts"},
		{name: "empty", accounts: []Account{}, target: "/accounts"},
		{name: "filtered", accounts: []Account{{ID: "a-1", Active: true}}, target: "/accounts?active=false"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer(":0", tt.accounts)
			rec := httptest.NewRecorder()
			srv.Handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, tt.target, nil))
			if got := strings.TrimSpace(rec.Body.String()); rec.Code != http.StatusOK || got != "[]" {
				t.Errorf("GET %s = %d %q, want 200 []", tt.target, rec.Code, got)
			}
		})
	}
}
