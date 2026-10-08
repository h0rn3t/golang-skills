package keyring

import (
	"testing"
	"time"
)

var now = time.Date(2026, 10, 8, 12, 0, 0, 0, time.UTC)

func TestIssuedKeyVerifies(t *testing.T) {
	r := New()
	secret := r.Issue("ci", now.Add(time.Hour))
	if !r.Verify("ci", secret, now) {
		t.Errorf("Verify(%q, issued key) = false, want true", "ci")
	}
}

func TestRevokedOrExpiredKeyFails(t *testing.T) {
	r := New()
	revoked := r.Issue("old", now.Add(time.Hour))
	r.Revoke("old")
	if r.Verify("old", revoked, now) {
		t.Errorf("Verify(%q, revoked key) = true, want false", "old")
	}
	expired := r.Issue("stale", now.Add(-time.Minute))
	if r.Verify("stale", expired, now) {
		t.Errorf("Verify(%q, expired key) = true, want false", "stale")
	}
}
