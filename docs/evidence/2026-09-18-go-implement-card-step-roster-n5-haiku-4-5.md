# The idiom card as a workflow step on `roster` — Haiku 4.5, two runs at n=5

The [2026-09-13 run](2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.md)
showed Haiku 4.5 is the model on which `roster` discriminates: unaided it
copies `sort.Strings` from the pre-1.21 neighbor in 5/5 sessions, the
normative rule alone leaves it in 2/5, and the idiom card takes it to 0/5.
These two runs measure the [same wording change as the Sonnet 5 pair of the
day](2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md):
the card read moved from a Resource Routing bullet into `go-code`'s workflow,
first as a numbered step of its own (run 1), then folded into step 2 (run 2),
each against release 1.20.1.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `89d702a` (release 1.20.1): the card named in a Resource Routing bullet of `go-style-core` and `go-code` and in "Write Current Go" |
| `baseline`, run 1 | the working tree with the card as step 3 of an eight-step workflow, plus the edit-hook record and the other edits of the change (plugin SHA-256 `868eac2c…`) |
| `baseline`, run 2 | the working tree with the card as the first clause of step 2, six steps as in 1.20.1 (plugin SHA-256 `bfcfed52…`) |

## Runs

- Run 1 finished 2026-09-18 16:45 UTC, seed `1`; run 2 finished 16:52 UTC, seed `2`
- Runner: `claude` 2.1.267; model `claude-haiku-4-5-20251001`, no reasoning-effort flag
- Corpus `implement`, fixture `roster`, arms `reference`, `baseline`, 5 repetitions per arm, 10 sessions a run, `-j 3`, 0 CLI errors
- `reference` plugin SHA-256 `d478dd31988660fcb8934eba8e56e7e378221b321b278073006d4ecb4af19ca0`; `baseline` run 1 `868eac2c01f7dfd350c56ada023dba7ac78e5dca37c709c17792ccec169b2466`, run 2 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0`
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook ran in every session, no session had a shell tool
- Run 1 report: [`2026-09-18-go-implement-card-step-roster-n5-haiku-4-5.json`](2026-09-18-go-implement-card-step-roster-n5-haiku-4-5.json) (SHA-256 `eefac1da5d0f1c3ba8e5e4cecf84e52d9cb0488d95cabcebd1f1898ce2aeb6bc`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-step-roster-n5-haiku-4-5.traces.tar.gz) (SHA-256 `78064631216afe8b80c96264f58bb775742ebab47d588350cb05a31def370215`); $1.58
- Run 2 report: [`2026-09-18-go-implement-card-in-step2-roster-n5-haiku-4-5.json`](2026-09-18-go-implement-card-in-step2-roster-n5-haiku-4-5.json) (SHA-256 `fff384c9e4d011cbefc00f0af886c64107700c084dd7f8537a68591005e624e6`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-in-step2-roster-n5-haiku-4-5.traces.tar.gz) (SHA-256 `cf0324030de1f86a03d0596641e9986c5b19c1ac27df631998282f49ae2d40aa`); $1.21

```bash
go run ./cmd/abrun -corpus implement -tasks roster -runner claude -model claude-haiku-4-5-20251001 \
  -reference-root <worktree of 89d702a> -arms reference,baseline \
  -n 5 -j 3 -seed 1 -timeout 10m -keep -verbose -out <run 1 json>     # seed 2 for run 2
```

## The reading

The reading is the 2026-09-13 one: a grep over each session's `roster.go` for
the neighbor's older forms (`sort.Strings`/`sort.Slice`, `for i := 0; i <`,
`strings.Index(`, `interface{}`), the card `Read` from the trace, and the
routing events (Skill turns, gate blocks, edit-hook findings) counted as in
the Sonnet report. `go fix` pending stayed at the neighbor's three hunks in
all 20 sessions: no session rewrote `legacy.go`, the file the fixture declares
off limits, where the 2026-09-13 card arm did so once in five.

| Run | Arm | Golden | Older form in `roster.go` | `slices.Sort` | `slices.ContainsFunc` | Card `Read` | Skill turns/s | Gate blocks | Hook findings | Turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|
| 1 | `reference` | 5/5 | 0/5 | 5/5 | 0/5 | 2/5 | 2.40 | 0 | 24 | 15.0 | 0.145 |
| 1 | `baseline` (own step) | 5/5 | **1/5** | 4/5 | 2/5 | 3/5 | 2.60 | 2 | 35 | 18.4 | 0.171 |
| 2 | `reference` | 5/5 | 1/5 | 4/5 | 2/5 | 3/5 | 2.40 | 1 | 20 | 13.6 | 0.123 |
| 2 | `baseline` (in step 2) | 5/5 | **0/5** | 5/5 | 2/5 | 4/5 | 2.40 | 0 | 21 | 13.4 | 0.119 |

Per session, run 1:

| Arm | Rep | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Older form | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 0 | +20 | go-code, go-style-core | 2 | yes | 0 | 4 | no | 14 | 0.116 |
| `reference` | 1 | +30 | go-code, go-style-core, go-testing, go-data-structures | 4 | yes | 0 | 5 | no | 17 | 0.158 |
| `reference` | 2 | +25 | go-code, go-style-core, go-testing, go-data-structures | 2 | no | 0 | 5 | no | 13 | 0.116 |
| `reference` | 3 | +22 | go-code, go-style-core, go-testing, go-data-structures | 2 | no | 0 | 4 | no | 12 | 0.141 |
| `reference` | 4 | +25 | go-code, go-style-core, go-data-structures, go-testing | 2 | no | 0 | 6 | no | 19 | 0.192 |
| `baseline` | 0 | +17 | go-code, go-style-core, go-testing | 2 | yes | 0 | 10 | no | 18 | 0.178 |
| `baseline` | 1 | +34 | go-code, go-style-core, go-testing, go-data-structures, go-linting | 3 | no | 0 | 5 | no | 18 | 0.160 |
| `baseline` | 2 | +28 | go-code, go-style-core, go-testing | 3 | yes | 1 | 9 | no | 24 | 0.223 |
| `baseline` | 3 | +25 | go-code, go-style-core, go-data-structures, go-testing | 3 | no | 1 | 6 | **`sort.Strings`** | 18 | 0.159 |
| `baseline` | 4 | +24 | go-code, go-style-core, go-testing, go-interfaces | 2 | yes | 0 | 5 | no | 14 | 0.135 |

Per session, run 2:

| Arm | Rep | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Older form | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | 0 | +18 | go-code, go-style-core, go-testing, go-linting | 4 | yes | 1 | 2 | no | 19 | 0.199 |
| `reference` | 1 | +17 | go-code, go-style-core | 2 | yes | 0 | 4 | no | 11 | 0.094 |
| `reference` | 2 | +22 | go-code, go-style-core | 2 | no | 0 | 4 | **`sort.Strings`** | 12 | 0.084 |
| `reference` | 3 | +22 | go-code, go-style-core, go-testing | 2 | no | 0 | 5 | no | 14 | 0.118 |
| `reference` | 4 | +22 | go-code, go-style-core, go-testing | 2 | yes | 0 | 5 | no | 12 | 0.119 |
| `baseline` | 0 | +22 | go-code, go-style-core, go-testing, go-data-structures | 2 | yes | 0 | 5 | no | 15 | 0.135 |
| `baseline` | 1 | +22 | go-code, go-style-core, go-data-structures | 2 | no | 0 | 5 | no | 9 | 0.089 |
| `baseline` | 2 | +19 | go-code, go-style-core, go-data-structures | 3 | yes | 0 | 1 | no | 14 | 0.103 |
| `baseline` | 3 | +27 | go-code, go-style-core, go-data-structures, go-testing | 3 | yes | 0 | 5 | no | 13 | 0.131 |
| `baseline` | 4 | +25 | go-code, go-style-core, go-interfaces, go-testing | 2 | yes | 0 | 5 | no | 16 | 0.138 |

Haiku reads the card where Sonnet 5 does not: 5/10 reference sessions, 3/5
with the card as its own step, 4/5 with it in step 2. The two sessions that
wrote `sort.Strings` — baseline r3 of run 1, reference r2 of run 2 — are two
of the eight that did not read it; no session that read the card wrote an
older form, in either arm of either run (12/12). The 1.20.1 text is not at
zero on this model as it was on 2026-09-13: one older form in ten reference
sessions, against 2/5 for the 1.16.0 rule-only text then.

The step of its own cost what it cost on Sonnet 5: two gate blocks against
none, 18.4 turns against 15.0, +18% a session ($0.171 against $0.145), with
the card read in one more session. Folded into step 2 the columns are level
or better — 4/5 reads, 0/5 older forms, no gate block, 13.4 turns against
13.6, $0.119 against $0.123 — at a sample where one session is the whole
difference.

## What this establishes

- On Haiku 4.5 the card read tracks the older form exactly (read 12, older
  form 0; not read 8, older form 2), so the read rate is the number to move.
  The wording moved it from 5/10 to 4/5 at best, one session's worth at n=5.
- The bundled step 2 is the wording to keep: no column is worse than 1.20.1
  and the routing shape is unchanged. The separate step is the wording to
  drop, on this model as on Sonnet 5.
- Scope held in 20/20 sessions: `legacy.go` untouched, `go fix` pending at
  the neighbor's three. The 2026-09-13 residual risk did not recur here.

## Next

- The read rate on Haiku is 4/5 at best from text; the host routes named in
  the Sonnet report are the way to 5/5, and `roster` on Haiku at n=5 is the
  cheap place to measure each ($1.2 a run).
