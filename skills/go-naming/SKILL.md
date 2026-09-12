---
name: go-naming
description: Use when naming Go packages, types, functions, methods, variables, constants, or receivers, including new types and exported APIs. Package organization belongs to go-packages.
allowed-tools: Bash(bash:*)
---

# Go Naming Conventions

## Resource Routing

- `scripts/check-naming.sh` - Run when checking SCREAMING_SNAKE_CASE constants, Get-prefixed getters, generic package names, or receivers named `this`/`self`.
- `scripts/check-naming-ast.go` - Implementation helper invoked by `check-naming.sh`; patch this when changing what counts as a naming violation.
- `references/IDENTIFIERS.md` - Read for the conventions themselves — package, interface, receiver, constant, initialism, getter and type-suffix names — when a choice for an exported identifier or package-level symbol is in doubt.
- `references/REPETITION.md` - Read when names repeat package, receiver, type, or local context.
- `references/VARIABLES.md` - Read when choosing local variable names, receiver names, or loop identifiers.

## Core Principle

A name is read where it is used, never where it is declared: it does not
repeat the package, the receiver, or the surrounding context, and its length
grows with the distance between declaration and use. Go names are shorter
than in most languages. The reader knows the conventions — MixedCaps,
lowercase packages, `-er` interfaces, short consistent receivers, no `Get`
prefix, initialisms in one case — so this skill carries the decisions and
`references/` the rules.

---

## Naming Decision Flow

```
What are you naming?
├─ Package       → Short, lowercase, singular noun (no underscores, no mixedCaps)
├─ Interface     → Method name + "-er" suffix when single-method (Reader, Writer)
├─ Receiver      → 1-2 letter abbreviation of type (c for Client); consistent across methods
├─ Constant      → MixedCaps; use iota for enums; no ALL_CAPS
├─ Exported func → Verb or verb-phrase in MixedCaps; no Get prefix for getters
├─ Variable      → Length proportional to scope distance
│                  ├─ Tiny scope (1-7 lines) → single letter (i, n, r)
│                  ├─ Medium scope           → short word (count, buf)
│                  └─ Package-level / wide   → descriptive (userAccountCount)
└─ Any name      → Check: does it repeat package name or context? If yes, shorten it
```

---

## Error Names

> **Normative**: Sentinel errors are `ErrNotFound` when exported and
> `errNotFound` when not; error types end in `Error` (`NotFoundError`), never
> `ErrNotFoundError`. `errname` in the lint baseline enforces both.

---

## Where Review Sends Names Back

- **Repetition against context.** `widget.New()` not `widget.NewWidget()`;
  `p.Name()` not `p.ProjectName()`; in package `sqldb`, `Connection` not
  `DBConnection`. A generic package name — `util`, `common`, `helper` — is a
  finding; name the package for what it provides (`httpauth`, `stringutil`).
  [REPETITION.md](references/REPETITION.md) has the cases.
- **`_` prefix on unexported globals** (Uber only; Google style does not use
  it): follow the repository. Never introduce the prefix into a codebase that
  lacks it — [go-style-core](../go-style-core/SKILL.md#house-style-wins) owns
  the house-style rule.
- **Type in the name.** `users` not `userSlice`, `name` not `nameString`;
  when functions differ only by type, the type goes at the end (`ParseInt`,
  `ParseInt64`).
- **A predeclared identifier as a name.** `error`, `string`, `len`, `cap`,
  `append`, `copy`, `new`, `make` as a variable, parameter, or type name
  compiles and hides the built-in for the rest of the scope;
  [SHADOWING.md](../go-style-core/references/SHADOWING.md) owns detection.

```go
for i, v := range items { ... }           // small scope
pendingOrders := filterPending(orders)    // larger scope
const _defaultPort = 8080                 // Uber-style prefix — only where the repo already uses it
```

---

## Validation

> **Validation**: `scripts/check-naming.sh` reports the anti-patterns above; it runs with the build and the rest of the [go-linting](../go-linting/SKILL.md) gate, once, at the end of the task.

## Related Skills

- [go-interfaces](../go-interfaces/SKILL.md): pointer versus value receivers; their names are [IDENTIFIERS.md](references/IDENTIFIERS.md#receiver-names)'s Receiver Names.
- [go-packages](../go-packages/SKILL.md): splitting packages, import collisions; package names are [IDENTIFIERS.md](references/IDENTIFIERS.md#package-names)'s Package Names.
- [go-style-core](../go-style-core/SKILL.md): name length by scope, shadowing of built-ins, clarity against concision.
