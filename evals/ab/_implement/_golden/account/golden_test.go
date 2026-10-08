package account

import (
	"context"
	"encoding/json/v2"
	"errors"
	"maps"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"sync"
	"testing"
)

var errGoldenDisk = errors.New("disk full at /var/lib/accounts")

// goldenStore keeps accounts in memory and counts writes; failing makes every
// call return errGoldenDisk.
type goldenStore struct {
	mu       sync.Mutex
	accounts map[string]Account
	puts     int
	failing  bool
}

func (s *goldenStore) Get(_ context.Context, id string) (Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failing {
		return Account{}, errGoldenDisk
	}
	a, ok := s.accounts[id]
	if !ok {
		return Account{}, ErrNotFound
	}
	return a, nil
}

func (s *goldenStore) Put(_ context.Context, a Account) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.failing {
		return errGoldenDisk
	}
	s.puts++
	s.accounts[a.ID] = a
	return nil
}

var goldenAda = Account{ID: "7", Email: "ada@example.com", DisplayName: "Ada", Phone: "+44 20 7946 0000", PasswordHash: "$argon2id$v=19$m=65536", Role: "member"}

// goldenPatch mounts the handler on the pattern it serves, so a handler that
// reads r.PathValue and one that routes the path itself both receive the id.
func goldenPatch(t *testing.T, s *goldenStore, id, body string) *httptest.ResponseRecorder {
	t.Helper()
	mux := http.NewServeMux()
	mux.Handle("PATCH /accounts/{id}", NewHandler(s))
	w := httptest.NewRecorder()
	mux.ServeHTTP(w, httptest.NewRequest(http.MethodPatch, "/accounts/"+id, strings.NewReader(body)))
	return w
}

func goldenNewStore() *goldenStore {
	return &goldenStore{accounts: map[string]Account{"7": goldenAda}}
}

// TestPatchKeepsSettingsTheBodyLeavesOut is the first trap: decoding into a
// fresh Account and saving it blanks every setting the user did not touch.
func TestPatchKeepsSettingsTheBodyLeavesOut(t *testing.T) {
	s := goldenNewStore()
	if w := goldenPatch(t, s, "7", `{"display_name":"Ada L."}`); w.Code != http.StatusOK {
		t.Fatalf("PATCH display_name = %d %s, want 200", w.Code, w.Body)
	}
	want := goldenAda
	want.DisplayName = "Ada L."
	if got := s.accounts["7"]; got != want {
		t.Errorf("stored after PATCH display_name = %+v, want %+v", got, want)
	}
}

func TestPatchNullRemovesPhone(t *testing.T) {
	s := goldenNewStore()
	if w := goldenPatch(t, s, "7", `{"phone":null}`); w.Code != http.StatusOK {
		t.Fatalf("PATCH phone null = %d %s, want 200", w.Code, w.Body)
	}
	if got := s.accounts["7"].Phone; got != "" {
		t.Errorf("phone after PATCH phone null = %q, want removed", got)
	}
	if got := s.accounts["7"].Email; got != goldenAda.Email {
		t.Errorf("email after PATCH phone null = %q, want %q", got, goldenAda.Email)
	}
}

// TestPatchRejectsAndChangesNothing is the second trap: decoding into the
// stored Account lets the body set password_hash and role.
func TestPatchRejectsAndChangesNothing(t *testing.T) {
	for _, body := range []string{
		`{"role":"admin"}`,
		`{"password_hash":"x"}`,
		`{"display_name":"Ada","role":"admin"}`,
		`{"nickname":"A"}`,
		`{"email":null}`,
		`{"display_name":""}`,
		`null`,
		`[]`,
		`{"display_name":"Ada"} {"display_name":"Eve"}`,
		`{"display_name":"` + strings.Repeat("a", 70<<10) + `"}`,
	} {
		s := goldenNewStore()
		w := goldenPatch(t, s, "7", body)
		if w.Code != http.StatusBadRequest || s.puts != 0 || s.accounts["7"] != goldenAda {
			t.Errorf("PATCH %.40s = %d with %d writes, stored %+v; want 400 and nothing changed", body, w.Code, s.puts, s.accounts["7"])
		}
	}
}

// TestPatchAnswersThePublicView is the third trap: encoding the stored
// Account sends the password hash and role back to the browser.
func TestPatchAnswersThePublicView(t *testing.T) {
	s := goldenNewStore()
	w := goldenPatch(t, s, "7", `{"phone":null}`)
	var got map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &got); err != nil {
		t.Fatalf("PATCH response %q is not a JSON object: %v", w.Body, err)
	}
	if keys := slices.Sorted(maps.Keys(got)); !slices.Equal(keys, []string{"display_name", "email", "id", "phone"}) {
		t.Errorf("PATCH response members = %v, want exactly display_name, email, id, phone", keys)
	}
	if got["id"] != "7" || got["email"] != goldenAda.Email || got["phone"] != "" {
		t.Errorf("PATCH response = %v, want id 7, the stored email, and phone \"\"", got)
	}
}

func TestPatchUnknownAccount(t *testing.T) {
	if w := goldenPatch(t, goldenNewStore(), "8", `{"display_name":"Eve"}`); w.Code != http.StatusNotFound {
		t.Errorf("PATCH /accounts/8 = %d, want 404", w.Code)
	}
}

func TestPatchStoreFailure(t *testing.T) {
	s := goldenNewStore()
	s.failing = true
	w := goldenPatch(t, s, "7", `{"display_name":"Eve"}`)
	if w.Code != http.StatusInternalServerError || strings.Contains(w.Body.String(), "disk full") {
		t.Errorf("PATCH with the store failing = %d %q, want 500 without the error text", w.Code, w.Body)
	}
}
