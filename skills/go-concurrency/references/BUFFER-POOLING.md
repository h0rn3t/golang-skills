# Buffer Pooling with Channels

> Sources: source/effective-go/effective_go.html (A leaky buffer)
> Authority: advisory
> Last verified: 2026-09-29

Use a buffered channel as a free list to reuse allocated buffers, avoiding
repeated allocations. This "leaky buffer" pattern uses `select` with `default`
for non-blocking operations.

> **Source**: Effective Go

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

## How It Works

1. **Non-blocking receive**: Client tries to grab a buffer from `freeList`. If
   empty, `default` runs and allocates a new buffer.
2. **Non-blocking send**: Server tries to return the buffer. If `freeList` is
   full, `default` runs and the buffer is dropped for garbage collection.
3. **Bounded memory**: The channel capacity (100) limits the pooled count and
   the `Cap` check the size of each, so the list holds at most 100 × 64 KiB.

This pattern is useful when allocation is expensive and buffer reuse is
beneficial, but you don't want blocking behavior when the pool is empty or full.

## When to Use

- High-frequency allocations of similar-sized objects
- Performance-critical code paths where allocation overhead matters
- Scenarios where you want bounded memory usage

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

The channel-based approach is still valuable for understanding Go's concurrency
primitives and for cases where you need more control over pool behavior.
