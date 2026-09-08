package evals_test

import "testing"

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
