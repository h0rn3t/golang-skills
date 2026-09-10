---
name: go-defensive
description: Use when hardening Go API boundaries or checking slice/map copies, aliasing, typed nils, cleanup/defer, narrowing overflow, float equality, nil channels, time, globals, interface compliance, or filesystem/crypto safety. Error strategy belongs to go-error-handling.
---

# Go Defensive Programming Patterns

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `url.URL.Clone` and
> `url.Values.Clone` require Go 1.27+; `crypto/rand.Text` and `os.Root` Go
> 1.24+; `slices.Clone`/`maps.Clone` Go 1.21+.

## Resource Routing

- `references/BOUNDARY-COPYING.md` - Read when copying slices/maps across API boundaries.
- `references/GLOBAL-STATE.md` - Read when introducing or removing package globals.
- `references/MUST-FUNCTIONS.md` - Read when deciding whether a panic-on-error helper is acceptable.
- `references/PANIC-RECOVER.md` - Read when evaluating panic, recover, or crash containment.
- `references/TIME-ENUMS-TAGS.md` - Read when handling time types or struct tags.
- `../go-http/references/JSON-V2.md` - Read when JSON v2 changes nil collections, tags, accepted input, or compatibility at an API boundary (Go 1.27+).

## Defensive Checklist Priority

When hardening code at API boundaries, check in this order:

```
Reviewing an API boundary?
├─ 1. Error handling     → Return errors; don't panic (see go-error-handling)
├─ 2. Input validation   → Copy slices/maps received from callers
├─ 3. Output safety      → Copy slices/maps before returning to callers
├─ 4. Resource cleanup   → Use defer for Close/Unlock/Cancel
├─ 5. Interface checks   → Route compile-time assertions to go-interfaces
├─ 6. Time correctness   → Use time.Time and time.Duration, not int/float
├─ 7. Enum safety        → Zero value must mean unset (see go-style-core)
├─ 8. Crypto safety      → crypto/rand for keys, never math/rand
└─ 9. Path safety        → os.Root for caller-supplied paths
```

---

## Common Pitfalls

| Pitfall | Rule |
|---|---|
| Typed nil in an interface | A `*T(nil)` stored in an interface — an `error` result, a field of interface type — is non-nil: `err != nil` is true and a call through it dereferences nil. Return a literal `nil`, and declare the result as `error`, not `*MyErr` ([go-error-handling](../go-error-handling/SKILL.md#core-rules) owns the API rule) |
| Bare `x.(T)` assertion | Comma-ok unless a mismatch is a programming error that should panic ([go-interfaces](../go-interfaces/SKILL.md#type-assertions-comma-ok-idiom)); reflection code prefers `reflect.TypeAssert[T]` (Go 1.25+) |
| `append` aliasing | Both slices share the backing array while capacity allows. `s[:len(s):len(s)]` only caps capacity so the next `append` reallocates — existing elements still alias; `slices.Clone(s)` is the copy (see [go-data-structures](../go-data-structures/SKILL.md)) |
| `int64` to `int32` without a bounds check | Values wrap silently; compare against `math.MaxInt32`/`math.MinInt32` first |
| Float `==` | Use an epsilon comparison; exact money math needs integer units or `math/big` |
| `defer` in a loop | Calls fire at function exit, not per iteration — extract the body (behavior note in [go-code-refactor](../go-code-refactor/references/BEHAVIOR-TRAPS.md)) |
| Nil channel | Send and receive block forever, so an unmade channel field is a hang, not an error — a deliberate `nil` in a `select` is the idiom for disabling that case (channel ownership: [go-concurrency](../go-concurrency/SKILL.md)) |
| Integer division by zero | Panics; guard the divisor (float division yields `Inf`/`NaN` instead) |

---

## Verify Interface Compliance

Route compile-time interface assertions to [go-interfaces](../go-interfaces/SKILL.md).
Use this skill only to notice API-boundary robustness risk; the interface skill
owns when an assertion is appropriate and the exact assertion shape.

## Copy Slices and Maps at Boundaries

Slices and maps contain pointers to underlying data. Copy at API boundaries to
prevent unintended modifications. Prefer stdlib clones when their nil and
capacity behavior fits the contract (conditions in the reference below):

```go
d.trips = slices.Clone(trips)      // receiving
return maps.Clone(s.counters)      // returning
return u.Clone()                   // *url.URL, Go 1.27+
return params.Clone()              // url.Values, Go 1.27+ (deep-copies values)
```

`Clone` is **shallow**: `[]*T` and `map[K][]V` copies still alias their
elements. See [BOUNDARY-COPYING.md](references/BOUNDARY-COPYING.md).

## Defer to Clean Up

Use `defer` to clean up resources (files, locks). Avoids missed cleanup on multiple return paths.

```go
p.Lock()
defer p.Unlock()

if p.count < 10 {
  return p.count
}
p.count++
return p.count
```

Defer overhead is negligible. Place `defer f.Close()` immediately after
`os.Open` for clarity. Arguments to deferred functions are evaluated when
`defer` executes, not when the function runs. Multiple defers execute in
LIFO order.

## Struct Field Tags

> **Advisory**: Always add explicit field tags to structs that are marshaled or unmarshaled.

```go
type User struct {
    Name  string `json:"name"  yaml:"name"`
    Email string `json:"email" yaml:"email"`
}
```

Field tags are a **serialization contract** — renaming a struct field without
updating the tag silently breaks wire compatibility. Treat tags as part of
the public API for any type that crosses a serialization boundary.
`encoding/json/v2` does not validate tag options: a copied `inline` or
`unknown` tag compiles and silently nests or drops data (`embed` is the
released spelling). Assert the encoded bytes in a test.

## Start Enums at One

An enum's zero value must not pass for a valid member — an unset field then
reads as a real state at the boundary. [go-style-core](../go-style-core/SKILL.md)
owns the `iota` form and the exception where zero is the sensible default.

## Time and Embedding

Use `time.Time` and `time.Duration` for instants and spans, never raw
integers ([TIME-ENUMS-TAGS.md](references/TIME-ENUMS-TAGS.md)). Embedding a
type in a public struct exports its whole method set —
[go-interfaces](../go-interfaces/SKILL.md#embedding) owns that rule.

## Avoid Mutable Globals

Inject dependencies instead of mutating package-level variables. This makes
code testable without global save/restore.

```go
type signer struct {
  now func() time.Time  // injected; tests replace with fixed time
}

func newSigner() *signer {
  return &signer{now: time.Now}
}
```

## Crypto Rand

Do not use `math/rand` or `math/rand/v2` to generate keys — both packages
document their output as predictable regardless of seeding and unsuitable for
security-sensitive work.

```go
import "crypto/rand"

func Key() string { return rand.Text() }
```

For text output, use `crypto/rand.Text` directly, or encode random bytes
with `encoding/hex` or `encoding/base64`.

## Confine Filesystem Access

When a path comes from a caller, request, or config file, open it through
`os.Root` (Go 1.24+) instead of `filepath.Join` + `os.Open`. `Root` resolves
every component inside the directory, so `../../etc/passwd` and a symlink
pointing out of the tree both fail instead of escaping.

```go
root, err := os.OpenRoot("/srv/uploads")
if err != nil { return err }
defer root.Close()

f, err := root.Open(userSuppliedName) // cannot escape /srv/uploads
```

`filepath.Clean` is **not** a substitute — it does not resolve symlinks.

---

## Panic and Recover

Use `panic` only for truly unrecoverable situations; library functions avoid
it. Never expose a panic across a package boundary — convert it to an error;
panicking in `init()` is acceptable only when a library cannot set itself up;
recover isolates panics in server goroutines. The recover guard and its traps
are in [PANIC-RECOVER.md](references/PANIC-RECOVER.md).

## Must Functions

`Must` functions panic on error — use them **only** during program
initialization (`regexp.MustCompile`, `template.Must` on package-level values)
where failure means the program cannot run; never on request-time input.
[MUST-FUNCTIONS.md](references/MUST-FUNCTIONS.md) has the shape and the exceptions.

---

## Related Skills

- **Error handling**: See [go-error-handling](../go-error-handling/SKILL.md) when choosing between returning errors and panicking, or wrapping errors at boundaries
- **Concurrency safety**: See [go-concurrency](../go-concurrency/SKILL.md) when protecting shared state with mutexes, atomics, or channels
- **Interface checks**: See [go-interfaces](../go-interfaces/SKILL.md) when adding compile-time interface satisfaction checks
- **Data structure copying**: See [go-data-structures](../go-data-structures/SKILL.md) when working with slice/map internals or pointer aliasing
- **Enum design**: See [go-style-core](../go-style-core/SKILL.md) when writing the `iota` block or choosing whether zero is a valid member
- **Threat model**: See [go-security](../go-security/SKILL.md) when the caller-supplied value is untrusted — injection, SSRF, secrets, TLS; this skill owns the `os.Root` and `crypto/rand` forms it routes to
