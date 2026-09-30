# Test Helpers, Assertions, and Comparisons

> Sources: source/google-go-styleguide/decisions.md (Test helpers, Assertion libraries); source/uber-go-style/style.md (Test Tables)
> Authority: advisory
> Last verified: 2026-09-10

---

## Test Helper Pattern

A helper that opens a database — the DSN from the environment, a skip when it
is unset, the `Close` error checked in `t.Cleanup` — is in
[INTEGRATION.md](INTEGRATION.md#real-databases).

**Key rules:**
- Call `t.Helper()` as the first statement to attribute failures to the caller
- Use `t.Fatal` for setup failures (don't return errors from helpers)
- Use `t.Cleanup()` for teardown in a helper, not `defer`: a `defer` runs when
  the helper returns, before the test has used what it built; `t.Cleanup` runs
  when the test and its subtests finish

---

## Assertion Style

With testify, use semantic helpers such as `ErrorIs` for wrapped errors:

```go
// Imports: github.com/stretchr/testify/assert and .../require
got, err := GetPost(id)
require.NoError(t, err, "GetPost(%q)", id)
require.NotNil(t, got, "GetPost(%q)", id)
assert.Equal(t, "blogPost", got.Type, "GetPost(%q).Type", id)
assert.Equal(t, 2, got.Comments, "GetPost(%q).Comments", id)
```
