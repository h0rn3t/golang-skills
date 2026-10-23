package signup

import (
	"context"
	"errors"
	"sync"
	"testing"
)

var errGoldenConn = errors.New("conn closed")

// goldenUsers is a users table with both UNIQUE constraints enforced on
// Insert. stale makes the two Find methods read a snapshot taken before any
// insert, as a concurrent registration that commits between the check and
// the insert looks to the checking call.
type goldenUsers struct {
	mu        sync.Mutex
	rows      []User
	stale     bool
	insertErr error
	inserts   int
}

func (g *goldenUsers) find(match func(User) bool) (User, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if !g.stale {
		for _, u := range g.rows {
			if match(u) {
				return u, nil
			}
		}
	}
	return User{}, ErrNoRows
}

func (g *goldenUsers) FindByEmail(_ context.Context, email string) (User, error) {
	return g.find(func(u User) bool { return u.Email == email })
}

func (g *goldenUsers) FindByUsername(_ context.Context, username string) (User, error) {
	return g.find(func(u User) bool { return u.Username == username })
}

func (g *goldenUsers) Insert(_ context.Context, email, username string) (int64, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	g.inserts++
	if g.insertErr != nil {
		return 0, g.insertErr
	}
	for _, u := range g.rows {
		if u.Email == email {
			return 0, &PgError{Code: "23505", ConstraintName: "users_email_key", Message: `duplicate key value violates unique constraint "users_email_key"`}
		}
		if u.Username == username {
			return 0, &PgError{Code: "23505", ConstraintName: "users_username_key", Message: `duplicate key value violates unique constraint "users_username_key"`}
		}
	}
	id := int64(len(g.rows) + 1)
	g.rows = append(g.rows, User{ID: id, Email: email, Username: username})
	return id, nil
}

func goldenTable() *goldenUsers {
	return &goldenUsers{rows: []User{{ID: 1, Email: "ada@example.com", Username: "ada"}}}
}

func TestRegisterCreatesTheAccount(t *testing.T) {
	g := goldenTable()
	got, err := Register(t.Context(), g, "grace@example.com", "grace")
	if err != nil {
		t.Fatalf("Register(grace) error = %v, want nil", err)
	}
	if want := (User{ID: 2, Email: "grace@example.com", Username: "grace"}); got != want {
		t.Errorf("Register(grace) = %+v, want %+v", got, want)
	}
}

func TestRegisterRejectsTakenValues(t *testing.T) {
	tests := []struct {
		name            string
		stale           bool
		email, username string
		want, notWant   error
	}{
		{"email", false, "ada@example.com", "ada2", ErrEmailTaken, ErrUsernameTaken},
		{"username", false, "ada2@example.com", "ada", ErrUsernameTaken, ErrEmailTaken},
		// The trap: a check that passes and an insert that the constraint
		// rejects, because another registration committed in between.
		{"email taken after the check", true, "ada@example.com", "ada2", ErrEmailTaken, ErrUsernameTaken},
		{"username taken after the check", true, "ada2@example.com", "ada", ErrUsernameTaken, ErrEmailTaken},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := goldenTable()
			g.stale = tt.stale
			_, err := Register(t.Context(), g, tt.email, tt.username)
			if !errors.Is(err, tt.want) || errors.Is(err, tt.notWant) {
				t.Errorf("Register(%q, %q) error = %v, want %v", tt.email, tt.username, err, tt.want)
			}
			if len(g.rows) != 1 {
				t.Errorf("Register(%q, %q) left %d rows, want 1", tt.email, tt.username, len(g.rows))
			}
		})
	}
}

func TestRegisterConcurrentSameEmail(t *testing.T) {
	g := &goldenUsers{stale: true}
	const n = 8
	errs := make([]error, n)
	var wg sync.WaitGroup
	for i := range n {
		wg.Go(func() {
			_, errs[i] = Register(t.Context(), g, "lin@example.com", "lin"+string(rune('a'+i)))
		})
	}
	wg.Wait()
	ok, taken := 0, 0
	for _, err := range errs {
		switch {
		case err == nil:
			ok++
		case errors.Is(err, ErrEmailTaken):
			taken++
		default:
			t.Errorf("concurrent Register(lin@example.com) error = %v, want nil or ErrEmailTaken", err)
		}
	}
	if ok != 1 || taken != n-1 {
		t.Errorf("concurrent Register(lin@example.com) x%d: %d succeeded, %d ErrEmailTaken, want 1 and %d", n, ok, taken, n-1)
	}
}

func TestRegisterOtherFailures(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"connection", errGoldenConn},
		{"not null", &PgError{Code: "23502", Message: `null value in column "email"`}},
		{"other unique constraint", &PgError{Code: "23505", ConstraintName: "users_pkey", Message: `duplicate key value violates unique constraint "users_pkey"`}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g := goldenTable()
			g.insertErr = tt.err
			_, err := Register(t.Context(), g, "grace@example.com", "grace")
			if err == nil || errors.Is(err, ErrEmailTaken) || errors.Is(err, ErrUsernameTaken) {
				t.Errorf("Register with Insert failing %v: error = %v, want a non-nil error that is neither sentinel", tt.err, err)
			}
		})
	}
}
