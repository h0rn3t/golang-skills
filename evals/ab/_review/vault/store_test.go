package vault

import (
	"errors"
	"io/fs"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestPathRejectsEscape(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	_, err := s.path("t1", "../t2/secret.txt")
	if err == nil || err.Error() != "vault: bad file name" {
		t.Fatalf("path() error = %v, want %v", err, ErrBadName)
	}
}

func TestPutEnforcesQuota(t *testing.T) {
	s := NewStore(t.TempDir(), 10)
	if _, err := s.Put("t1", "a.txt", "text/plain", nil, strings.NewReader("123456")); err != nil {
		t.Fatalf("Put(a) error = %v", err)
	}
	if _, err := s.Put("t1", "b.txt", "text/plain", nil, strings.NewReader("123456")); !errors.Is(err, ErrQuota) {
		t.Fatalf("Put(b) error = %v, want ErrQuota", err)
	}
	if got := s.Used("t1"); got != 6 {
		t.Errorf("Used() = %d, want 6", got)
	}
}

func TestDeleteRemovesFile(t *testing.T) {
	s := NewStore(t.TempDir(), 1<<20)
	if _, err := s.Put("t1", "a.txt", "text/plain", nil, strings.NewReader("123456")); err != nil {
		t.Fatalf("Put() error = %v", err)
	}
	if err := s.Delete("t1", "a.txt"); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}
	if _, _, err := s.Open("t1", "a.txt"); !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("Open() after Delete() error = %v, want fs.ErrNotExist", err)
	}
}

func TestETagIsStable(t *testing.T) {
	m := Meta{Name: "a.txt", Size: 6, ModTime: time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC), Labels: map[string]string{"kind": "report"}}
	if a, b := m.ETag(), m.ETag(); a != b {
		t.Errorf("ETag() = %s then %s, want the same tag", a, b)
	}
	changed := m
	changed.Size = 7
	if changed.ETag() == m.ETag() {
		t.Error("ETag() unchanged after the size changed")
	}
}

func TestPutThenGetWithETag(t *testing.T) {
	srv := NewServer(NewStore(t.TempDir(), 1<<20), map[string]string{"tok-1": "t1"}, nil, slog.New(slog.DiscardHandler))
	do := func(method, target, body string, header http.Header) *httptest.ResponseRecorder {
		req := httptest.NewRequestWithContext(t.Context(), method, target, strings.NewReader(body))
		req.Header = header
		req.Header.Set("Authorization", "Bearer tok-1")
		rec := httptest.NewRecorder()
		srv.ServeHTTP(rec, req)
		return rec
	}
	if rec := do(http.MethodPut, "/files/t1/report.txt", "hello", http.Header{"Content-Type": {"text/plain"}}); rec.Code != http.StatusCreated {
		t.Fatalf("PUT = %d %s, want 201", rec.Code, rec.Body)
	}
	rec := do(http.MethodGet, "/files/t1/report.txt", "", http.Header{})
	if rec.Code != http.StatusOK || rec.Body.String() != "hello" {
		t.Fatalf("GET = %d %q, want 200 hello", rec.Code, rec.Body)
	}
	if rec := do(http.MethodGet, "/files/t1/report.txt", "", http.Header{"If-None-Match": {rec.Header().Get("ETag")}}); rec.Code != http.StatusNotModified {
		t.Errorf("GET with If-None-Match = %d, want 304", rec.Code)
	}
}
