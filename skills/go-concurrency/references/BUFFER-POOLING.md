# Buffer Pooling with Channels

> Sources: source/effective-go/effective_go.html (A leaky buffer); https://pkg.go.dev/net/http#RoundTripper
> Authority: advisory
> Last verified: 2026-10-11

Use a buffered channel as a free list to reuse allocated buffers, avoiding
repeated allocations. This "leaky buffer" pattern uses `select` with `default`
for non-blocking operations.

```go
var freeList = make(chan *bytes.Buffer, 100)

func getBuffer() *bytes.Buffer {
    select {
    case b := <-freeList:
        return b
    default:
        return new(bytes.Buffer)
    }
}

func putBuffer(b *bytes.Buffer) {
    if b.Cap() > 64<<10 {
        return // oversized: drop it, or the list pins the peak size
    }
    b.Reset()
    select {
    case freeList <- b:
    default: // free list full: drop b for the GC
    }
}
```

## Production Alternative

For a measured allocation hot spot
([go-performance](../../go-performance/SKILL.md#benchmarking-and-profiling)),
`sync.Pool` provides similar functionality with better integration into the
garbage collector. Whatever leaves the function is copied out before the buffer
goes back, because the next `getBuffer` reuses its array:

```go
var bufferPool = sync.Pool{
    New: func() any {
        return new(bytes.Buffer)
    },
}

func getBuffer() *bytes.Buffer {
    return bufferPool.Get().(*bytes.Buffer)
}

func putBuffer(b *bytes.Buffer) {
    if b.Cap() > 64<<10 {
        return // oversized: drop it, or the pool pins the peak size
    }
    b.Reset()
    bufferPool.Put(b)
}

func render(name string) []byte {
    buf := getBuffer()
    defer putBuffer(buf)
    fmt.Fprintf(buf, "hello %s", name)
    return bytes.Clone(buf.Bytes()) // not buf.Bytes(): putBuffer hands this array to the next caller
}
```

A buffer passed as an HTTP request body has left the function too: the
Transport may still be reading it after `Do` returns, and on HTTP/2 after it
closes the body ([go-http](../../go-http/SKILL.md#clients)). Such a buffer
does not go back to the pool: neither a deferred `putBuffer` around `Do` nor a
`Close` hook on the body is a point where those bytes are free.

`sync.Pool` advantages:
- Automatic cleanup during garbage collection
- No pool-size bound to manage, but no per-object size bound either: put back
  only buffers under a capacity cap, as `fmt` drops print buffers over 64 KiB
  ([golang.org/issue/23199](https://golang.org/issue/23199))
- Thread-safe by design
- Better performance under high concurrency
