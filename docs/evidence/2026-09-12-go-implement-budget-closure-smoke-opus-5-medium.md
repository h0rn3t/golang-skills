# Budget Closure Smoke — five fixtures, three arms, Opus 5 medium (n=1)

A first look at the 2026-09-12 edits on the model whose traces motivated
them, and the first run of the two new fixtures. One repetition per fixture
and arm supports no claim about size, correctness rate, or cost, and none is
made below. What this run reads is whether the closure rule changed the shape
of the `gateway` helpers, whether the two new harness columns see what the
traces showed, and whether `pool` and `fetch` are live.

The edits in the `baseline` arm:

- **`go-code` Declaration Budget**: a function literal bound to a name that
  captures nothing from its enclosing function counts under the same three
  rules as a package-level declaration and, when it earns its place, is
  written as a small unexported function. The count is described as a
  record, not a score.
- **`go-code`** moves the worked Contract Table and the `Manifest` body to
  `references/NEW-CODE-EXAMPLES.md` and the Route Before The First Edit table
  behind the gate, so Workflow, Writing New Code, and Close With The Gate end
  inside the first 5K tokens.
- **`go-error-handling`** drops its Core Rules to `references/ERROR-TYPES.md`;
  **`go-naming`** drops the convention restatements to
  `references/IDENTIFIERS.md`.
- **`abrun`** prints `Δclos` (named function literals inside bodies) and a
  lint-findings line (bundled `golangci.yml`, production files, before and
  after) for every run.
- **Fixtures** `pool` (go-concurrency) and `fetch` (go-resilience), see
  [`evals/ab/_implement/README.md`](../../evals/ab/_implement/README.md).

The `reference` arm is the committed tree at `2fab3d6` (release 1.12.0).

## Run

- Finished: 2026-09-12 08:31 UTC
- Runner: `claude` 2.1.267
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`
- Fixtures: `catalog`, `feed`, `gateway`, `pool`, `fetch`
- Arms: `no-skill`, `reference`, `baseline`
- Repetitions: 1 per fixture and arm, 15 sessions total
- `reference`: a `git worktree` of `2fab3d6`, plugin SHA-256
  `778830a584281cc9fcd6949cabaf2c9596ee73f581e0e22b670832ab9acffe8f`
- `baseline`: the working tree at `2fab3d6` plus the uncommitted edits above,
  plugin SHA-256 `ce495ac2088148f18754c0f99e60095443491440344f8551bcc4e4657416fbf6`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2 for the lint line
- Raw report: [`2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.json`](2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.json)
  (SHA-256 `f0d214548ea3446e323459a8baa8af30698ef79906c364ee4d06cd25289e605a`)
- Session transcripts: [`2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.traces.tar.gz`](2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.traces.tar.gz)
  (SHA-256 `52dc954d4e2a2cef04bf75535981f5b84acb02c19fd871c5c97339e5dc3bfdae`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $1.2808 control, $5.0149 reference, $4.3442 baseline — **3.92x** and **3.39x** the control

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of 2fab3d6> -arms no-skill,reference,baseline \
  -tasks catalog,feed,gateway,pool,fetch -n 1 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.json
```

All 15 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell.

`go-code` fired in 5/5 reference and 5/5 baseline sessions, as the first tool
call in every one of them.

## Results

| Fixture | Arm | Golden | Δlines | Δtypes | Δfuncs | Δclos | Δbranch | lint before→after | `Skill` msgs | API calls | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `catalog` | no-skill | 1/1 | 23 | 0 | 0 | 0 | 4 | 0→0 | — | 7 | 0.082 |
| `catalog` | reference | 1/1 | 20 | 0 | 0 | 0 | 4 | 0→0 | 2 | 7 | 0.538 |
| `catalog` | baseline | 1/1 | 22 | 0 | 0 | 0 | 4 | 0→0 | 2 | 8 | 0.589 |
| `feed` | no-skill | 1/1 | 46 | 2 | 0 | 0 | 4 | 0→0 | — | 7 | 0.087 |
| `feed` | reference | 1/1 | 35 | 0 | 0 | 0 | 3 | 0→0 | 2 | 7 | 0.630 |
| `feed` | baseline | 1/1 | 52 | 0 | 0 | 0 | 4 | 0→0 | 2 | 7 | 0.657 |
| `gateway` | no-skill | **0/1** | 109 | 0 | 3 | 0 | 8 | 0→1 | — | 12 | 0.427 |
| `gateway` | reference | **0/1** | 86 | 0 | 0 | **2** | 7 | 0→0 | 3 | 11 | 1.193 |
| `gateway` | baseline | 1/1 | 84 | 0 | 2 | 0 | 7 | 0→0 | 2 | 9 | 0.868 |
| `pool` | no-skill | 1/1 | 65 | 0 | 0 | 0 | 11 | 0→2 | — | 7 | 0.157 |
| `pool` | reference | 1/1 | 47 | 0 | 0 | 0 | 8 | 0→0 | 3 | 11 | 1.147 |
| `pool` | baseline | 1/1 (own test failed) | 50 | 0 | 0 | 0 | 10 | 0→0 | 3 | 10 | 0.945 |
| `fetch` | no-skill | **0/1** | 224 | 0 | 10 | 0 | 37 | 0→4 | — | 16 | 0.527 |
| `fetch` | reference | 1/1 | 117 | 0 | 2 | 0 | 23 | 0→1 | 4 | 13 | 1.508 |
| `fetch` | baseline | **0/1** | 124 | 0 | 1 | 0 | 23 | 0→3 | 3 | 10 | 1.285 |

| Arm | golden | valid | Δlines | Δfuncs | Δclos | lint after, mean | lint clean | `Skill` msgs | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `no-skill` | 3/5 | 3 | 44.7 | 0.00 | 0.00 | 0.67 | 2/3 | 0.0 | 9.8 | 7.8K | 88K | 5318 | 1427 | 1355 | 0.256 |
| `reference` | 4/5 | 4 | 54.8 | 0.50 | 0.00 | 0.25 | 3/4 | 2.8 | 9.8 | 47.3K | 336K | 14410 | 8274 | 1620 | 1.003 |
| `baseline` | 4/5 | 3 | 52.7 | 0.67 | 0.00 | 0.00 | 3/3 | 2.4 | 8.8 | 43.8K | 272K | 11751 | 5696 | 1303 | 0.869 |

Means are over valid runs (golden, build, and the model's own tests all
passing); the `baseline` `pool` session is excluded from them for its own
test, see below.

Skill loads — reference: `catalog` 5 owners in 2 messages; `feed` 6 in 2;
`gateway` `go-code`, `go-style-core`, `go-http`, `go-error-handling`,
`go-testing`, `go-security`, `go-defensive`, `go-data-structures` in 3;
`pool` `go-code`, `go-style-core`, `go-concurrency`, `go-context`,
`go-error-handling`, `go-testing`, `go-defensive` in 3; `fetch` `go-code`,
`go-style-core`, `go-http`, `go-error-handling`, `go-resilience`,
`go-context`, `go-testing`, `go-defensive` in 4.
Baseline: `catalog` 5 in 2; `feed` 6 in 2; `gateway` `go-code`,
`go-style-core`, `go-http`, `go-testing`, `go-error-handling` in 2; `pool` the
same 7 as reference in 3; `fetch` reference's 8 plus `go-security` in 3.

Budget lines: `reference gateway` reported `added package-level declarations:
0` with `writeJSON` and `methodNotAllowed` as closures inside `NewServer`;
`baseline gateway` reported `2 — getOnly: /healthz, /accounts, /accounts/{id}
routes; writeJSON: list handler, single-account handler`, both package-level
functions. `fetch` reported 2 in both skilled arms (`send` shared by
`GetOrder` and `PlaceOrder`, plus a `Retry-After` parser in reference and a
`defaultHTTP` client in baseline). Every other skilled session reported 0.

## Reading

- **The closure rule acted on `gateway`, at n=1.** The reference session is
  the 2026-09-11 shape exactly: two capture-nothing closures, budget line
  `0`, `Δclos +2`. The baseline session wrote the same two helpers as
  package-level functions, counted them, named both call sites, and passed
  the golden test. The reference session did not: it registered
  `/accounts/{$}` as a 404 and `ServeMux` answered `POST /accounts` with a
  307 redirect to it — a default leaking through the "any other method" class
  clause, the kind the Contract Table names. One session each; the `gateway`
  n=5 pair measures the rule.
- **The unaided `gateway` session answered `HEAD` with 200** on every route
  and left one `errcheck` finding, at 109 lines with three helpers — the same
  miss as on 2026-09-11. Both skilled arms handled the write
  (`_, _ = w.Write(body)` with a reason) and are lint-clean.
- **`fetch` is live.** The unaided session wrote 224 lines, ten helpers and
  four lint findings, and its error message lost the status (`partner
  returned ,`), which the golden test reads as the id-and-status clause
  failing. The reference session passed at 117 lines. The baseline session
  sent `PlaceOrder` once and honored `Retry-After`, but shaved its first
  wait to 110 ms with jitter in `[wait/2, wait]` against a documented
  `Pause` of 200 ms. The fixture's doc comment then said "Pause is the wait
  before the second send"; after this run it says "the least wait", and the
  golden test was rewritten to accept jitter on top of a growing schedule
  while still failing a constant pause or one below `Pause`. The baseline
  session's code still fails the revised test on its first wait, so its cell
  above is unchanged; the run predates the revision and is not evidence
  about it.
- **`pool` passed golden 3/3.** The unaided session took 65 lines, two lint
  findings and two pending `go fix` rewrites; the skilled arms 47 and 50,
  clean. The baseline session's own contract test panicked with `wait
  already in progress`: it called `synctest.Sleep` from inside every job
  goroutine, and `synctest.Sleep` calls `synctest.Wait`, which two goroutines
  may not reach at once. `go-testing`'s table row that recommends
  `synctest.Sleep` now says it belongs in the test goroutine and that a
  goroutine the code under test starts sleeps with `time.Sleep`. The
  production code passed the hidden test; the run is excluded from the means
  by the model-test rule.
- **The new columns see what the traces showed.** `Δclos` is non-zero in
  exactly the session that hid its helpers; the lint line separates the three
  unaided sessions that left findings (1, 2, 4) from the skilled sessions,
  and the one skilled finding (`reference fetch`, 1) from a clean tree.
- **Size and cost, one repetition:** `catalog` 20 → 22, `feed` 35 → 52,
  `gateway` 86 → 84, `pool` 47 → 50, `fetch` 117 → 124 reference to baseline;
  $1.00 to $0.87 a session, `Skill` messages 2.8 to 2.4, cache reads 336K to
  272K. No claim.

## What this changes

- The closure rule goes to the `gateway` n=5 pair before it is read as a
  result.
- `pool` and `fetch` stay unadmitted until a repeated run shows `no-skill`
  measurably worse than `baseline`; this run puts the unaided arm at 0/1 on
  `fetch` and lint-dirty on both.
- `go-testing` carries the `synctest.Sleep` caveat.
- The README table is unchanged: n=1 does not displace the n=3 Opus 5 cell.
