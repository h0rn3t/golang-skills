---
name: go-testing
description: Use when writing, reviewing, or improving Go test code — including table-driven tests, subtests, parallel tests, test helpers, test doubles, and assertions with cmp.Diff. Also use when a user asks to write a test for a Go function, even if they don't mention specific patterns like table-driven tests or subtests. Does not cover benchmark performance testing (see go-performance).
allowed-tools: Bash(bash:*)
---

# Go Testing

> Compatibility: Baseline Go 1.27 (see the plugin repository's COMPATIBILITY.md).
> `httptest.NewTestServer` and `synctest.Sleep` require Go 1.27+;
> `testing/synctest` Go 1.25+; `t.Context` Go 1.24+. Diff examples use
> `github.com/google/go-cmp`, in a module that already requires it.

## Resource Routing

- `scripts/gen-table-test.sh` - Run from the project as `bash <installed-skill-dir>/scripts/gen-table-test.sh <FuncName> <package>` when generating a table-driven test scaffold.
- `assets/table-test-template.go` - Use as a copyable table-test starting point.
- `references/TEST-ORGANIZATION.md` - Read when choosing or naming test doubles, structuring packages, black-box tests, or larger test suites.
- `references/VALIDATION-APIS.md` - Read when designing an exported validation function or a `*test` package that other packages' tests call.
- `references/INTEGRATION.md` - Read when testing external services, HTTP handlers, databases, or long-running setup.
- `../go-http/references/JSON-V2.md` - Read when testing JSON v2 defaults, migration compatibility, or golden bytes (Go 1.27+).

## What Earns a Test

> **Normative**: A test exists for the bug it catches. Before writing a test
> or a case, name the plausible change to the code that turns it red; one that
> no such change can fail is not written. Scope: the tests this task writes —
> an existing passing test stays unless the user asks
> ([Holding the Bar](../go-linting/SKILL.md#holding-the-bar)).

- **Through the contract.** Call the exported entry point and assert what the
  documentation promises. An unexported helper gets its own test only when
  the entry point cannot reach its inputs (a parser, money arithmetic).
  Observe through the contract too: after `Create`, a `Get`, not a `SELECT`
  on the table behind it.
- **`want` comes from outside the code**: a literal or a hand-checked
  fixture, never the function under test, its helpers, or the same formula —
  `want := a.Price + b.Price` for `Total` passes whatever `Total` does.
- **One case per input class, plus each boundary**: nil and empty, one and
  many, the limit and one past it, each error path. A row that takes the same
  branch with the same class of input as another is a duplicate.
- **Nothing that cannot break on its own**: a getter or setter, a constructor
  that only assigns, a pass-through to the standard library or a dependency,
  a restated constant, generated code, or what the compiler checks —
  `var _ Store = (*memStore)(nil)` proves the interface. For a constant, test
  what depends on it: not `MaxRetries == 5` but five attempts and no sixth.
- **No mock echo.** Asserting that a double received what the code handed it
  proves the wiring, not the behavior: assert the observable result, through a
  fake ([TEST-ORGANIZATION.md](references/TEST-ORGANIZATION.md#test-double-types)).
  A spy stays where the call is the contract — an email sent, an event emitted.
  Where the package has the real store too, the cases live in one
  `func testStore(t *testing.T, s Store)` run against the fake and, under the
  `integration` tag, the real store; a case only the fake passes is a bug in
  the fake.
- **Coverage is a question, not a target.** An uncovered line asks which bug
  would go unnoticed there; when the answer is none, it stays uncovered.

Before closing, mutate the body in your head: a wrong constant or comparison,
the other branch, a dropped side effect, a zero-value return, a missing nil,
empty, or unauthorized check. Each needs a test that turns red; the one none
catches is the missing case. Review runs the same probe for real
([go-code-review](../go-code-review/SKILL.md#testing)).

---

## Use the Toolchain's Test APIs

> **Normative**: Prefer the standard-library helper over hand-rolled setup and
> teardown. Each one below removes a class of flake.

| Instead of | Use | Since |
|---|---|---|
| `context.Background()` in a test | `t.Context()`; work inside `t.Cleanup` uses `context.WithoutCancel(t.Context())`, since `t.Context()` is cancelled before cleanups run | 1.24 |
| `httptest.NewServer` + `defer srv.Close()` | `httptest.NewTestServer(t, h)` — registers cleanup; only `srv.Client()` reaches it until `srv.Start()` (below). At a `go 1.26` directive: `srv := httptest.NewServer(h); t.Cleanup(srv.Close)` | 1.27 |
| Real waits for timeout paths | `synctest.Sleep` inside a bubble (fake clock), from the test goroutine only: it calls `synctest.Wait`, which panics with `wait already in progress` when two goroutines reach it at once, so a goroutine the code under test starts sleeps with `time.Sleep`. At 1.26: `time.Sleep(d); synctest.Wait()` | 1.27 |
| A logger or command under test writing to `os.Stdout`/`os.Stderr` | `slog.New(slog.NewTextHandler(t.Output(), nil))` — printed with its test, on failure or `-v`; a `fmt.Println` in the test itself is `t.Log` | 1.25 |
| Ad-hoc temp dir for output to keep | `t.ArtifactDir()` with `go test -artifacts -outputdir=DIR` — otherwise removed after the test | 1.26 |

```go
func TestFetch(t *testing.T) {
    srv := httptest.NewTestServer(t, http.HandlerFunc(handle))
    client := srv.Client() // starts the in-memory server and fills srv.URL
    got, err := Fetch(t.Context(), client, srv.URL)
    if err != nil {
        t.Fatalf("Fetch(%q) error = %v, want nil", srv.URL, err)
    }
    if got != want {
        t.Errorf("Fetch(%q) = %q, want %q", srv.URL, got, want)
    }
}
```

`NewTestServer` serves over an in-memory network, so it works inside a
`synctest` bubble — and only `srv.Client()` reaches it. Its `srv.URL` is
`http://example.com` (empty until the first `Client()` call starts the server),
so code that builds its own client from that URL talks to the real
example.com instead of the handler. When the code under test dials `srv.URL`
with a client of its own, `srv.Start()` puts the same server on loopback and
keeps the registered cleanup and the handler-panic failure:

```go
srv := httptest.NewTestServer(t, http.HandlerFunc(handle))
srv.Start() // FetchURL builds its own client: listen on 127.0.0.1
got, err := FetchURL(t.Context(), srv.URL)
```

Inside a `synctest.Test` bubble the clock is fake and starts at 2000-01-01 UTC;
time advances only when every goroutine in the bubble is durably blocked. That
turns a two-second timeout test into a microsecond one, deterministically.
`go fix -testingcontext ./...` rewrites the `context.WithCancel` form.

---

## Useful Test Failures

> **Normative**: Test failures must be diagnosable without reading the test
> source.

Every failure message must include: function name, inputs, actual (got), and
expected (want). Use the format `YourFunc(%v) = %v, want %v`.

```go
t.Errorf("Add(2, 3) = %d, want %d", got, 5)
```

---

## Assertions: Match the Repository

> **Project policy**: `testify/assert` and `testify/require` are allowed,
> including in new projects. Follow the user's choice and existing repository
> conventions; do not replace working assertions just to change style. With
> no chosen convention, default to standard comparisons; for structured values
> use `cmp.Diff` when the module already requires `github.com/google/go-cmp`,
> otherwise `reflect.DeepEqual`, `slices.Equal`, or `maps.Equal`. A test never
> adds a module dependency. Enable `testifylint` when using testify.
> [go-style-core](../go-style-core/SKILL.md) owns the house-style rule.

Use `assert` for independent checks that can continue after failure and
`require` for prerequisites, only from the test goroutine; a wrapped error is
`require.ErrorIs`, not a match on its text. An `assert.New(t)` or
`require.New(t)` bound to the parent `t` and used inside `t.Run` fails the
wrong test. Through `assert` the failure lands on the parent and the subtest
stays green; the tell in `-v` output is a parent FAIL over subtests that all
PASS. A `require` through it calls `FailNow` on the parent from the
subtest's goroutine; the tell is `subtest may have called FailNow on a
parent test`; under `t.Parallel` subtests the binary dies instead with
`panic: test executed panic(nil) or runtime.Goexit`. Build it from the
subtest's own `t`.

For protocol buffers, add `protocmp.Transform()` as a cmp option. Always
include the direction key `(-want +got)` in diff messages. Compare serialized
output semantically when formatting is irrelevant; compare bytes when exact
serialization is the contract. For JSON v2 options and golden-test limits, see
[JSON v2 at API boundaries](../go-http/references/JSON-V2.md#json-in-golden-tests).

---

## t.Error vs t.Fatal

`t.Error` by default, so one run reports every failure; `t.Fatal` only when
the next check cannot run — setup failed, or it reads what the previous one
produced. Never `t.Fatal`/`t.FailNow` from a goroutine other than the test's.

---

## Table-Driven Tests

**Use table tests when:** all cases run the same code path with no conditional
setup, mocking, or assertions. A single `shouldErr` bool is acceptable.

**Don't use table tests when:** cases need complex setup, conditional mocking,
or multiple branches — write separate test functions instead.

**Key rules:**
- Use field names when cases span many lines or have same-type adjacent fields
- Include inputs in failure messages — never identify rows by index
- Subtest names are short and carry no `/`, which splits a `-run` pattern; no
  subtest depends on another's state or order
- `t.Parallel()` on each subtest is the default for a pure function
  (`gen-table-test.sh --parallel` emits it); leave it off when the table
  touches globals, `t.Setenv`, `t.Chdir`, or a shared fixture. Under a `go`
  directive of 1.22 or later no `tt := tt` capture line (`go fix -forvar`
  removes one in touched lines); below 1.22 it stays.

A table holds the cases someone thought of. A parser, decoder, or encoder
also gets a `FuzzX` seeded with the table's inputs and a property that needs
no `want`: a round trip (`Decode(Encode(x))` equals `x`), or an error instead
of a panic or a partial value. Plain `go test` runs the seeds; `-fuzz`
explores. The time-capped fuzz run for a parser at a trust boundary is
[go-security](../go-security/SKILL.md)'s.

> **Validation**: Running the tests, with `-race` and the pipeline's flags,
> belongs to the [go-linting](../go-linting/SKILL.md) gate, once, at the end
> of the task.

---

## Test Helpers

> **Normative**: Test helpers must call `t.Helper()` first and use `t.Cleanup()`
> for teardown: a `defer` in the helper runs when the helper returns, before
> the test has used what it built. A helper that opens a real database is in
> [INTEGRATION.md](references/INTEGRATION.md#real-databases).

```go
func newStore(t *testing.T) *Store {
    t.Helper()
    dir := t.TempDir()
    s, err := Open(dir)
    if err != nil {
        t.Fatalf("Open(%q) error = %v", dir, err)
    }
    t.Cleanup(func() {
        if err := s.Close(); err != nil {
            t.Errorf("close: %v", err)
        }
    })
    return s
}
```

---

## Test Error Semantics

> **Advisory**: Test the error a Go caller receives by its semantics
> (`errors.Is`, `errors.AsType`), not its message string. Text the
> documentation promises on the wire (a response body, a CLI message) is the
> contract and is compared as text.

A case that expects an error names which one, as `wantErr error`; a bare
`err != nil` stays green when an earlier step fails first. A nil `wantErr`
makes the same check require success:

```go
if !errors.Is(err, tt.wantErr) {
    t.Errorf("Parse(%q) error = %v, want %v", tt.in, err, tt.wantErr)
}
```

Presence alone fits only a function with one error path.

---

## Related Skills

- [go-error-handling](../go-error-handling/SKILL.md): designing the sentinels and error types a test matches.
- [go-interfaces](../go-interfaces/SKILL.md): test doubles implemented at the consumer side; their names are [TEST-ORGANIZATION.md](references/TEST-ORGANIZATION.md#test-double-naming-conventions)'s.
- [go-naming](../go-naming/SKILL.md): the identifiers a test declares.
- [go-linting](../go-linting/SKILL.md): linters alongside tests in CI.
