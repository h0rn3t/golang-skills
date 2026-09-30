# Embedding Patterns in Go

> Sources: source/effective-go/effective_go.html (Embedding); source/uber-go-style/style.md (Avoid Embedding Types in Public Structs)
> Authority: advisory
> Last verified: 2026-09-29

## Don't Embed in Public Structs

Embedding exposes the inner type's full method set as part of your public API.
This creates a maintenance burden: changes to the embedded type's methods
break your API's compatibility guarantees.

**Bad**
```go
type SMap struct {
    sync.Mutex  // Lock and Unlock are now part of SMap's API
    data map[string]string
}
```

**Good**
```go
type SMap struct {
    mu   sync.Mutex  // unexported field — implementation detail
    data map[string]string
}

func (m *SMap) Get(k string) string {
    m.mu.Lock()
    defer m.mu.Unlock()
    return m.data[k]
}
```

Exception: Embedding is acceptable in test types and internal structs where
API stability is not a concern.
