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
- `references/TIME-ENUMS-TAGS.md` - Read when handling time types, struct tags, or embedding in public structs.

## Defensive Checklist Priority

When hardening code at API boundaries, check in this order:

```
Reviewing an API boundary?
├─ 1. Error handling     → Return errors; don't panic (see go-error-handling)
├─ 2. Input ownership    → Copy retained data when callers keep mutation rights
├─ 3. Output ownership   → Copy internal data when callers need independent mutation
├─ 4. Resource cleanup   → Use defer for Close/Unlock/Cancel
├─ 5. Interface checks   → Route compile-time assertions to go-interfaces
├─ 6. Time correctness   → Use time.Time and time.Duration, not int/float
├─ 7. Enum safety        → Choose useful default or unset by contract (go-style-core)
├─ 8. Crypto safety      → crypto/rand for keys, never math/rand
└─ 9. Path safety        → os.Root for caller-supplied paths
```

---

## Quick Reference

| Pattern | Rule | Details |
|---------|------|---------|
| Boundary copies | Copy to the depth required by independent ownership | [BOUNDARY-COPYING.md](references/BOUNDARY-COPYING.md) |
| Untrusted paths | `os.Root`, never `filepath.Join` + `os.Open` | Below |
| Defer cleanup | `defer f.Close()` right after `os.Open` | Below |
| Interface check | Compile-time satisfaction assertion | See go-interfaces |
| Time types | `time.Time` / `time.Duration`, never raw int | [TIME-ENUMS-TAGS.md](references/TIME-ENUMS-TAGS.md) |
| Enum start | Preserve a useful zero default; otherwise reserve unset | See go-style-core |
| Crypto rand | `crypto/rand` for keys, never `math/rand` | Below |
| Must functions | Only at init; panic on failure | [MUST-FUNCTIONS.md](references/MUST-FUNCTIONS.md) |
| Panic/recover | Return ordinary failures; recover only at a defined containment boundary | [PANIC-RECOVER.md](references/PANIC-RECOVER.md) |
| Mutable globals | Replace with dependency injection | Below |

---

## Common Pitfalls

| Pitfall | Rule |
|---|---|
| Typed nil in an interface | Return explicit `nil`; a `*T(nil)` in an `error` slot is non-nil (see [go-error-handling](../go-error-handling/SKILL.md)) |
| Bare `x.(T)` assertion | Use comma-ok; reflection code prefers `reflect.TypeAssert[T]` (Go 1.25+, see [go-interfaces](../go-interfaces/SKILL.md)) |
| `append` aliasing | Both slices share the backing array while capacity allows. `s[:len(s):len(s)]` only caps capacity so the next `append` reallocates — existing elements still alias; `slices.Clone(s)` is the copy (see [go-data-structures](../go-data-structures/SKILL.md)) |
| `int64` to `int32` without a bounds check | Values wrap silently; compare against `math.MaxInt32`/`math.MinInt32` first |
| Float comparison | Preserve exact comparisons required by the contract; approximate results need domain-chosen absolute/relative tolerances and explicit NaN/Inf handling. Money needs exact units or arithmetic |
| `defer` in a loop | Calls fire at function exit, not per iteration — extract the body (behavior note in [go-code-refactor](../go-code-refactor/references/BEHAVIOR-TRAPS.md)) |
| Nil channel | Send and receive block forever, so an unmade channel field is a hang, not an error — a deliberate `nil` in a `select` is the idiom for disabling that case (channel ownership: [go-concurrency](../go-concurrency/SKILL.md)) |
| Integer division by zero | Panics; guard the divisor (float division yields `Inf`/`NaN` instead) |

---

## Verify Interface Compliance

Route compile-time interface assertions to [go-interfaces](../go-interfaces/SKILL.md).
Use this skill only to notice API-boundary robustness risk; the interface skill
owns when an assertion is appropriate and the exact assertion shape.

## Copy Slices and Maps at Boundaries

Slices and maps reference underlying data. Copy when mutation must be independent;
fresh caller-owned results, explicit ownership transfer, and documented shared
views do not require a second copy. Stdlib clones implement shallow copying:

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

Field tags are a **serialization contract**. Preserve the wire name when
renaming a Go field; changing or removing its tag can break compatibility.

## Start Enums at One

Reserve zero for unset when absence must be detected. Keep zero as a valid
member when it is the useful documented default. [go-style-core](../go-style-core/SKILL.md)
owns this choice and the `iota` form; do not renumber an existing wire enum.

## Time, Struct Tags, and Embedding

Use `time.Time` and `time.Duration` for instants and spans, tag every
marshaled field, and avoid embedding types in public structs. See
[TIME-ENUMS-TAGS.md](references/TIME-ENUMS-TAGS.md).

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

Do not use `math/rand` or `math/rand/v2` to generate keys — this is a
**security concern**. Time-seeded generators have predictable output.

```go
import "crypto/rand"

func Key() string { return rand.Text() }
```

For text output, use `crypto/rand.Text` directly, or encode random bytes
with `encoding/hex` or `encoding/base64`.

## Confine Filesystem Access

When an untrusted path must stay inside a designated directory, open it through
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

Use `panic` only for truly unrecoverable situations. Library functions
should avoid panic.

```go
func safelyDo(work *Work) {
    defer func() {
        if err := recover(); err != nil {
            log.Println("work failed:", err)
        }
    }()
    do(work)
}
```

**Key rules:**
- Return ordinary failures as errors. Convert only intentional internal panics
  at the matching boundary; unexpected programming panics normally propagate.
- Acceptable to panic in `init()` if a library truly cannot set itself up
- Use recover to isolate panics in server goroutine handlers

## Must Functions

`Must` functions panic on error — use them **only** during program
initialization where failure means the program cannot run.

```go
var validID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)
var tmpl = template.Must(template.ParseFiles("index.html"))
```

---

## Related Skills

- **Error handling**: See [go-error-handling](../go-error-handling/SKILL.md) when choosing between returning errors and panicking, or wrapping errors at boundaries
- **Concurrency safety**: See [go-concurrency](../go-concurrency/SKILL.md) when protecting shared state with mutexes, atomics, or channels
- **Interface checks**: See [go-interfaces](../go-interfaces/SKILL.md) when adding compile-time interface satisfaction checks
- **Data structure copying**: See [go-data-structures](../go-data-structures/SKILL.md) when working with slice/map internals or pointer aliasing
- **Enum design**: See [go-style-core](../go-style-core/SKILL.md) when writing the `iota` block or choosing whether zero is a valid member
- **Threat model**: See [go-security](../go-security/SKILL.md) when the caller-supplied value is untrusted — injection, SSRF, secrets, TLS; this skill owns the `os.Root` and `crypto/rand` forms it routes to
