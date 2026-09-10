# Time and Struct Tag Patterns

> Sources: source/uber-go-style/style.md (Use "time" to handle time, Start Enums at One, Use field tags in marshaled structs)
> Authority: advisory
> Last verified: 2026-09-10

## Use time.Time and time.Duration

Always use the `time` package. Avoid raw `int` for time values.

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

When `time.Duration` isn't possible, include unit in field name:

**Bad**
```go
type Config struct {
  Interval int `json:"interval"`
}
```

**Good**
```go
type Config struct {
  IntervalMillis int `json:"intervalMillis"`
}
```

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
