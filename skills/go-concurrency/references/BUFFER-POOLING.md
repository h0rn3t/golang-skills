# Buffer Pooling with Channels

> Sources: source/effective-go/effective_go.html (A leaky buffer)
> Authority: advisory
> Last verified: 2026-09-29

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

For production code, consider `sync.Pool` which provides similar functionality
with better integration into the garbage collector:

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
```

`sync.Pool` advantages:
- Automatic cleanup during garbage collection
- No pool-size bound to manage, but no per-object size bound either: put back
  only buffers under a capacity cap, as `fmt` drops print buffers over 64 KiB
  ([golang.org/issue/23199](https://golang.org/issue/23199))
- Thread-safe by design
- Better performance under high concurrency
