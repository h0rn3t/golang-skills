# Budget Closure — `gateway`, reference against baseline, Opus 5 medium (n=5)

The measurement behind the 2026-09-12 Declaration Budget edit. In the
2026-09-11 Opus 5 traces every skilled `gateway` session (6/6) wrote
`writeJSON` and `methodNotAllowed` as closures inside `NewServer` so the
report could say `added package-level declarations: 0`, and no unaided
session did. The edit says a function literal bound to a name that captures
nothing from its enclosing function counts under the same three rules and,
when it earns its place, is a small unexported function of the package; the
count is a record, not a score. This run asks one question: did the helpers
move back out, and at what cost in lines, correctness, lint, and money.

Both arms are skilled. The `reference` arm is the committed tree at `2fab3d6`
(release 1.12.0); `baseline` is the working tree with the 2026-09-12 edits,
the same tree the [five-fixture smoke](2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.md)
describes. There is no control here: the unaided model's `gateway` rate on
this model is on record at
[1/3](2026-09-11-go-implement-newcode-workflow-opus-5-medium.md) and
[0/1](2026-09-12-go-implement-budget-closure-smoke-opus-5-medium.md), and the
effort sweep of the same day repeats it.

## Run

- Finished: 2026-09-12 08:39 UTC
- Runner: `claude` 2.1.267
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`; fixture: `gateway`
- Arms: `reference`, `baseline`; 5 repetitions each, 10 sessions total
- `reference`: a `git worktree` of `2fab3d6`, plugin SHA-256
  `778830a584281cc9fcd6949cabaf2c9596ee73f581e0e22b670832ab9acffe8f`
- `baseline`: the working tree at `2fab3d6` plus the uncommitted edits,
  plugin SHA-256 `ce495ac2088148f18754c0f99e60095443491440344f8551bcc4e4657416fbf6`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2 for the lint line
- Raw report: [`2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.json`](2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.json)
  (SHA-256 `be3d77f495179bff27e564c163dc4b27488b14e19d257f8a5c90f1a424df6b93`)
- Session transcripts: [`2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.traces.tar.gz`](2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.traces.tar.gz)
  (SHA-256 `2e1c4ab6e084b09b97a2713c69094420c64bd7799bb772ee189ad93e71a44a94`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $4.7871 reference, $4.8483 baseline — $0.957 against $0.970 a session

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of 2fab3d6> -arms reference,baseline \
  -tasks gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-budget-closure-gateway-n5-opus-5-medium.json
```

All 10 sessions completed without a CLI error, changed the fixture, stayed out
of the repository checkout, and passed the hidden golden test. The tool set
had no shell. `go-code` fired first in every session; every session loaded
`go-http`, `go-testing`, `go-error-handling` and `go-style-core`, in two or
three messages.

## Results

| Arm | rep | Golden | Δlines | Δfuncs | Δclos | helpers, as written | lint after | `Skill` msgs | API calls | $ / run |
|---|---|---|---:|---:|---:|---|---:|---:|---:|---:|
| `reference` | 0 | 1/1 | 80 | 0 | 2 | `writeJSON`, `methodNotAllowed` closures | 2 | 2 | 7 | 0.856 |
| `reference` | 1 | 1/1 | 73 | 0 | 2 | `writeJSON`, `methodNotAllowed` closures | 0 | 2 | 9 | 0.921 |
| `reference` | 2 | 1/1 | 76 | 0 | 2 | `writeJSON`, `methodNotAllowed` closures | 0 | 3 | 14 | 1.213 |
| `reference` | 3 | 1/1 | 79 | 0 | 1 | `writeJSON` closure | 2 | 2 | 7 | 0.815 |
| `reference` | 4 | 1/1 | 78 | 0 | 2 | `writeJSON`, `methodNotAllowed` closures | 2 | 2 | 11 | 0.982 |
| `baseline` | 0 | 1/1 | 84 | 2 | 0 | `requireGET`, `writeJSON` functions | 0 | 2 | 7 | 0.933 |
| `baseline` | 1 | 1/1 | 85 | 2 | 0 | `writeJSON`, `methodNotAllowed` functions | 2 | 3 | 8 | 0.923 |
| `baseline` | 2 | 1/1 | 92 | 2 | 0 | `getOnly`, `writeJSON` functions | 0 | 2 | 8 | 0.905 |
| `baseline` | 3 | 1/1 | 85 | 2 | 0 | `getOnly`, `writeJSON` functions | 2 | 2 | 11 | 1.058 |
| `baseline` | 4 | 1/1 | 75 | 2 | 0 | `headNotAllowed`, `writeJSON` functions | 0 | 3 | 11 | 1.029 |

| Arm | golden | Δlines | Δfuncs | Δclos | budget line says | lint after, mean | lint clean | `Skill` msgs | API calls | cache writes | cache reads | output tokens | thinking tokens | median report chars | $ / run |
|---|---|---:|---:|---:|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `reference` | 5/5 | 73, 78, 80, 76, 79 → 77.2 | 0.00 | 1.80 | `0` in 5/5, each naming the closures | 1.20 | 2/5 | 2.2 | 9.6 | 47.6K | 335K | 12517 | 6041 | 1463 | 0.957 |
| `baseline` | 5/5 | 92, 75, 84, 85, 85 → 84.2 | 2.00 | 0.00 | `2` in 5/5, each naming both call sites | 0.80 | 3/5 | 2.4 | 9.0 | 48.1K | 299K | 13544 | 7235 | 1669 | 0.970 |

Lines: +7.0 baseline over reference, exact permutation p = 0.048. Golden:
5/5 against 5/5. The lint findings are the same two in every dirty session,
`errcheck` on `w.Write` in the health handler and in `writeJSON`.

## Reading

- **The helpers moved out, 5/5.** Every reference session hid `writeJSON`
  (and in 4/5 `methodNotAllowed`) in a closure and reported `0`; every
  baseline session wrote the same two helpers as unexported package functions,
  counted `2`, and named both call sites — `writeJSON: the list and
  single-account handlers; getOnly: the /healthz, /accounts, /accounts/{id}
  registrations`. `Δclos` is 9 against 0. The 2026-09-11 artifact is gone on
  this fixture at n=5.
- **It costs seven lines.** A package-level function carries a doc comment, a
  signature line, and blank lines around it that a closure does not; two of
  them are the +7.0 (p = 0.048). That is the honest price of the shape a
  reviewer accepts, and it is the number the 2026-09-11 arm was avoiding by
  moving the helpers rather than removing them. `Δlines` and `Δclos` have to
  be read together from here on; the corpus README says so.
- **Correctness is level and every session passed**, including the class
  clause the reference smoke session missed (`POST /accounts` → 307): no
  session in either arm registered a redirecting pattern here.
- **The unchecked write is not fixed by either tree.** 3/5 reference and 2/5
  baseline sessions wrote `w.Write(body)` bare; the other sessions wrote
  `_, _ = w.Write(body)` with a reason. Without a shell the gate never runs,
  so nothing in the session catches it; the lint line is the only place it
  shows. Whether `go-http` should carry the write-error line as a rule is a
  separate decision this run does not make.
- **Cost, loading, and report shape are level**: $0.96 against $0.97 a
  session, 2.2 against 2.4 `Skill` messages, every session carrying the
  `checks:` line and the budget line; median report 1463 against 1669
  characters.

## What this changes

- The Declaration Budget closure rule ships as measured: the budget count now
  records the helpers instead of hiding them, at +7 lines on `gateway` and no
  change in correctness or cost.
- The README's Opus 5 new-code cell stays on the three-arm n=3 run, which
  has a control; this pair qualifies it and does not set it.
