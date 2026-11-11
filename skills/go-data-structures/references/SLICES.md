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
fmt.Println(original) // [1 99 3 4 5] - modified!

// Fix: make independent copy
subset = slices.Clone(original[1:3])
```

### 2. Append May or May Not Reallocate

```go
a := make([]int, 3, 5)  // len=3, cap=5
b := append(a, 4)       // fits in capacity: b shares a's array
c := append(a, 5)       // fits too, and writes the same slot
fmt.Println(b[3], c[3]) // 5 5: the second append overwrote the first
d := append(a, 6, 7, 8) // exceeds capacity: d has its own array
d[0] = 9
fmt.Println(a[0], d[0]) // 0 9
```

### 3. Memory Leaks from Large Backing Arrays

```go
// Bad: small slice keeps entire file in memory
func getHeader(file []byte) []byte { return file[:100] }

// Good: copy to release the large array
func getHeader(file []byte) []byte { return bytes.Clone(file[:100]) }
```
