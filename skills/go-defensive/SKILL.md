---
name: go-defensive
description: Use when hardening Go API boundaries or checking slice/map copies, aliasing, typed nils, cleanup/defer, narrowing overflow, float equality, nil channels, time, globals, interface compliance, or filesystem/crypto safety. Error strategy belongs to go-error-handling.
---

# Go Defensive Programming Patterns

> Compatibility: Baseline Go 1.27 (see `COMPATIBILITY.md`). `url.URL.Clone` and
> `url.Values.Clone` require Go 1.27+; `Root.ReadFile` and `Root.MkdirAll` Go
> 1.25+; `crypto/rand.Text` and `os.Root` Go 1.24+; `slices.Clone`/`maps.Clone`
> Go 1.21+.

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
├─ 1. Error handling     → Return errors; panic only on a programming error (below)
├─ 2. Input validation   → Copy slices/maps received from callers
├─ 3. Output safety      → Copy slices/maps before returning to callers
├─ 4. Resource cleanup   → Use defer for Close/Unlock/Cancel
├─ 5. Interface checks   → Route compile-time assertions to go-interfaces
├─ 6. Time correctness   → time.Time/time.Duration in process; wire spans name a unit
├─ 7. Enum safety        → Zero value must mean unset (see go-style-core)
├─ 8. Crypto safety      → crypto/rand for keys, never math/rand
└─ 9. Path safety        → os.Root for caller-supplied paths
```

---

## Common Pitfalls

| Pitfall | Rule |
|---|---|
| Typed nil in an interface | A `*T(nil)` stored in an interface — an `error` result, a field of interface type — is non-nil: `err != nil` is true and a call through it dereferences nil. Return a literal `nil`, and declare the result as `error`, not `*MyErr` ([go-error-handling](../go-error-handling/SKILL.md#error-types) owns the API rule) |
| Bare `x.(T)` assertion | Comma-ok unless a mismatch is a programming error that should panic ([go-interfaces](../go-interfaces/SKILL.md#type-assertions-comma-ok-idiom)); reflection code prefers `reflect.TypeAssert[T]` (Go 1.25+) |
| `append` aliasing | Both slices share the backing array while capacity allows. `s[:len(s):len(s)]` only caps capacity so the next `append` reallocates — existing elements still alias; `slices.Clone(s)` is the copy (see [go-data-structures](../go-data-structures/SKILL.md)) |
| `int64` to `int32` without a bounds check | Values wrap silently; compare against `math.MaxInt32`/`math.MinInt32` first |
| Float `==` | A relative tolerance, `math.Abs(a-b) <= tol*max(math.Abs(a), math.Abs(b))` — a fixed epsilon is wrong at large and tiny magnitudes; money in integer minor units or `big.Rat`, never `float64` or `big.Float` (binary: ten `0.1`s do not sum to `1`) |
| `time.Time` `==`, or as a map key | `==` also compares the `Location` and the monotonic reading, so it fails after a JSON or database round trip or `t.In(loc)`: compare with `t.Equal(u)`; a key stores `t.UTC().Round(0)` or `t.UnixNano()` |
| `defer` in a loop | Calls fire at function exit, not per iteration — extract the body (behavior note in [go-code-refactor](../go-code-refactor/references/BEHAVIOR-TRAPS.md)) |
| Nil channel | Send and receive block forever, so an unmade channel field is a hang, not an error — a deliberate `nil` in a `select` is the idiom for disabling that case (channel ownership: [go-concurrency](../go-concurrency/SKILL.md)) |
| Integer division by zero | Panics; guard the divisor (float division yields `Inf`/`NaN` instead) |

---

## Copy Slices and Maps at Boundaries

Slices and maps contain pointers to underlying data. Copy at API boundaries to
prevent unintended modifications. Prefer stdlib clones when their nil and
capacity behavior fits the contract (conditions in the reference below):

```go
d.trips = slices.Clone(trips)        // receiving; a nil argument stays nil
d.tags = append([]string{}, tags...) // receiving, when d.tags must encode as []
return maps.Clone(s.counters)        // returning; a nil map stays nil
return u.Clone()                     // *url.URL, Go 1.27+
return params.Clone()                // url.Values, Go 1.27+ (deep-copies values)
```

`slices.Clone` and `maps.Clone` are **shallow**: `[]*T` and `map[K][]V` copies
still alias their elements. A type-specific `Clone` has only the copy depth in
its documented contract; do not substitute shallow and deep copies as a style
change. `slices.Clone` and `maps.Clone` also preserve nil, which
`encoding/json` v1 writes as `null`; for a required `[]` or `{}` output, keep
the non-nil `append` or `make` copy above. See
[BOUNDARY-COPYING.md](references/BOUNDARY-COPYING.md#copy-depth-is-part-of-the-contract).

## Defer to Clean Up

`Close` on a written file can report a write that failed late, so its error
joins the function's result through a named `err`:

```go
func save(name string, r io.Reader) (err error) {
    f, err := os.Create(name)
    if err != nil {
        return err
    }
    defer func() { err = errors.Join(err, f.Close()) }()
    _, err = io.Copy(f, r)
    return err
}
```

A file that is only read takes `os.ReadFile` or `root.ReadFile` when its whole
content fits in memory, and the same deferred `errors.Join` when it is
streamed.

## Struct Field Tags

> **Advisory**: Always add explicit field tags to structs that are marshaled or unmarshaled.

```go
type User struct {
    Name  string `json:"name"  yaml:"name"`
    Email string `json:"email" yaml:"email"`
}
```

JSON v2 tag options (`embed`, not the experimental `inline`) are in
[JSON-V2.md](../go-http/references/JSON-V2.md).

## Time

Use `time.Time` and `time.Duration` for instants and spans in process, never
raw integers. On the wire, `encoding/json/v2` (Go 1.27+) has no default
representation for `time.Duration` and fails to marshal it, so a wire field is
an integer with the unit in its name or a string parsed with
`time.ParseDuration` ([TIME-ENUMS-TAGS.md](references/TIME-ENUMS-TAGS.md#json-fields)).

## Avoid Mutable Globals

Inject dependencies instead of mutating package-level variables. This makes
code testable without global save/restore. The smallest injection is an
argument: the caller passes `time.Now()` and a test passes a fixed instant;
code that sleeps is tested inside `synctest.Test`, whose clock is fake
([GLOBAL-STATE.md](references/GLOBAL-STATE.md#injecting-time)).

## Crypto Rand

Do not use `math/rand` or `math/rand/v2` to generate keys.

```go
token := rand.Text() // crypto/rand: at least 128 bits, base32
```

For raw key material, `rand.Read(buf)` has no error to check: it never
returns one, and it crashes the program if the OS source fails.

## Confine Filesystem Access

When a path comes from a caller, request, config file, or archive entry,
open it through `os.Root` (Go 1.24+) instead of `filepath.Join` + `os.Open`.
`Root` resolves every component inside the directory, so `../../etc/passwd`
and a symlink pointing out of the tree both fail instead of escaping. A
server opens the root once at startup and keeps it on the struct that serves
requests:

```go
type server struct {
    uploads *os.Root // opened once, held for the server's lifetime
}

func newServer(dir string) (*server, error) {
    uploads, err := os.OpenRoot(dir)
    if err != nil {
        return nil, err
    }
    return &server{uploads: uploads}, nil
}

func (s *server) upload(name string) ([]byte, error) {
    return s.uploads.ReadFile(name) // cannot escape dir
}
```

`filepath.Clean` is **not** a substitute: it keeps a leading `..`, so
`filepath.Join(dir, filepath.Clean("../../etc/passwd"))` is `/etc/passwd`. A
name that is stored or compared but never opened — an object-store key, a
manifest entry — is checked with `filepath.IsLocal(name)`, which is lexical:
a name that is opened still goes through the root.

A subprocess is outside the root, because it resolves a name itself. Open the
file through the root and hand the child the open file (`cmd.Stdin`,
`cmd.ExtraFiles`), kept open until the child exits; a program that needs a
path needs a trusted staging directory or OS isolation, and `cmd.Dir`
confines nothing.

### Archive Entries

An entry's name, type, and size are all input. Create parents with
`root.MkdirAll`, never `os.MkdirAll(filepath.Join(dest, dir))`, which runs
outside the root; create no link an entry names; and cap the total size:

```go
func extract(root *os.Root, tr *tar.Reader, limit int64) error {
    for {
        hdr, err := tr.Next()
        if errors.Is(err, io.EOF) {
            return nil
        }
        if err != nil {
            return err
        }
        switch hdr.Typeflag {
        case tar.TypeReg: // extracted below
        case tar.TypeSymlink, tar.TypeLink:
            return fmt.Errorf("tar entry %q: links are not extracted", hdr.Name)
        default:
            continue // directories come from MkdirAll; devices and pax headers are skipped
        }
        if limit -= hdr.Size; limit < 0 {
            return errors.New("archive exceeds the size limit")
        }
        if err := root.MkdirAll(path.Dir(hdr.Name), 0o750); err != nil {
            return err // "../x" and "/x" escape the root and fail here
        }
        f, err := root.Create(hdr.Name)
        if err != nil {
            return err
        }
        _, err = io.Copy(f, tr)
        if err = errors.Join(err, f.Close()); err != nil {
            return err
        }
    }
}
```

An `archive/zip` entry takes the same three checks: `f.Mode().IsRegular()`,
`f.UncompressedSize64` against the cap, and the name through the root.

---

## Panic and Recover

An error caused by input or the environment — a malformed request, a missing
file, a refused connection — never crosses a package boundary as a panic:
return it. A programming error may cross as one: an argument the API's
documentation forbids, a `MustX` call at initialization, the unreachable
`default` of a switch over a closed enum. A panic a package raises for its own
control flow is recovered at its boundary and returned as an error.
`net/http` recovers only the goroutine running the handler, so a goroutine you
start recovers in its own deferred function or its panic ends the process.
The patterns and their traps are in
[PANIC-RECOVER.md](references/PANIC-RECOVER.md).

## Must Functions

`Must` functions panic on error — use them **only** at program
initialization on a value fixed when the program is built (`regexp.MustCompile`
on a pattern literal, `template.Must` on an embedded template), so a failure is
a bug; never on a file read at run time, the environment, or request-time input.
[MUST-FUNCTIONS.md](references/MUST-FUNCTIONS.md) has the shape and the exceptions.

---

## Related Skills

- [go-error-handling](../go-error-handling/SKILL.md): wrapping, sentinels, and typed errors at boundaries; this skill owns panic versus return.
- [go-interfaces](../go-interfaces/SKILL.md): compile-time interface assertions.
- [go-style-core](../go-style-core/SKILL.md): the `iota` block and whether zero is a valid member.
- [go-security](../go-security/SKILL.md): untrusted values: injection, SSRF, secrets, TLS; this skill owns the `os.Root` and `crypto/rand` forms.
