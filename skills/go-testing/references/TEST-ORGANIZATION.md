# Test Organization Reference

> Sources: source/google-go-styleguide/best-practices.md (Test Doubles, Test Structure); source/google-go-styleguide/decisions.md (Test packages)
> Authority: advisory
> Last verified: 2026-09-10

---

## Test Double Types

| Double | Purpose | State? | Verifies calls? |
|--------|---------|--------|-----------------|
| Stub   | Returns canned data | No | No |
| Fake   | Working but simplified implementation | Yes | No |
| Spy    | Records calls for later inspection | Yes | Yes |

**Prefer fakes over mocks.** Fakes are more readable and don't require mock
frameworks. Reserve spies for verifying side effects (e.g., an analytics event).

```go
// Fake: Working in-memory implementation
type FakeUserStore struct {
    users map[string]*User
}

func (f *FakeUserStore) GetUser(id string) (*User, error) {
    u, ok := f.users[id]
    if !ok {
        return nil, ErrNotFound
    }
    return u, nil
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
    store := &FakeUserStore{users: make(map[string]*user.User)}
    for _, u := range users {
        store.users[u.ID] = u
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
