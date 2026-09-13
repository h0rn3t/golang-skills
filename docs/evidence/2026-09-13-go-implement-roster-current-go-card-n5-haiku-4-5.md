# The idiom card on `roster` at n=5 — Haiku 4.5

The [afternoon run](2026-09-13-go-arch-current-go-n5-report-store-roster-n3-sonnet-5-medium.md)
found that Sonnet 5 medium writes no older form on `roster` with either the
1.15.0 text or the normative "Write Current Go", so the rule could not be
measured there, and named the model that could measure it: one whose unaided
arm copies the pre-1.21 neighbor *and* whose skilled arm still does. This run
is that measurement, on the change made since: the write-time idiom card
`go-style-core/references/CURRENT-GO.md` (one line per idiom with version and
trap, built from the 54 JetBrains go-modern-guidelines items) and the owner
additions that went in with it.

| Arm | Tree |
|---|---|
| `no-skill` | no plugin |
| `reference` | `git worktree` of `fe71e2f` (release 1.16.0): the normative "Write Current Go", no card |
| `baseline` | the working tree: 1.16.0 plus `CURRENT-GO.md`, its routes from `go-style-core`, `go-code` and the review checklist, and the owner additions of the same change |

## Run

- Finished: 2026-09-13 18:08 UTC
- Runner: `claude` 2.1.267; model `claude-haiku-4-5-20251001`, no reasoning-effort flag; seed `1`
- Corpus `implement`, fixture `roster`, arms `no-skill`, `reference`, `baseline`, 5 repetitions per arm, 15 sessions, `-j 4`, 0 CLI errors
- `reference` plugin SHA-256 `d89942461c1ba8d5612d704fd298ac8ad1f57d7059cd2ae1d9dc20ebf05ef9f4`
- `baseline` plugin SHA-256 `b5b5ebc4a3134261020dd54f6116bd232e7043fc0605ade2ce0802eb2e660c9b`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint not installed, so the lint columns are unmeasured in every session
- Report: [`2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.json`](2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.json) (SHA-256 `64c041650f192bad269b6b84d1e9a14339488c391cb8d2e8739378fd70cb03bb`); traces [`….traces.tar.gz`](2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.traces.tar.gz) (SHA-256 `7177346a4ed71115a287ac2a0335c3698c316409ce38ba8a8d72aa969eb6bbee`), one `traces/<arm>-roster-r<rep>.jsonl` per session
- Cost: $0.135 `no-skill`, $0.497 `reference`, $0.632 `baseline`; $1.27 in all

```bash
go run ./cmd/abrun -corpus implement -tasks roster -runner claude -model claude-haiku-4-5-20251001 \
  -reference-root <worktree of fe71e2f> -arms no-skill,reference,baseline \
  -n 5 -j 4 -seed 1 -timeout 10m -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.json
```

No session had a shell tool; the plugin's edit hook ran in every skilled
session (`gofmt`, `go vet`, `go fix -diff`, the package's tests; no linter on
this machine). `go-code` loaded in 10/10 skilled sessions.

## The reading

The fixture's trap is not a golden failure: the golden test pins behavior and
the legacy functions, and the reading is a grep over each session's
`roster.go` for the neighbor's older forms (`sort.Strings`/`sort.Slice`,
`for i := 0; i <`, `m := m`, `strings.Index(`, `interface{}`). `go fix`
pending hunks stay at the neighbor's three whatever the new code does, because
`sort.Strings` is no modernizer's target.

| Arm | Valid | Golden | Older form in `roster.go` | `slices.Sort` | `slices.ContainsFunc` | Card read (`Read` call) | Δlines | $/session |
|---|---|---|---|---|---|---|---|---|
| `no-skill` | 4/5 | 4/5 | **5/5** `sort.Strings` | 0/5 | 0/5 | — | +22.0 | 0.027 |
| `reference` | 5/5 | 5/5 | **2/5** `sort.Strings` | 3/5 | 0/5 | — | +22.4 | 0.099 |
| `baseline` | 5/5 | 5/5 | **0/5** | 5/5 | 4/5 | 4/5 | +18.8 | 0.126 |

Per session:

| Arm | Rep | Δlines | Golden | Older form | Modern forms | Skills fired | Card `Read` | $ |
|---|---|---|---|---|---|---|---|---|
| `no-skill` | 0 | +22 | pass | `sort.Strings` | — | — | — | 0.024 |
| `no-skill` | 1 | +20 | **build failed** | `sort.Strings` | — | — | — | 0.020 |
| `no-skill` | 2 | +22 | pass | `sort.Strings` | — | — | — | 0.034 |
| `no-skill` | 3 | +22 | pass | `sort.Strings` | — | — | — | 0.030 |
| `no-skill` | 4 | +22 | pass | `sort.Strings` | — | — | — | 0.027 |
| `reference` | 0 | +22 | pass | — | `slices.Sort` | go-code, go-style-core | — | 0.098 |
| `reference` | 1 | +22 | pass | `sort.Strings` | — | go-code, go-style-core | — | 0.083 |
| `reference` | 2 | +22 | pass | `sort.Strings` | — | go-code, go-style-core | — | 0.072 |
| `reference` | 3 | +22 | pass | — | `slices.Sort` | go-code, go-style-core, go-testing | — | 0.170 |
| `reference` | 4 | +24 | pass | — | `slices.Sort` | go-code, go-style-core | — | 0.074 |
| `baseline` | 0 | +22 | pass | — | `slices.Sort` | go-code, go-linting, go-style-core | yes | 0.142 |
| `baseline` | 1 | +18 | pass | — | `slices.Sort`, `ContainsFunc`, `Compact` | go-code, go-data-structures, go-linting, go-style-core, go-testing | no | 0.143 |
| `baseline` | 2 | +15 | pass | — | `slices.Sort`, `ContainsFunc` | go-code | yes | 0.123 |
| `baseline` | 3 | +19 | pass | — | `slices.Sort`, `ContainsFunc`, `Compact` | go-code | yes | 0.077 |
| `baseline` | 4 | +20 | pass | — | `slices.Sort`, `ContainsFunc` | go-code, go-testing | yes | 0.148 |

The unaided failure is the stale idiom itself: session 1 wrote
`sort.Strings(names)` and never added the import, so the package did not
build (`roster/roster.go:26:2: undefined: sort`). The other four unaided
sessions are the same 22 lines as the neighbor would have written them —
`map[string]bool`, an append loop, `sort.Strings`, a hand-written search.

The two `reference` sessions that copied `sort.Strings` had `go-style-core`
in context with the normative rule and its one example; the rule alone moved
Haiku from 5/5 to 2/5. Every `baseline` session wrote `slices.Sort`, four
wrote `slices.ContainsFunc` for `OnTeam`, two `slices.Compact` after the sort
instead of a seen-map. The card reached sessions that never loaded
`go-style-core` as a skill: sessions 2 and 3 fired `go-code` only and read the
card through its routing line. Session 1 did not `Read` the card and still
wrote three modern forms; it had loaded `go-data-structures`, whose table
carries the same rows.

Two things the card did not move. `map[string]bool` as a set stayed in 5/5,
5/5 and 3/5 sessions — a pack preference (`map[T]struct{}`), not a JetBrains
item, and not on the card. And `baseline` session 2 rewrote `legacy.go`, the
neighbor the fixture's package comment says "is not part of this change":
`for i := range members`, the `m := m` line deleted, `slices.Sort`,
`strings.Cut`, `any` — five correct modernizations of a file it was told to
leave, which is why `go fix` pending fell 3 → 0 in that session and the arm
mean reads 2.40. Behavior held (golden 5/5), so no column caught it; the
"Scope stays too: untouched neighbors are not rewritten" sentence of Write
Current Go did not hold in 1/5 sessions on this model.

Cost: the card arm is $0.126 a session against $0.099, +27% over the
reference, 4.7× the unaided $0.027. The card is ~9.2K characters; the
reference arm read `MODERNIZATION.md` in no session and the baseline arm read
the card in four.

## What this establishes

- `roster` discriminates on Haiku 4.5 the way it could not on Sonnet 5
  medium: 5/5 older forms unaided, 2/5 with the 1.16.0 text, 0/5 with the
  card, golden 5/5 in both skilled arms.
- The card is what closed the gap, not the rule: the arm that differs from
  `reference` only by the card and the owner rows went from 2/5 to 0/5, and
  the modern forms spread beyond the sort — `slices.ContainsFunc` 4/5,
  `slices.Compact` 2/5.
- The card reaches a session through `go-code`'s routing line even when
  `go-style-core` is never loaded as a skill (2/5 sessions), so its cost is
  paid on every `go-code` task on this model.
- Scope is the residual risk: one card session modernized the neighbor it was
  told to leave, and nothing in the report caught it but the `go fix` column
  reading zero.

## Next

- An older-form column in `abrun` — the grep above over the fixture's
  production file — and a `neighbor_changed` flag for fixtures that declare a
  file off limits, so the `legacy.go` rewrite reads as a finding and not as a
  clean `go fix` count.
- Repeat with `-effort` on a model that takes it, to see whether the 2/5 of
  the rule-only arm is Haiku's ceiling or its default effort.
