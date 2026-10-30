# Test Organization Reference

> Sources: source/google-go-styleguide/best-practices.md (Test Doubles, Test Structure); source/google-go-styleguide/decisions.md (Test packages); `testing` and `github.com/stretchr/testify` docs (Assertion Objects in Subtests)
> Authority: advisory
> Last verified: 2026-10-30

---

## Contents

- [Test Double Types](#test-double-types)
- [Test Double Naming Conventions](#test-double-naming-conventions)
- [Standalone Test Helper Packages](#standalone-test-helper-packages)
- [Test Packages](#test-packages)
- [Setup Scoping](#setup-scoping)
- [Assertion Objects in Subtests](#assertion-objects-in-subtests)

## Test Double Types

| Double | Purpose | State? | Verifies calls? |
|--------|---------|--------|-----------------|
| Stub   | Returns canned data | No | No |
| Fake   | Working but simplified implementation | Yes | No |
| Spy    | Records calls for later inspection | Yes | Yes |

**Prefer fakes over mocks.** Fakes are more readable and don't require mock
frameworks. Reserve spies for verifying side effects (e.g., an analytics event).

A fake keeps the real contract: the same sentinel errors (`ErrNotFound`
below, never `nil, nil`), the fields the real store fills (an ID, a
`CreatedAt`) even when this test reads none of them, and the same copy and
ordering semantics. A fake looser than the implementation it stands in for
passes tests that the integration fails.

```go
// Fake: Working in-memory implementation
type FakeUserStore struct {
    users map[string]User
}

func (f *FakeUserStore) GetUser(id string) (*User, error) {
    u, ok := f.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return &u, nil // a copy: an edit reaches the store only through SaveUser
}

func (f *FakeUserStore) SaveUser(u *User) error {
    f.users[u.ID] = *u
    return nil
}

// Spy: Records calls for later assertion
type SpyEmailSender struct{ Sent []string }

func (s *SpyEmailSender) Send(to, body string) error {
    s.Sent = append(s.Sent, to)
    return nil
}
```

---

## Test Double Naming Conventions

> **Advisory**: Follow consistent naming for test doubles (stubs, fakes, spies).

**Package naming**: Create a `*test` package alongside production code (e.g.,
`creditcardtest` for package `creditcard`, `fakeauthservice` for a standalone
fake service).

```go
// Good: In package creditcardtest

// Single double — use simple name
type Stub struct{}
func (Stub) Charge(*creditcard.Card, money.Money) error { return nil }

// Multiple behaviors — name by behavior
type AlwaysCharges struct{}
type AlwaysDeclines struct{}

// Multiple types — include type name
type StubService struct{}
type StubStoredValue struct{}
```

**Local variables**: Prefix test double variables with the double type for
clarity at the call site:

```go
// Good: Double type is immediately visible
spyCC := &creditcardtest.Spy{}
stubDB := &dbtest.Stub{Balance: 100}

// Bad: Ambiguous — is this real or a double?
cc := &creditcardtest.Spy{}
db := &dbtest.Stub{Balance: 100}
```

---

## Standalone Test Helper Packages

Create a standalone test helper package when multiple packages need the same
double, the helper has enough logic to warrant its own tests, or you want to
provide an acceptance test suite for interface implementers.

| Pattern | When to use | Example |
|---------|-------------|---------|
| `footest` | General test helpers for package `foo` | `creditcardtest`, `usertest` |
| `fakeX` | Standalone fake service package | `fakeauthservice`, `fakestorage` |

```go
package usertest

func NewFakeStore(t *testing.T, users ...*user.User) *FakeUserStore {
	t.Helper()
	store := &FakeUserStore{users: make(map[string]user.User)}
	for _, u := range users {
		store.users[u.ID] = *u
	}
	return store
}
```

Export constructors that accept `*testing.T` so they can call `t.Helper()` and
`t.Cleanup()`.

---

## Test Packages

If a black-box test (`package foo_test`) needs an unexported symbol, create
`export_test.go` in `package foo` (not `foo_test`) that exposes it. Use this
sparingly.

---

## Setup Scoping

Use `TestMain` only as a last resort (see [INTEGRATION.md](INTEGRATION.md)).

## Assertion Objects in Subtests

An `assert.New(t)` or `require.New(t)` bound to the parent `t` and used inside `t.Run` fails the
wrong test. Through `assert` the failure lands on the parent and the subtest
stays green; the tell in `-v` output is a parent FAIL over subtests that all
PASS. A `require` through it calls `FailNow` on the parent from the
subtest's goroutine; the tell is `subtest may have called FailNow on a
parent test`; under `t.Parallel` subtests the binary dies instead with
`panic: test executed panic(nil) or runtime.Goexit`. Build it from the
subtest's own `t`.
