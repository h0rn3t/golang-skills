# Functional Options vs Config Structs

> Sources: source/google-go-styleguide/best-practices.md; source/uber-go-style/style.md; [Go comparison rules](https://go.dev/ref/spec#Comparison_operators); [Go compatibility](https://go.dev/doc/go1compat)
> Authority: project policy
> Minimum Go: any supported Go version
> Last verified: 2026-09-05

Both functional options and config structs solve the same problem — optional
configuration for constructors — but they have different trade-offs. Choose
based on API audience, extensibility needs, and complexity budget.

Project policy: use Google guidance to decide whether optional configuration is
worth the complexity; when functional options are chosen, prefer Uber's
interface-with-unexported-method implementation where no repository convention
exists. Existing closure-based APIs remain valid. Test constructor behavior;
an interface does not make all concrete option values comparable.

## Contents

- [Decision Framework](#decision-framework)
- [Config Struct Pattern](#config-struct-pattern)
- [Comparison](#comparison)
- [When to Prefer Config Structs](#when-to-prefer-config-structs)
- [When to Prefer Functional Options](#when-to-prefer-functional-options)
- [Caller Ergonomics](#caller-ergonomics)
- [Functional Options Implementation](#functional-options-implementation)
- [Hybrid Approach](#hybrid-approach)

## Decision Framework

```
Need optional configuration?
├─ Internal or test-only API?
│   └─ Config struct (simpler, less boilerplate)
├─ All options usually specified together?
│   └─ Config struct (one literal, no With* ceremony)
├─ Public API with independently optional, growing settings?
│   └─ Consider functional options if they improve call sites
└─ Validation or interdependencies?
    └─ Either form; validate the final configuration in the constructor
```

Keep required inputs separate. An option count is a prompt to examine caller
needs, not a threshold that mandates a pattern. Support unset versus explicit
zero intentionally, and validate after defaults and overrides are combined.

## Config Struct Pattern

A config struct groups optional parameters into a single struct passed to the
constructor. Zero values serve as defaults, or provide a `DefaultConfig()`.

**Good**
```go
type Config struct {
    Timeout  time.Duration // zero = no timeout
    MaxRetry int           // zero = no retries
    Logger   *log.Logger   // nil = discard
}

func NewClient(addr string, cfg Config) *Client {
    if cfg.Logger == nil {
        cfg.Logger = log.New(io.Discard, "", 0)
    }
    return &Client{addr: addr, cfg: cfg}
}
```

```go
c := NewClient("localhost:8080", Config{
    Timeout:  5 * time.Second,
    MaxRetry: 3,
})
```

**Bad** — Relying on unexported config fields in a public API:
```go
type config struct {  // unexported: callers can't construct it
    timeout time.Duration
}

func NewClient(addr string, cfg config) *Client { ... }
```

### When Zero Values Don't Work

If zero is a valid non-default value (e.g., timeout of 0 means "no timeout"
but the desired default is 30s), use a pointer field or a sentinel value:

```go
type Config struct {
    Timeout *time.Duration // nil = use default (30s), zero = no timeout
}
```

## Comparison

| Aspect | Functional Options | Config Struct |
|--------|-------------------|---------------|
| **Boilerplate** | High (type + apply + With* per option) | Low (one struct) |
| **Extensibility** | Add `With*`; preserve existing defaults and behavior | Adding fields preserves keyed literals; unkeyed callers can break |
| **Backward compat** | Preserve required parameters and option semantics | Preserve zero-value meaning and constructor semantics |
| **Defaults** | Built into constructor | Zero values or `DefaultConfig()` |
| **Validation** | In `apply` or constructor loop | In constructor after struct received |
| **Discoverability** | `With*` functions appear in godoc | All fields visible in one struct |
| **Testability** | Assert configured behavior; compare only known comparable values | Assert fields/behavior; slices and maps also prevent `==` |
| **Caller experience** | Only specify what differs from defaults | Keyed literals can omit fields and be shared as data |
| **Zero-value ambiguity** | None — unset options not applied | May need pointer fields |

## When to Prefer Config Structs

- **Internal APIs** — less ceremony, easier to read at call sites
- **Few options (1-3)** — functional options overhead not justified
- **All options typically set together** — no benefit to variadic style
- **Validation is straightforward** — the constructor can validate a config struct too
- **Options are data, not behavior** — struct fields map naturally

```go
srv := NewServer(Config{
    Port:    8080,
    TLSCert: "/path/to/cert.pem",
    TLSKey:  "/path/to/key.pem",
})
```

## When to Prefer Functional Options

- **Public/library APIs** — callers shouldn't track internal config evolution
- **Several settings** that are independently optional at call sites
- **Complex defaults** — default computation depends on other options
- **Validation per option** — reject bad values at apply time
- **Options may grow** — new `With*` functions are purely additive

```go
srv := NewServer(
    WithPort(8080),
    WithTLS("/path/to/cert.pem", "/path/to/key.pem"),
    WithLogger(logger),
)
```

## Caller Ergonomics

Functional options are most valuable when they keep call sites focused on the
settings that differ from defaults.

**Before** — positional optional arguments hide meaning and force callers to
remember default sentinel values:

```go
conn, err := db.Open(addr, false, zap.NewNop(), 30*time.Second)
```

**After** — options name each non-default setting and allow the rest to stay
inside the constructor:

```go
conn, err := db.Open(
    addr,
    db.WithCache(false),
    db.WithLogger(logger),
)
```

## Functional Options Implementation

The pack's default implementation uses an exported `Option` interface with an
unexported `apply` method and concrete option values. Keep required inputs
separate, apply options over defaults, and validate their combined result before
allocating external resources. Document how repeated options behave.

The following excerpt shows the option application mechanics:

```go
package db

import "go.uber.org/zap"

type options struct {
    cache  bool
    logger *zap.Logger
}

type Option interface {
    apply(*options)
}

type cacheOption bool

func (c cacheOption) apply(opts *options) {
    opts.cache = bool(c)
}

func WithCache(c bool) Option {
    return cacheOption(c)
}

type loggerOption struct {
    Log *zap.Logger
}

func (l loggerOption) apply(opts *options) {
    opts.logger = l.Log
}

func WithLogger(log *zap.Logger) Option {
    return loggerOption{Log: log}
}

func Open(addr string, opts ...Option) (*Connection, error) {
    options := options{
        cache:  defaultCache,
        logger: zap.NewNop(),
    }

    for _, o := range opts {
        o.apply(&options)
    }

    return &Connection{}, nil
}
```

An existing API may instead use closure-based options:

```go
type Option func(*options)
```

Function values cannot be compared to each other with `==`; interface values
also panic on equality when their matching dynamic type is not comparable.
Named function types can have methods and implement interfaces. Choose by API
needs and repository convention, and test observable configured behavior.

## Hybrid Approach

For APIs that need both convenience and extensibility, accept a config struct
for common settings and functional options for advanced overrides:

```go
func NewServer(cfg Config, opts ...Option) *Server {
    s := &Server{cfg: cfg}
    for _, o := range opts {
        o.apply(&s.cfg)
    }
    return s
}
```

Use this sparingly — it adds complexity. Prefer one approach per API.
