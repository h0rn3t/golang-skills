package evals_test

import (
	"strings"
	"testing"
)

func TestContiguousRowsDoNotOverwriteOnAppend(t *testing.T) {
	code := exampleBlock(t, "skills/go-data-structures/SKILL.md", "**Single allocation**")
	runExampleTest(t, `package example
import "testing"
func TestRowOwnership(t *testing.T) {
 const XSize, YSize = 2, 2
`+code+`
 picture[1][0] = 7
 picture[0] = append(picture[0], 9)
 if got := picture[1][0]; got != 7 { t.Errorf("row 1 after appending row 0 = %d, want 7", got) }
}
`)
}

func TestGo127GenericExamples(t *testing.T) {
	method, _, _ := strings.Cut(exampleBlock(t, "skills/go-generics/SKILL.md", "## Generic Methods"), "\nn, ok :=")
	hashing := exampleBlock(t, "skills/go-generics/SKILL.md", "### Hashing generic keys")
	inference, _, _ := strings.Cut(exampleBlock(t, "skills/go-generics/references/CONSTRAINTS.md", "## Type Inference"), "\n")
	runExampleTest(t, `package example
import ("hash/maphash"; "slices"; "testing")
`+method+hashing+`
func TestGenerics(t *testing.T) {
 store := &Store{data: map[string]any{"count": 42, "name": "alice"}}
 for _, tt := range []struct { key string; want int; ok bool }{
  {"count", 42, true}, {"name", 0, false}, {"missing", 0, false},
 } {
  if got, ok := store.Get[int](tt.key); got != tt.want || ok != tt.ok {
   t.Errorf("Get[int](%q) = %d, %t; want %d, %t", tt.key, got, ok, tt.want, tt.ok)
  }
 }
 if got, ok := store.Get[string]("name"); got != "alice" || !ok { t.Errorf("Get[string](name) = %q, %t", got, ok) }
 names := []string{"alice"}
`+inference+`
 if !result { t.Error("Contains([alice], alice) = false, want true") }
 tab := table[string, int]{hasher: maphash.ComparableHasher[string]{}, seed: maphash.MakeSeed()}
 var hash maphash.Hash
 hash.SetSeed(tab.seed)
 tab.hasher.Hash(&hash, "alice")
 first := hash.Sum64()
 hash.Reset()
 tab.hasher.Hash(&hash, "alice")
 if !tab.hasher.Equal("alice", "alice") || first != hash.Sum64() { t.Error("equal keys have inconsistent hashes") }
}
`)
}

func TestGo127PromotedLiteralExample(t *testing.T) {
	code := exampleBlock(t, "skills/go-style-core/references/INITIALIZATION.md", "### Embedded Fields in Go 1.27")
	runExampleTest(t, `package example
import "testing"
func TestLiteral(t *testing.T) {
 owner, name := "alice", "report"
`+code+`
 if doc.Name != name || doc.Audit.CreatedBy != owner {
  t.Errorf("literal=%+v, want Name=%q and Audit.CreatedBy=%q", doc, name, owner)
 }
}
`)
}

func TestGo127RetainedStringExample(t *testing.T) {
	code := exampleBlock(t, "skills/go-performance/references/STRING-OPTIMIZATION.md", "## Retained Substrings")
	runExampleTest(t, `package example
import ("strings"; "testing"; "unsafe")
func TestOwnership(t *testing.T) {
 record := "order-42" + strings.Repeat("x", 1<<20)
 for _, cut := range []int{0, 8} {
`+code+`
  if id != record[:cut] { t.Errorf("retained id=%q, want %q", id, record[:cut]) }
  // Pointer identity tests ownership without relying on GC timing or heap-size noise.
  if cut > 0 && unsafe.StringData(id) == unsafe.StringData(record) {
   t.Error("retained id still points into the large input allocation")
  }
 }
}
`)
}

func TestGo127MapUpdateExample(t *testing.T) {
	code := exampleBlock(t, "skills/go-data-structures/SKILL.md", "### Update and Filter an Existing Map")
	runExampleTest(t, `package example
import ("maps"; "testing")
func update(dst, updates map[string]int, minimum int) {
`+code+`
}
func TestUpdate(t *testing.T) {
 dst := map[string]int{"keep": 5, "drop": 1, "replace": 8}
 alias := dst
 updates := map[string]int{"replace": 0, "add": 3}
 before := maps.Clone(updates)
 update(dst, updates, 3)
 want := map[string]int{"keep": 5, "add": 3}
 if !maps.Equal(alias, want) { t.Errorf("alias=%v, want %v", alias, want) }
 if !maps.Equal(updates, before) { t.Errorf("updates mutated: %v, want %v", updates, before) }
 update(dst, nil, 5)
 if !maps.Equal(dst, map[string]int{"keep": 5}) { t.Errorf("nil updates: %v, want keep=5", dst) }
 update(dst, dst, 6)
 if len(dst) != 0 { t.Errorf("same-map update=%v, want empty", dst) }
 update(nil, nil, 0)
}
`)
}

func TestQueueExampleBoundsConcurrency(t *testing.T) {
	code := exampleBlock(t, "skills/go-concurrency/SKILL.md", "## Goroutine Lifetimes")
	runExampleTest(t, `package example
import ("context"; "sync"; "sync/atomic"; "testing"; "testing/synctest")
func consume(ctx context.Context, queue <-chan int, maxWorkers int, process func(context.Context, int)) {
`+code+`
}
func TestQueue(t *testing.T) {
 synctest.Test(t, func(t *testing.T) {
  queue := make(chan int, 64)
  for i := range cap(queue) { queue <- i }
  close(queue)
  release, done := make(chan struct{}), make(chan struct{})
  var active, completed atomic.Int32
  go func() {
   consume(t.Context(), queue, 8, func(_ context.Context, _ int) {
    active.Add(1)
    <-release
    active.Add(-1)
    completed.Add(1)
   })
   close(done)
  }()
  synctest.Wait()
  if got := active.Load(); got != 8 { t.Errorf("active tasks = %d, want 8", got) }
  close(release)
  <-done
  if got := completed.Load(); got != 64 { t.Errorf("completed tasks = %d, want 64", got) }
 })
}
func TestIdleCancellation(t *testing.T) {
 synctest.Test(t, func(t *testing.T) {
  ctx, cancel := context.WithCancel(t.Context())
  defer cancel()
  done := make(chan struct{})
  go func() {
   consume(ctx, make(chan int), 8, func(context.Context, int) { t.Error("unexpected item") })
   close(done)
  }()
  synctest.Wait()
  cancel()
  synctest.Wait()
  select {
  case <-done:
  default: t.Fatal("idle queue workers did not exit after cancellation")
  }
 })
}
`)
}
