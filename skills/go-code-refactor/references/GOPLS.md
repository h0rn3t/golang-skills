# Rename and Extract with gopls

> Sources: golang.org/x/tools/gopls docs (`gopls help`, `gopls mcp`); Claude Code LSP tool docs
> Authority: advisory — the mechanics; behavior preservation rules stay in SKILL.md
> Minimum Go: gopls v0.20+ on PATH (`go install golang.org/x/tools/gopls@latest`)
> Last verified: 2026-09-02

`grep` finds text; `gopls` finds meaning. For a refactor that is a promise of
identical behavior, the difference is the whole job: a textual rename misses
the method that satisfies an interface in another package and hits the
unrelated local that happens to share the name. Reach for gopls whenever the
question is "who depends on this symbol" rather than "where does this string
appear".

## Three ways in

| Route | Addressing | Best for |
|---|---|---|
| gopls MCP server — `claude mcp add gopls -- gopls mcp` | Symbol names, file paths, fuzzy queries (`go_search`, `go_symbol_references`, `go_diagnostics`, `go_package_api`) | Agent workflows: no cursor position needed |
| Native `LSP` tool (`ENABLE_LSP_TOOL=1` + gopls wired) | `line:character` (`findReferences`, `goToImplementation`, `rename`, call hierarchy) | Right after a read or grep gave you a location; diagnostics arrive after every edit for free |
| `gopls` CLI — `gopls rename -w file.go:12:6 newName` | `file:line:col` | Nothing else is wired; one-shot scripted checks. Documented as experimental |

Prefer MCP → LSP → CLI. Absent all three, fall back to `go build ./... && go
vet ./...` after every rename and accept that interface satisfaction breaks are
found by the compiler, not before the edit.

## Before touching a definition

1. **References, not grep.** `go_symbol_references` / `findReferences` on the
   symbol. The count is the blast radius; read every referencing file that
   needs a matching edit before the first change.
2. **Implementations both ways.** For a method: `goToImplementation` on the
   interface it might satisfy; for an interface: every type that implements it.
   Renaming `Close` on one type silently un-implements `io.Closer` — gopls's
   rename refuses that; a text replace does not.
3. **Exported symbol?** References outside the module are invisible to gopls
   too. Exported renames are a findings-list item, not a diff (PLAYBOOK §3).

## Applying the change

- **Rename**: gopls `rename` updates every reference in the workspace,
  including test files, doc comments that mention the identifier in backticks,
  and struct-literal field keys. It rejects a rename that would shadow or
  collide. Review the diff anyway — a reject is safe, an accept is merely
  consistent.
- **Extract / inline**: the `refactor.extract.*` and `refactor.inline.*` code
  actions preserve side-effect order by construction, but may drop comments
  and produce a six-parameter helper when the seam is wrong (PLAYBOOK §2).
  Extract, then read the signature; if it is ugly, the split is in the wrong
  place — undo rather than patch.
- **Fill struct / add tags / remove unused parameter**: `refactor.rewrite.*`
  actions; `removeUnusedParam` also updates every call site.
- **Generated files** (`// Code generated ... DO NOT EDIT`) receive no code
  actions. Trace and update their source inputs as described in `SKILL.md`.

## After every edit

`go_diagnostics` on each changed file (automatic with the native tool). Fix
compiler errors before moving to the next transformation; a half-applied
rename across two files is the state in which `verify-refactor.sh after` lies
to you — the package that failed to compile ran no tests at all.

## Gotchas

- References reflect the **build configuration of the queried file**: a query
  from `x_linux.go` does not see `x_windows.go`. Re-run under `GOOS=windows`
  when build-tagged files are in the package (SKILL.md, Orient).
- Call hierarchy shows **static** calls only; calls through function values or
  interface methods are invisible — corroborate with references.
- gopls reasons about the locally resolved build (`go.sum`, `replace`
  directives). It cannot tell you who imports your exported API from other
  modules.
