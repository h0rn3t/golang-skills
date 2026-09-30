# Must Functions

> Sources: source/uber-go-style/style.md (Do not Panic); https://pkg.go.dev/regexp#MustCompile
> Authority: advisory
> Last verified: 2026-09-29

`Must` functions wrap a fallible function and panic on error. Use them **only**
during program initialization, on a value fixed when the program is built, so
a failure is a bug rather than a missing file or a bad environment.

## Standard Library Examples

```go
// regexp.MustCompile panics if the pattern is invalid.
var validID = regexp.MustCompile(`^[a-z][a-z0-9-]{0,62}$`)

//go:embed index.html
var files embed.FS

// template.Must panics if the embedded template does not parse.
var tmpl = template.Must(template.ParseFS(files, "index.html"))
```

These are safe because their input is compiled into the binary: a failure is
a bug the first test run reports. `template.ParseFiles("index.html")` at
package init is not the same case — it reads the working directory at run
time, so it belongs in `main`, returning its error.

## When to Use Must

### Appropriate Uses

- **Package-level `var`**: Compiling regexp literals, parsing embedded
  templates
- **`init()`**: Building tables from the package's own constants
- **Test helpers**: `t.Fatal` is preferred in tests, but Must can be
  acceptable for test fixtures

Required config is not on this list: `main` (or the `run` it calls) loads it
and returns the error, so a missing file exits non-zero with a message.

### Never Use Must For

- Runtime request handling
- User-supplied input
- Network or file operations that can legitimately fail
- Anything called after program startup

## Writing a Must Function

Follow the naming convention `MustX` where `X` is the fallible function name:

```go
func MustParseRule(s string) Rule {
    r, err := ParseRule(s)
    if err != nil {
        panic(fmt.Sprintf("parsing rule %q: %v", s, err))
    }
    return r
}
```

```go
// MustParseRule is like ParseRule but panics if s is invalid. It simplifies
// the initialization of package-level variables holding rule literals.
func MustParseRule(s string) Rule { ... }
```
