# Go Testing: Integration and Advanced Patterns

> Sources: source/google-go-styleguide/best-practices.md (Test Structure, Test Doubles); https://pkg.go.dev/testing#hdr-Main
> Authority: advisory
> Last verified: 2026-10-01

TestMain, acceptance testing, real transports, and a real database in tests.

---

## TestMain

> **Source**: Google Go Style Guide (best-practices)

Use `func TestMain(m *testing.M)` when **all tests in the package** require
common setup that needs teardown (e.g., a shared database). This should **not be
your first choice**---prefer scoped test helpers or `t.Cleanup` when possible.

```go
var db *sql.DB

func TestInsert(t *testing.T) { /* uses db */ }
func TestSelect(t *testing.T) { /* uses db */ }

func TestMain(m *testing.M) {
    d, err := setupDatabase(context.Background())
    if err != nil {
        log.Fatal(err)
    }
    defer func() {
        if err := d.Close(); err != nil {
            log.Fatalf("close database: %v", err)
        }
    }()
    db = d
    m.Run()
}
```

Key points:
- `TestMain` returns rather than calling `os.Exit`: the test wrapper passes
  the result of `m.Run` to `os.Exit` itself (`go doc testing`), after the
  deferred `Close` has run. A `runMain` helper returning an exit code is the
  workaround for toolchains that predate this, not a pattern to copy
- Write failure messages to stderr via `log.Fatal`
- Ensure individual test cases remain hermetic---reset any global state they modify

---

## Acceptance Testing

> **Source**: Google Go Style Guide (best-practices)

Acceptance testing validates that an implementation upholds a contract, treating
it as a black box. This pattern is useful when users implement your interfaces
and you want to provide a reusable validation suite.

### Structure

1. Create a test helper package (e.g., `chesstest` for package `chess`)
2. Export a validation function that accepts the implementation under test:

```go
// Package chesstest provides acceptance tests for chess.Player implementations.
package chesstest

// ExercisePlayer tests a Player implementation in a single turn.
// Returns nil if the player makes a correct move, or an error describing
// the violation.
func ExercisePlayer(b *chess.Board, p chess.Player) error {
    move := p.Move()
    if putsOwnKingIntoCheck(b, move) {
        return &IllegalMoveError{Move: move, Reason: "puts own king in check"}
    }
    return nil
}
```

3. End users write simple tests against the validation function:

```go
func TestAcceptance(t *testing.T) {
    player := deepblue.New()
    if err := chesstest.ExercisePlayer(chesstest.StartingBoard(), player); err != nil {
        t.Errorf("Deep Blue player failed acceptance test: %v", err)
    }
}
```

Reserve `t.Fatal` for setup failures only---validation errors should be returned,
not fataled.

---

## Use Real Transports

> **Source**: Google Go Style Guide (best-practices)

When testing component integrations over HTTP or RPC, prefer real transport
round-trips over hand-implemented client mocks:

```go
func TestAPIIntegration(t *testing.T) {
    srv := httptest.NewTestServer(t, newFakeHandler())
    httpClient := srv.Client() // starts the in-memory server and fills srv.URL
    client := api.NewClient(httpClient, srv.URL)
    result, err := client.GetUser(t.Context(), "user-123")
    if err != nil {
        t.Fatalf("GetUser(%q) error = %v", "user-123", err)
    }
    if result.Name != "Test User" {
        t.Errorf("GetUser(%q).Name = %q, want %q", "user-123", result.Name, "Test User")
    }
}
```

The production client with a test server exercises as much real code as
possible. `httptest.NewTestServer` (Go 1.27+) registers its own cleanup and
serves over an in-memory network, usable inside a `synctest` bubble and
reachable only through `srv.Client()`: its URL is `http://example.com`, so a
constructor that builds its own client from the URL alone reaches the real
example.com. Take an `*http.Client` as a parameter, or, when the code under
test dials `srv.URL` itself, call `srv.Start()` before the first request: the
same server then listens on loopback, `srv.URL` names `127.0.0.1`, and its
cleanup stays registered.

---

## Real Databases

A test that needs a database runs against a real one — a container or a CI
service — in a file behind the `integration` build tag, and skips when no
database is configured:

```go
//go:build integration

package store_test

import (
    "database/sql"
    "os"
    "testing"

    _ "github.com/jackc/pgx/v5/stdlib" // registers "pgx" in this test binary; main's import does not reach it
)

// openTestDB opens the database named by TEST_DATABASE_URL and skips the test
// when it is unset.
func openTestDB(t *testing.T) *sql.DB {
    t.Helper()
    dsn := os.Getenv("TEST_DATABASE_URL")
    if dsn == "" {
        t.Skip("TEST_DATABASE_URL is not set")
    }
    db, err := sql.Open("pgx", dsn)
    if err != nil {
        t.Fatalf("sql.Open: %v", err)
    }
    t.Cleanup(func() {
        if err := db.Close(); err != nil {
            t.Errorf("close: %v", err)
        }
    })
    if err := db.PingContext(t.Context()); err != nil {
        t.Fatalf("ping the test database: %v", err)
    }
    return db
}
```

Run it with `go test -tags integration -race ./...`. A skip is not a pass:
`go test -v` prints `--- SKIP` with the reason, and a run whose database tests
skipped is reported as skipped.

---

## Common Mistakes

### Calling os.Exit directly in TestMain

`os.Exit` terminates the process immediately — deferred cleanup functions never
run. Return from `TestMain` instead; the test wrapper exits with the code
`m.Run` returned:

```go
// Bad: defers won't run
func TestMain(m *testing.M) {
    setup()
    defer cleanup()
    os.Exit(m.Run()) // cleanup() never executes
}

// Good: cleanup runs, then the wrapper exits with m.Run's code
func TestMain(m *testing.M) {
    setup()
    defer cleanup()
    m.Run()
}
```
