# Time and Struct Tag Patterns

> Sources: source/uber-go-style/style.md (Use "time" to handle time, Start Enums at One, Use field tags in marshaled structs)
> Authority: advisory
> Last verified: 2026-09-29

## Use time.Time and time.Duration

Use the `time` package for time values in process. A raw integer appears
only on the wire, with its unit in the field name ([JSON Fields](#json-fields)).

### Instants

**Bad**
```go
func isActive(now, start, stop int) bool {
  return start <= now && now < stop
}
```

**Good**
```go
func isActive(now, start, stop time.Time) bool {
  return (start.Before(now) || start.Equal(now)) && now.Before(stop)
}
```

### Durations

**Bad**
```go
func poll(delay int) {
  time.Sleep(time.Duration(delay) * time.Millisecond)
}
poll(10)  // seconds? milliseconds?
```

**Good**
```go
func poll(delay time.Duration) {
  time.Sleep(delay)
}
poll(10 * time.Second)
```

### JSON Fields

`time.Duration` is the in-process type, not a wire form: `encoding/json/v2`
has no default representation for it (`Marshal` and `Unmarshal` both fail with
`no default representation`), and v1 writes bare nanoseconds. A `time.Time`
field is fine in both, as an RFC 3339 string.

**Bad**
```go
type Config struct {
  Interval int           `json:"interval"` // unit unknown
  Timeout  time.Duration `json:"timeout"`  // json/v2: an error; v1: nanoseconds
}
```

**Good**: an integer with the unit in the name, converted at the boundary

```go
type Config struct {
  IntervalMillis int64 `json:"intervalMillis"`
}

func (c Config) interval() time.Duration {
  return time.Duration(c.IntervalMillis) * time.Millisecond
}
```

**Good**: a string, when people write the value by hand

```go
type Config struct {
  Timeout string `json:"timeout"` // "1m30s"
}

func (c Config) timeout() (time.Duration, error) {
  return time.ParseDuration(c.Timeout)
}
```

An existing wire format that carries nanoseconds keeps working under v2 with
the `jsonv1.FormatDurationAsNano(true)` option
([JSON-V2.md](../../go-http/references/JSON-V2.md) owns the v2 defaults).

## Embedding in Public Structs

Owned by [go-interfaces](../../go-interfaces/references/EMBEDDING.md#dont-embed-in-public-structs):
an embedded type's method set becomes public API, so adding, removing, or
replacing it is a breaking change. Use an unexported field and forward methods.

## Use Field Tags in Marshaled Structs

Always use explicit field tags for JSON, YAML, etc.

**Bad**
```go
type Stock struct {
  Price int
  Name  string
}
```

**Good**
```go
type Stock struct {
  Price int    `json:"price"`
  Name  string `json:"name"`
  // Safe to rename Name to Symbol
}
```

Tags make the serialization contract explicit and safe to refactor.
