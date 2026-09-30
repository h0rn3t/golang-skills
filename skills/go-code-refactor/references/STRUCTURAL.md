# Moving Code Across Package Boundaries

> Sources: Go spec (package initialization, type identity); [Go 1.9 type alias proposal](https://go.googlesource.com/proposal/+/master/design/18130-type-alias.md); Go Modules Reference (semantic import versioning); `go doc` deprecation convention
> Authority: advisory — the mechanics; target layout belongs to `go-packages`
> Minimum Go: 1.27 baseline; type-parameterized aliases need 1.24+
> Last verified: 2026-09-06

The recipes below stage a cross-package move as a sequence where old and new
names both work, so no commit has to be a flag day. Where a type *should* live
is [go-packages](../../go-packages/SKILL.md)'s call.

---

## Breaking an import cycle

Four fixes, in preference order:

| # | Strategy | Call-site cost | Best when |
|---|---|---|---|
| 1 | Consumer-side interface | None, anywhere | The consumer calls one or two methods on the producer's type |
| 2 | Extract the shared type to a leaf package | Import path changes on both sides | Both packages need the same *concrete type*, not just its behavior |
| 3 | `internal/` package | Import path changes for the shared code | The shared code must never become public API |
| 4 | Mediator package | A new package both sides delegate to | 1–3 do not fit the shape of the coupling |

**1 is the idiomatic first move and costs nothing at any call site.**
[go-interfaces](../../go-interfaces/SKILL.md) owns the shape of that interface.
**4** is a last resort — confirm a consumer-side interface cannot express the
relationship first; it usually can.

## Moving a type: alias for gradual repair

`type A = B` declares that `A` and `B` are the **same** type, not merely
convertible. Code written against either name interoperates exactly, at zero
runtime cost and with no wrapper anywhere. The feature exists for this: its
design proposal names gradual code repair during large-scale refactoring, and
specifically moving a type between packages, as the motivation.

```go
// Step 1 — the new home carries the real definition.
package newpkg

type NewName struct { /* real fields */ }
```

```go
// Step 2 — the old home keeps the name working, and says it is going away.
package oldpkg

// Deprecated: use newpkg.NewName instead.
type OldName = newpkg.NewName
```

3. Migrate callers to the new import path incrementally — one commit, one
   package at a time. Both names stay valid and interchangeable throughout, so
   there is no window where some callers are broken and others are fixed.
4. Delete the alias once nothing references the old name.

Go 1.24 gave aliases type parameters, so the recipe is unchanged for a generic
type. A moved function can keep its old API through a thin forwarding function.
There is no cross-package alias for a mutable variable: `var Old = newpkg.New`
copies the current value, so later assignments diverge. Keep one storage owner
and expose accessors, or treat the move as a breaking API migration.

> **Normative**: The `=` is the whole mechanism. `type OldName NewName` without
> it declares a distinct type, and every existing value of the old type stops
> assigning to the new one — precisely the break the alias exists to avoid.

## Changing an exported API

- **Deprecate before deleting.** A doc comment starting `// Deprecated: ` is
  recognized by tooling and editors and shows up at every call site. Deleting an
  exported identifier outright breaks every downstream module at its next build
  with no warning beforehand.
- **Prefer additive, but audit compatibility.** A new top-level function or type
  is usually additive. An exported struct field can break unkeyed literals; a
  concrete method can collide with promoted methods; adding a method to an
  interface breaks its implementors. A change that must break callers is not a
  minor version after v1 — it needs a new major version, and Go expresses that
  with a `/vN` suffix in the module path and in every importer's import path,
  which is what lets v1 and v2 coexist in one build.

None of this is inside a refactor's promise of identical behavior. An exported
change is a findings-list item (PLAYBOOK §3) unless the user asked for it.

## Common mistakes

| Mistake | Fix | Why |
|---|---|---|
| Breaking a cycle by moving the producer's concrete type into the consumer | Declare the interface in the consumer, leave the type where it is | Moving the type usually relocates the cycle to whatever else that type depends on |
| Renaming a struct field during the move | Leave the field name alone, or list it as a finding | The `json`/`db` tag silently desyncs from the field, and a clean build says nothing about the wire format |
