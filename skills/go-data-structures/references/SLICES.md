# Go Slice Internals

> Sources: source/effective-go/effective_go.html (Slices); https://go.dev/blog/slices-intro
> Authority: normative for slice semantics
> Last verified: 2026-09-10

## Slice Gotchas

### 1. Shared Underlying Array

```go
original := []int{1, 2, 3, 4, 5}
subset := original[1:3]
subset[0] = 99
fmt.Println(original)  // [1, 99, 3, 4, 5] - modified!

// Fix: make independent copy
subset := slices.Clone(original[1:3])
```

### 2. Append May or May Not Reallocate

```go
a := make([]int, 3, 5)  // len=3, cap=5
b := a[0:3]
a = append(a, 4)    // Fits in capacity - still shared
a = append(a, 5, 6) // Exceeds capacity - now independent
```

### 3. Memory Leaks from Large Backing Arrays

```go
// Bad: small slice keeps entire file in memory
func getHeader(file []byte) []byte { return file[:100] }

// Good: copy to release the large array
func getHeader(file []byte) []byte { return bytes.Clone(file[:100]) }
```
