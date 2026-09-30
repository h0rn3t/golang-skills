---
name: go-testing
description: Use when writing, reviewing, or improving Go test code — including table-driven tests, subtests, parallel tests, test helpers, test doubles, and assertions with cmp.Diff. Also use when a user asks to write a test for a Go function, even if they don't mention specific patterns like table-driven tests or subtests. Does not cover benchmark performance testing (see go-performance).
allowed-tools: Bash(bash:*)
---

# Go Testing

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`).
> `httptest.NewTestServer` and `synctest.Sleep` require Go 1.27+;
> `testing/synctest` Go 1.25+; `t.Context` Go 1.24+. Diff examples use
> `github.com/google/go-cmp`, in a module that already requires it.

## Resource Routing

- `scripts/gen-table-test.sh` - Run when generating a table-driven test scaffold.
- `assets/table-test-template.go` - Use as a copyable table-test starting point.
- `references/TEST-ORGANIZATION.md` - Read when choosing or naming test doubles, structuring packages, black-box tests, or larger test suites.
- `references/VALIDATION-APIS.md` - Read when designing an exported validation function or a `*test` package that other packages' tests call.
- `references/INTEGRATION.md` - Read when testing external services, HTTP handlers, databases, or long-running setup.
- `../go-http/references/JSON-V2.md` - Read when testing JSON v2 defaults, migration compatibility, or golden bytes (Go 1.27+).

## Use the Toolchain's Test APIs

> **Normative**: Prefer the standard-library helper over hand-rolled setup and
> teardown. Each one below removes a class of flake.

| Instead of | Use | Since |
|---|---|---|
| `context.Background()` in a test | `t.Context()` | 1.24 |
| `httptest.NewServer` + `defer srv.Close()` | `httptest.NewTestServer(t, h)` — registers cleanup | 1.27 |
| Real waits for timeout paths | `synctest.Sleep` inside a bubble (fake clock), from the test goroutine only: it calls `synctest.Wait`, which panics with `wait already in progress` when two goroutines reach it at once, so a goroutine the code under test starts sleeps with `time.Sleep` | 1.27 |
| `fmt.Println` in a test | `t.Output()` — interleaves correctly under `-parallel` | 1.25 |
| Ad-hoc temp dir for output to keep | `t.ArtifactDir()` with `go test -artifacts -outputdir=DIR` — otherwise removed after the test | 1.26 |

```go
func TestFetch(t *testing.T) {
    srv := httptest.NewTestServer(t, http.HandlerFunc(handle))
    client := srv.Client() // starts the in-memory server and fills srv.URL
    got, err := Fetch(t.Context(), client, srv.URL)
    if err != nil {
        t.Fatalf("Fetch(%q) error = %v, want nil", srv.URL, err)
    }
    if diff := cmp.Diff(want, got); diff != "" {
        t.Errorf("Fetch(%q) mismatch (-want +got):\n%s", srv.URL, diff)
    }
}
```

`NewTestServer` serves over an in-memory network, so it works inside a
`synctest` bubble — and only `srv.Client()` reaches it. Its `srv.URL` is
`http://example.com` (empty until the first `Client()` call starts the server),
so code that builds its own client from that URL talks to the real
example.com instead of the handler. When the code under test cannot be handed
an `*http.Client`, keep `httptest.NewServer`.

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
`require.ErrorIs`, not a match on its text.

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

> **Advisory**: Test error semantics, not error message strings.

For simple presence checks when specific semantics don't matter:

```go
if gotErr := err != nil; gotErr != tt.wantErr {
    t.Errorf("f(%v) error = %v, want error presence = %t", tt.input, err, tt.wantErr)
}
```

---

## Related Skills

- [go-error-handling](../go-error-handling/SKILL.md): `errors.Is`/`errors.AsType` and sentinels under test.
- [go-interfaces](../go-interfaces/SKILL.md): test doubles implemented at the consumer side.
- [go-naming](../go-naming/SKILL.md): test, subtest, and helper names.
- [go-linting](../go-linting/SKILL.md): linters alongside tests in CI.
