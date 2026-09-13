# Architecture reference at n=5 on `report` and `store`, and the `roster` neighbor fixture — Sonnet 5 medium

The [morning smoke](2026-09-13-go-arch-current-go-three-arm-n1-sonnet-5-medium.md)
read −14.2 against −11.5 lines on the refactor corpus at one repetition and
could not say whether the architecture reference had cost the skill anything.
This run repeats the two fixtures that carried that gap at five repetitions,
reference against baseline only. Between the two runs the working tree
changed: the user rewrote the architecture reference (669 lines, the
`handlers/services/repositories/models` vocabulary, "modules own the tree;
layers live inside the module") and it was split into `ARCHITECTURE.md`,
`ARCHITECTURE-SHAPES.md` and `ARCHITECTURE-ENFORCEMENT.md` to fit the
300-line cap; the baseline digest below is that tree.

The second half is the first run of `roster`, a fixture built the same day to
measure "Write Current Go": a package whose `legacy.go` neighbor is written in
pre-1.21 forms, beside three documented stubs.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `e77d572` (release 1.15.0) |
| `baseline` | the working tree: 1.15.0, the three architecture references, the normative "Write Current Go" and its routes |
| `no-skill` (roster only) | no plugin |

## Run

- Finished: 2026-09-13 12:48 UTC (`report`/`store`), 2026-09-13 12:48 UTC (`roster`)
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, reasoning effort `medium`; seed `1`
- `report`/`store`: corpus `refactor`, arms `reference`, `baseline`, 5 repetitions per fixture and arm, 20 sessions, `-j 4`, 0 CLI errors
- `roster`: corpus `implement`, arms `no-skill`, `reference`, `baseline`, 3 repetitions per arm, 9 sessions, `-j 3`, 0 CLI errors
- `reference` plugin SHA-256 `c3e5df8395d78c9bd143d0102a1f179d2e2539f771fdb267c3a7371fef461a26`
- `baseline` plugin SHA-256 `9b2ca49f67c6340aa780b9f1e482051d7acf22c2d3bcf21724c3196e08d28d8e`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2
- Refactor report: [`2026-09-13-go-refactor-architecture-report-store-n5-sonnet-5-medium.json`](2026-09-13-go-refactor-architecture-report-store-n5-sonnet-5-medium.json) (SHA-256 `59041337044cadef977e73c48c83f954f2a79b8f62a5b0b9523281aef35febc1`); traces [`….traces.tar.gz`](2026-09-13-go-refactor-architecture-report-store-n5-sonnet-5-medium.traces.tar.gz) (SHA-256 `2c89e6954cf1a12b3827c1e73ed0c81042dd12835025b5cdfee32aada72030d5`)
- Roster report: [`2026-09-13-go-implement-roster-neighbor-n3-sonnet-5-medium.json`](2026-09-13-go-implement-roster-neighbor-n3-sonnet-5-medium.json) (SHA-256 `47c59ac8143c5a0ac16c43fbbedbc8c5a70f8831cf956242a61d283498fd1cf3`); traces [`….traces.tar.gz`](2026-09-13-go-implement-roster-neighbor-n3-sonnet-5-medium.traces.tar.gz) (SHA-256 `84efac81f4c03398e18f8070b3569cd5d38cbde70a33d7baef930cbffeece6a2`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session in each
- Cost: `report`/`store` $2.17 reference, $1.80 baseline; `roster` $0.10, $0.60, $0.76; $5.43 in all

```bash
go run ./cmd/abrun -corpus refactor -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of e77d572> -arms reference,baseline \
  -tasks report,store -n 5 -j 4 -seed 1 -keep -verbose -out ../docs/evidence/2026-09-13-go-refactor-architecture-report-store-n5-sonnet-5-medium.json
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of e77d572> -arms no-skill,reference,baseline \
  -tasks roster -n 3 -j 3 -seed 1 -keep -verbose -out ../docs/evidence/2026-09-13-go-implement-roster-neighbor-n3-sonnet-5-medium.json
```

No session had a shell tool; the plugin's edit hook ran in every skilled
session and every skilled session loaded `go-code-refactor` (refactor) or
`go-code` (`roster`).

## `report` and `store` at n=5

| Fixture | Arm | Δlines per session | Mean | Δfuncs | Line gate | Golden | Lint clean | $/session |
|---|---|---|---|---|---|---|---|---|
| `report` | `reference` | +12, +15, +1, +13, +12 | +10.6 | 3, 3, 0, 3, 3 | 0/5 | 5/5 | 5/5 | 0.257 |
| `report` | `baseline` | +9, +7, −1, +12, +1 | +5.6 | 3, 1, 0, 3, 0 | 1/5 | 5/5 | 5/5 | 0.222 |
| `store` | `reference` | −6, −5, −6, −6, −15 | −7.6 | 0, 1, 0, 0, 0 | 5/5 | 5/5 | 5/5 | 0.178 |
| `store` | `baseline` | −11, −6, −10, −10, −6 | −8.6 | 0 | 5/5 | 5/5 | 5/5 | 0.139 |

Baseline against reference: `report` −5.0 lines (permutation p = 0.21),
`store` −1.0 (p = 0.80), both fixtures together −3.0 (p = 0.50).
Golden 10/10 and lint-clean 10/10 in both arms; `go fix` left nothing pending
in any session; new functions 13 against 7, the difference two `report`
sessions that extracted no helper. The morning's −14.2 against −11.5 was one
repetition of noise: at five the sign is the other way and neither fixture
separates.

**Reading and cost.** The baseline arm read all three architecture files in
10/10 sessions — 8/10 sessions read exactly the thirteen `go-code-refactor`
references and stopped. The reference arm read the ten references it has in
6/10 sessions and 55–58 references, the whole pack, in the other four. That
is where the cost went: baseline $0.18 a session against $0.22, −17%, with
about 16K tokens more architecture text per session. The routing condition
still does not gate the read; the wider spree is not attributable to either
tree.

## `roster` at n=3

`roster` has `legacy.go` in production form — `for i := 0; i < len(members); i++`,
`m := m`, `sort.Slice`, `strings.Index` plus slicing, `interface{}` — and three
stubs beside it. The golden test pins behavior only, including the legacy
functions so they cannot move.

| Arm | Δlines | Golden | `go fix` pending before → after | Lint before → after | `legacy.go` edited | Older form in the new bodies | $/session |
|---|---|---|---|---|---|---|---|
| `no-skill` | 22.0 | 3/3 | 3 → 3 | 5 → 5 | 0/3 | `sort.Strings` 3/3 | 0.032 |
| `reference` | 19.7 | 3/3 | 3 → 3 | 5 → 5 | 0/3 | 0/3; `slices.` in 3/3 | 0.201 |
| `baseline` | 18.3 | 3/3 | 3 → 3 | 5 → 5 | 0/3 | 0/3; `slices.` in 3/3 | 0.252 |

Three things the run establishes. The neighbor pulls: 3/3 unaided sessions
sorted with `sort.Strings` beside a `sort.Slice` neighbor, and 0/6 skilled
sessions did. No session in any arm touched `legacy.go` — the edit hook
printed its five `modernize` findings after every edit, and nobody widened
the scope. And the two skilled arms are indistinguishable: the reference tree
already wrote `slices.Sort` here through the reach-for table and
`go-data-structures`, so the normative "Write Current Go" had nothing left to
move on this model at this effort.

One thing the run corrects. The fixture's README named `go fix` pending hunks
after − before as the reading; it stayed 3 → 3 in 9/9 sessions because
`sort.Strings` is not a `slicessort` target (that modernizer rewrites
`sort.Slice` with a basic comparator). The older-form count above is a grep
over each session's `roster.go` for `sort.Strings`/`sort.Slice`,
`for i := 0; i <`, `m := m`, `strings.Index(` and `interface{}`; the README now
says so, and `abrun` has no such column yet.

## What this establishes

- The architecture reference costs the refactor corpus nothing measurable at
  n=5: lines −3.0 in the skill's favor (p = 0.50), correctness and lint
  tied at 10/10, cost −17%.
- `roster` separates the pack from no skill on the idiom (3/3 against 0/6) and
  does not separate the current-Go rule from the 1.15.0 text; its admission
  as a "Write Current Go" discriminator is not earned on Sonnet 5 medium.
- Sonnet 5 medium reads every reference a loaded skill names and, in 4/10
  reference sessions, every reference in the pack.

## Next

An older-form column in `abrun` (the grep above) would make `roster` readable
without a script. A model whose unaided arm copies the neighbor *and* whose
1.15.0 arm does too is the one on which the rule can be measured; Sonnet 5
medium is not it.
