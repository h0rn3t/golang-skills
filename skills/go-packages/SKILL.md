---
name: go-packages
description: Use when creating Go packages, organizing imports, managing dependencies, or deciding how to structure Go code into packages. Also use when starting a new Go project or splitting a growing codebase into packages, even if the user doesn't explicitly ask about package organization. Does not cover naming individual identifiers (see go-naming).
---

# Go Packages and Imports

## Resource Routing

- `references/IMPORTS.md` - Read when grouping imports, using blank imports, dot imports, or import aliases.
- `references/PACKAGE-SIZE.md` - Read when splitting packages, avoiding `init`, structuring `main`, or designing CLI flags/subcommands.

> **When this skill does NOT apply**: For naming individual identifiers within a package, see [go-naming](../go-naming/SKILL.md). For organizing functions within a single file, see [go-functions](../go-functions/SKILL.md). For configuring linters that enforce import rules, see [go-linting](../go-linting/SKILL.md).

## Dependency Ladder

> **Normative**: Before adding a module to `go.mod`, check the stdlib. Go 1.27
> absorbed several of the most-added dependencies.

Stop at the first rung that works:

1. **Standard library** — check `go doc <pkg>` before assuming it is missing
2. **`golang.org/x/...`** — same release process, no third-party trust
3. **A module already in `go.mod`**
4. **A new module** — only when the above cost materially more code

Commonly added modules the standard library now covers:

| Was | Use instead | Since |
|---|---|---|
| `github.com/google/uuid` | `uuid` — `New`, `NewV4`, `NewV7`, `Parse`, `MustParse` | 1.27 |
| A faster JSON encoder | `encoding/json/v2` + `encoding/json/jsontext` | 1.27 |
| `github.com/sirupsen/logrus`, `go.uber.org/zap` | `log/slog` (see [go-logging](../go-logging/SKILL.md)) | 1.21 |
| `github.com/pkg/errors` | `fmt.Errorf` with `%w`, `errors.Is`, `errors.AsType` | 1.13 / 1.26 |
| `golang.org/x/exp/slices`, `.../maps` | `slices`, `maps` | 1.21 |
| A CSPRNG string helper | `crypto/rand.Text` | 1.24 |

`encoding/json/v2` is not a drop-in replacement for `encoding/json`: it changes
`omitempty`, case-matching, and error semantics. Keep v1 for existing wire
formats; reach for v2 for new code or when you need `jsontext` streaming. Both
ship in the toolchain — no build tag or `GOEXPERIMENT` needed on Go 1.27.

`github.com/google/uuid` stays justified for the algorithms stdlib `uuid` does
not have (v1, v3, v5, custom sources).

---

## Package Organization

### Avoid Util Packages

Package names should describe what the package provides. Avoid generic names
like `util`, `helper`, `common` — they obscure meaning and cause import
conflicts.

```go
// Good: Meaningful package names
db := spannertest.NewDatabaseFromFile(...)
_, err := f.Seek(0, io.SeekStart)

// Bad: Vague names obscure meaning
db := test.NewDatabaseFromFile(...)
_, err := f.Seek(0, common.SeekStart)
```

Generic names can be used as *part* of a name (e.g., `stringutil`) but should
not be the entire package name.

### Package Size

| Question | Action |
|----------|--------|
| Can you describe its purpose in one sentence? | No → split by responsibility |
| Do files never share unexported symbols? | Those files could be separate packages |
| Distinct user groups use different parts? | Split along user boundaries |
| Godoc page overwhelming? | Split to improve discoverability |

**Do NOT split** just because a file is long, to create single-type packages, or
if it would create circular dependencies.

---

## Imports

Imports are organized in groups separated by blank lines. Standard library
packages always come first. Use
[goimports](https://pkg.go.dev/golang.org/x/tools/cmd/goimports) to manage this
automatically.

```go
import (
    "fmt"
    "os"

    "github.com/foo/bar"
    "rsc.io/goversion/version"
)
```

**Quick rules:**

| Rule | Guidance |
|------|----------|
| Grouping | stdlib first, then external. Extended: stdlib → other → protos → side-effects |
| Renaming | Avoid unless collision. Rename the most local import. Proto packages get `pb` suffix |
| Blank imports (`import _`) | Only in `main` packages or tests |
| Dot imports (`import .`) | Never use, except for circular-dependency test files |

---

## Avoid init()

Avoid `init()` where possible. When unavoidable, it must be:

1. Completely deterministic
2. Independent of other `init()` ordering
3. Free of environment state (env vars, working dir, args)
4. Free of I/O (filesystem, network, system calls)

**Acceptable uses**: complex expressions that can't be single assignments,
pluggable hooks (e.g., `database/sql` dialects), deterministic precomputation.

---

## Exit in Main

Call `os.Exit` or `log.Fatal*` **only in `main()`**. All other functions should
return errors.

**Why**: Non-obvious control flow, untestable, `defer` statements skipped.

**Best practice**: Use the `run()` pattern — extract logic into
`func run() error`, call from `main()` with a single exit point:

```go
func main() {
    if err := run(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Command-Line Flags

> **Advisory**: Define flags only in `package main`.

- Flag names use `snake_case`: `--output_dir` not `--outputDir`
- Libraries should accept configuration as parameters, not read flags directly —
  this keeps them testable and reusable
- Prefer the standard `flag` package; use `pflag` only when POSIX conventions
  (double-dash, single-char shortcuts) are required

```go
// Good: Flag in main, passed as parameter to library
func main() {
    outputDir := flag.String("output_dir", ".", "directory for output files")
    flag.Parse()
    if err := mylib.Generate(*outputDir); err != nil {
        log.Fatal(err)
    }
}
```

---

## Build Constraints and Embedded Files

- Platform-specific code lives in `_linux.go` / `_windows.go` suffix files or
  behind `//go:build linux`; a runtime `if runtime.GOOS == ...` switch is the
  last resort. Every constrained file needs a fallback so the package still
  compiles under `go vet ./...` on any GOOS.
- Static assets (templates, SQL, schemas) ship via `//go:embed` in the package
  that uses them — never by reading a relative path at runtime, which breaks
  as soon as the binary runs from another directory.

```go
//go:embed schema/*.sql
var schemaFS embed.FS
```

---

## Related Skills

- **Package naming**: See [go-naming](../go-naming/SKILL.md) when choosing package names, avoiding stuttering, or naming exported symbols
- **Error handling across packages**: See [go-error-handling](../go-error-handling/SKILL.md) when wrapping errors at package boundaries with `%w` vs `%v`
- **Import linting**: See [go-linting](../go-linting/SKILL.md) when configuring goimports local-prefixes or enforcing import grouping
- **Global state**: See [go-defensive](../go-defensive/SKILL.md) when replacing `init()` with explicit initialization or avoiding mutable globals
