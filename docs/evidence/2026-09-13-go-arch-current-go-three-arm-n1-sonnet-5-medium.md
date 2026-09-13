# Architecture reference and the current-Go rule — both corpora, three arms, Sonnet 5 medium (n=1)

Two edits landed in the working tree on 2026-09-13 without a measurement:
`go-code-refactor` gained `references/ARCHITECTURE.md` (296 lines, the
package-scale smells, five target shapes, the staged move and a boundary
test) with a routing line, an "Architecture at Package Scale" section and
the boundary move in its High tier; and `go-style-core`'s "Write Current Go"
became normative — the module's `go` directive sets the idiom, an older
neighbor is not a reason to write the older form — routed from `go-code`,
`go-code-refactor`, `MODERNIZATION.md`, `PRINCIPLES.md`, `go-packages` and
the review checklist. This run is the smoke for both: one repetition per
fixture and arm, on the corpus each edit can reach — the refactor corpus for
the architecture text, the implementation corpus for the idiom rule — with an
unaided arm beside the before/after pair. One repetition supports no line
claim; what it can show is a correctness or lint regression, a reading
pattern, and the cost of the added text.

| Arm | Tree |
|---|---|
| `no-skill` | no plugin |
| `reference` | `git worktree` of `e77d572` (release 1.15.0) |
| `baseline` | the working tree: 1.15.0 plus the two edits |

## Run

- Finished: 2026-09-13 12:21 UTC (refactor), 2026-09-13 12:22 UTC (implement); the two ran side by side
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 3` each
- Corpora and fixtures: `refactor` — `dispatch`, `pricing`, `report`, `store`; `implement` — `catalog`, `feed`, `gateway`
- Arms: `no-skill`, `reference`, `baseline`; 1 repetition per fixture and arm; 12 + 9 = 21 sessions, 0 CLI errors
- `reference` plugin SHA-256 `c3e5df8395d78c9bd143d0102a1f179d2e2539f771fdb267c3a7371fef461a26`
- `baseline` plugin SHA-256 `27582bbf2a2bac2687cf2fe109b21e49f0b70c9a7c828325200be5a3ff2e3b9d`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2
- Refactor report: [`2026-09-13-go-refactor-architecture-current-go-n1-sonnet-5-medium.json`](2026-09-13-go-refactor-architecture-current-go-n1-sonnet-5-medium.json)
  (SHA-256 `f2cca1cccd9c0e82e2b3e0955e1fe6b4c4958af7efe940c078ce56b6ca96af57`);
  traces [`….traces.tar.gz`](2026-09-13-go-refactor-architecture-current-go-n1-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `fb715a935dd95aa987dc897b9dd3b762da200c1fa97577b6f268f96728cc3a5e`)
- Implement report: [`2026-09-13-go-implement-current-go-catalog-feed-gateway-n1-sonnet-5-medium.json`](2026-09-13-go-implement-current-go-catalog-feed-gateway-n1-sonnet-5-medium.json)
  (SHA-256 `c481355d19c2578399847c25be2b4504926839d3d02f3cfc74e42726a83bf47d`);
  traces [`….traces.tar.gz`](2026-09-13-go-implement-current-go-catalog-feed-gateway-n1-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `343242c96d6b7f2504d4d37a4e0853bdb3300651e884af7c9087ad6c68053845`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session in each
- Cost: refactor $0.17 no-skill, $0.81 reference, $0.84 baseline; implement $0.16, $0.86, $1.21; $4.05 in all

```bash
go run ./cmd/abrun -corpus refactor -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of e77d572> -arms no-skill,reference,baseline \
  -tasks dispatch,pricing,report,store -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-refactor-architecture-current-go-n1-sonnet-5-medium.json
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of e77d572> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway -n 1 -j 3 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-implement-current-go-catalog-feed-gateway-n1-sonnet-5-medium.json
```

No session had a shell tool; the plugin's edit hook ran gofmt, vet,
`go fix -diff`, the package tests and golangci-lint after each edit in the
two skilled arms, and every skilled session loaded `go-code-refactor` (refactor)
or `go-code` (implement).

## Refactor corpus

| Arm | Δlines | Δfuncs | Line gate | Golden | Lint clean | $/run |
|---|---|---|---|---|---|---|
| `no-skill` | −7.0 | +1.50 | 2/4 | 4/4 | 2/4 | 0.043 |
| `reference` | −14.2 | +1.00 | 3/4 | 4/4 | 3/4 | 0.203 |
| `baseline` | −11.5 | +0.75 | 3/4 | 4/4 | 3/4 | 0.209 |

| Fixture | `no-skill` | `reference` | `baseline` |
|---|---|---|---|
| `dispatch` | +4, 3 funcs | −11 | −10 |
| `pricing` | −33 | −39, 1 func | −43 |
| `report` | +12, 3 funcs | +8, 3 funcs | +13, 3 funcs |
| `store` | −11 | −15 | −6 |

Correctness and lint do not move: golden 4/4 in every arm, lint-clean 3/4 in
both skilled arms against 2/4 unaided, `go fix` left nothing pending in any
session. The 2.7-line corpus gap is two sessions pulling opposite ways.
`store` −15 against −6: the reference session deleted the `if m.cache != nil`
and `if m.remote != nil` guards around the lookups as redundant (a nil-map
read returns `ok == false`), the baseline session kept them and flattened
only the miss path; both preserve behavior, and the choice is not one the
new text speaks to. `report` +8 against +13: both arms extracted the same
three helpers (`header`/`rowLine`/`totalLine` against
`writeHeader`/`writeRow`/`writeTotal`) and both failed the line gate, as
every skilled Sonnet session on this fixture has; the five lines are helper
layout. `pricing` moved the other way, −43 against −39. `dispatch` is level.

**Reading.** Every skilled session in both arms read all ten
`go-code-refactor` references named in Resource Routing, whatever their "Read
when" clause says; the baseline arm therefore read `ARCHITECTURE.md` in 4/4
sessions on four single-package fixtures where no package-scale smell exists
(the file is about 7.3K tokens at 2.6 chars/token). The `store` baseline
session also loaded `go-style-core` and its nine references, which the
reference session did not. The cost difference is +3% ($0.209 against
$0.203 a session), so the reads are cheap, but the routing condition did not
gate them on this model at this effort.

## Implementation corpus

| Arm | Δlines | Δfuncs | Golden | Lint clean | $/run |
|---|---|---|---|---|---|
| `no-skill` | +52.0 | 0 | 3/3 | 2/3 | 0.052 |
| `reference` | +32.0 (2 valid) | 0 | 2/3 | 2/2 | 0.285 |
| `baseline` | +48.3 | +1.67 | 3/3 | 3/3 | 0.405 |

| Fixture | `no-skill` | `reference` | `baseline` |
|---|---|---|---|
| `catalog` | +22 | +24 | +19 |
| `feed` | +47, `sort.Slice` | +40 | +40 |
| `gateway` | +87, `sort.Slice`, 3 lint findings | +69, **golden failed** | +86, 5 funcs, contract test file |

`gateway` is the fixture whose `HEAD` clause has flipped on every consecutive
day of Sonnet 5 runs: the reference session wrote no test file and failed the
golden test, the baseline session wrote `gateway_contract_test.go`, five
package-level handlers and helpers, registered explicit `HEAD` routes
returning 405, and passed at $0.70 against $0.29. At n=1 that is the recorded
base rate, not the edit. `feed` is identical at +40 in both skilled arms and
`catalog` is five lines shorter in baseline.

**The idiom rule had nothing to move.** Counting the older forms the rule
targets in every session's production files (`for i := 0; i < n`,
`interface{}`, `sort.Slice`/`sort.Strings`, `x := x`, `errors.As(`,
`WriteString(fmt.Sprintf`): the unaided arm wrote `sort.Slice` in 2/3
sessions (`feed`, `gateway`); the reference and baseline arms wrote none, and
both used `slices.` in `feed` (3 calls) and `gateway` (1). The reference
tree already produced current Go on these fixtures through the reach-for
table and `go-data-structures`, so a difference between the skilled arms
could not appear here; the fixture that would show it needs an older
neighbor in the same file, which none of the three has.

**Reading.** `MODERNIZATION.md` (about 5.7K tokens) was read in 3/3 baseline
sessions and 0/3 reference sessions — the new "Write Current Go" text links
to it for the behavior-changing swaps, and Sonnet 5 follows the link. Cost by
fixture: `catalog` $0.25 against $0.31, `feed` $0.26 against $0.25,
`gateway` $0.70 against $0.29; the arm total of $1.21 against $0.86 is the
`gateway` session's test file and hook-run tests, not the reference.

## What this run does and does not establish

- No correctness or lint regression from either edit: golden 4/4 and
  lint-clean 3/4 in both refactor arms; golden 3/3 against 2/3 and
  lint-clean 3/3 against 2/2 in the implementation arms.
- Lines are within one-repetition noise in both corpora; no line claim.
- Both edits add reading: Sonnet 5 medium reads every reference a loaded
  skill's Resource Routing names and every reference the skill text links,
  so `ARCHITECTURE.md` costs its full payload on every refactor session and
  `MODERNIZATION.md` on every implement session that loads `go-style-core`.
  The measured cost effect is +3% on the refactor corpus and not separable
  from `gateway` variance on the implement corpus.
- The idiom rule's effect is unmeasured, not absent: the reference arm was
  already current on these fixtures. A fixture that measures it needs a file
  with an older neighbor beside the stub.

## Next

An n=5 two-arm run (`reference`, `baseline`) on `report` and `store` decides
whether the −14.2 against −11.5 reading is noise; a `-tasks` fixture carrying
an older neighbor is what would measure "Write Current Go".
