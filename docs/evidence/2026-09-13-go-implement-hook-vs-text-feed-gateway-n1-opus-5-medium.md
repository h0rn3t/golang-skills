# Edit hook against skill text — `feed` and `gateway`, three arms, Opus 5 medium (n=1)

The 2026-09-12 afternoon measurement put four changes in one arm: the edit
hook that runs the package's tests and golangci-lint after every `.go` edit,
and three skill sentences (`go-http`'s open-discard bullet, the nil-built
empty case in `go-code`'s Contract Table, the safe-form-only standard-library
bullet). Nothing there attributes a reading to the hook or to the text. This
run separates them into three arms on the same day, model and effort, at one
repetition per fixture and arm:

| Arm | Tree | Hook | Skill text |
|---|---|---|---|
| `1.14.0` | `git worktree` of `d1b7124` | gofmt, vet, `go fix -diff` | release 1.14.0 |
| `hook-only` | `d1b7124` with `hooks/go-vet-on-edit.sh` and `hooks/hooks.json` copied from the working tree | + `go test`, golangci-lint | release 1.14.0 |
| `tree` | the working tree | + `go test`, golangci-lint | the four sentences, including the step-6 checks-line wording added after the 2026-09-12 runs |

`abrun` takes one `-reference-root`, so the three arms are two runs whose
`baseline` (`tree`) is the same plugin digest: run A is `1.14.0` against
`tree`, run B is `hook-only` against `tree`. `tree` therefore has two
sessions per fixture, the other arms one.

## Run

- Finished: 2026-09-13 08:25 UTC (A), 08:24 UTC (B); the two ran side by side
- Runner: `claude` 2.1.267
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 4` each
- Corpus: `implement`; fixtures: `feed`, `gateway`
- Arms per run: `reference`, `baseline`; 1 repetition per fixture and arm, 4
  sessions a run, 8 in all
- `1.14.0`: plugin SHA-256 `0d360ceb91aaddfc327a72e0c79bf2b911ba5b0e1fd759ceadf8c227a84e52fe`
- `hook-only`: plugin SHA-256 `907894f3bd15b59090d4b8d8a45b800e57c36c7905c6fb4ecefbe84ef37faa9e`
- `tree`: plugin SHA-256 `503d79cc6e8e3a4a8ab192a83743937032727d818fd369f07dbd470139a20d04`
  (the 2026-09-12 runs measured `c59720d3…`; the difference is the step-6
  wording and the `.claude-plugin` version bump)
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2
- Run A report: [`2026-09-13-go-implement-hook-vs-text-ref-1140-feed-gateway-n1-opus-5-medium.json`](2026-09-13-go-implement-hook-vs-text-ref-1140-feed-gateway-n1-opus-5-medium.json)
  (SHA-256 `84c895fe5d559a938cf05387e29d1b13ad654db41573b53bb1978c93ee9e663e`);
  traces [`….traces.tar.gz`](2026-09-13-go-implement-hook-vs-text-ref-1140-feed-gateway-n1-opus-5-medium.traces.tar.gz)
  (SHA-256 `b57c6586a1d65a909d9d5fd190fa582438a8aa3fe6694ea0c80f4a151c1d373b`)
- Run B report: [`2026-09-13-go-implement-hook-vs-text-ref-hook-only-feed-gateway-n1-opus-5-medium.json`](2026-09-13-go-implement-hook-vs-text-ref-hook-only-feed-gateway-n1-opus-5-medium.json)
  (SHA-256 `efbb67d22445a9ff1564731d601b2816fb12f5bcc6eefd4178d6b836e1f10964`);
  traces [`….traces.tar.gz`](2026-09-13-go-implement-hook-vs-text-ref-hook-only-feed-gateway-n1-opus-5-medium.traces.tar.gz)
  (SHA-256 `7805c51c2bec008fbe85b2e1b2f54f3e7b058bce2fa918ee865b5c007e8ba7b5`),
  one `traces/<arm>-<fixture>-r0.jsonl` per session in each
- Cost: A $2.06 reference, $1.81 baseline; B $1.88 reference, $1.64 baseline;
  $7.39 in all

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of d1b7124> -arms reference,baseline \
  -tasks feed,gateway -n 1 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-implement-hook-vs-text-ref-1140-feed-gateway-n1-opus-5-medium.json
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of d1b7124 + working-tree hooks/> -arms reference,baseline \
  -tasks feed,gateway -n 1 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-13-go-implement-hook-vs-text-ref-hook-only-feed-gateway-n1-opus-5-medium.json
```

All 8 sessions completed without a CLI error, changed the fixture, stayed out
of the repository checkout, passed the hidden golden test, and were
lint-clean after. The tool set had no shell. `go-code` fired first in every
session; the `tree` `feed` session of run A loaded its four owners through
`Read` rather than `Skill`, so the harness lists it as `skills=[go-code]`.

## Results

| Arm | Fixture | Golden | Δlines | Δfuncs | Δbcom | discard reasons | edits | `go test` sections | `golangci-lint` sections | lint after | $ |
|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `1.14.0` | `feed` | 1/1 | 43 | 0 | 2 | — | 4 | — | — | 0 | 0.674 |
| `hook-only` | `feed` | 1/1 | 39 | 0 | 2 | — | 4 | 2 | 0 | 0 | 0.657 |
| `tree` (A) | `feed` | 1/1 | 33 | 0 | 1 | — | 4 | 2 | 2 | 0 | 0.732 |
| `tree` (B) | `feed` | 1/1 | 38 | 0 | 2 | — | 5 | 2 | 0 | 0 | 0.676 |
| `1.14.0` | `gateway` | 1/1 | 67 | 1 | 5 | 2 | 9 | — | — | 0 | 1.385 |
| `hook-only` | `gateway` | 1/1 | 72 | 2 | 6 | 2 | 7 | 0 | 10 | 0 | 1.222 |
| `tree` (A) | `gateway` | 1/1 | 76 | 2 | 5 | 2 | 10 | 12 | 10 | 0 | 1.075 |
| `tree` (B) | `gateway` | 1/1 | 72 | 2 | 3 | 2 | 7 | 4 | 10 | 0 | 0.966 |

A section is one hook message carrying a failing `go test` or a lint finding
in the edited file. The `1.14.0` hook has neither. Every `gateway` session in
every arm wrote both writes as `_, _ = w.Write(…)` or `io.WriteString` with
a reason on the line; none was bare.

Per arm, over both fixtures:

| Arm | sessions | golden | lint clean | Δlines | Δbcom | $ / session |
|---|---:|---:|---:|---:|---:|---:|
| `1.14.0` | 2 | 2/2 | 2/2 | 55.0 | 3.5 | 1.030 |
| `hook-only` | 2 | 2/2 | 2/2 | 55.5 | 4.0 | 0.940 |
| `tree` | 4 | 4/4 | 4/4 | 54.75 | 2.75 | 0.862 |

### What the checks line said

The step-6 wording — the checks line carries what the hook reported, check
by check, and a hook's silence is not a result — was written after the
2026-09-12 runs and is measured here for the first time, in the `tree` arm
only.

| Arm | Fixture | checks line |
|---|---|---|
| `1.14.0` | `feed` | `unavailable (no shell)`, then a sentence that the hook's `go vet` compiled the package |
| `1.14.0` | `gateway` | `gofmt pass · vet pass (both from the repo's post-edit hook) · test unavailable (no shell) · lint skipped (no shell)` |
| `hook-only` | `feed` | `unavailable (no shell)`, then: the hook reported nothing on the final two edits, "I'm not treating that silence as a verified pass" |
| `hook-only` | `gateway` | `gofmt pass · vet pass · go fix -diff clean · test pass · lint pass (bundled golangci.yml, via the edit hook)` |
| `tree` (A) | `feed` | `gofmt pass (hook) · vet pass (hook) · test pass (hook) · lint pass (hook)` |
| `tree` (A) | `gateway` | the same four `(hook)` entries, "the edit hook's results on the final edit, not a separately run gate" |
| `tree` (B) | `feed` | `gofmt pass · vet pass · test pass · lint pass` "via the plugin's edit hook", then: the hook "reported nothing after the final edits" |
| `tree` (B) | `gateway` | `gofmt pass · vet pass · test pass · lint pass` "(edit hook, no shell in this session)" |

The `hook-only` arm, whose text still says `unavailable (no shell)`, split:
one session wrote the literal line and refused to read silence as a pass, the
other wrote four passes from the hook. The `tree` arm wrote the per-check
form in 4/4, three of them with the `(hook)` marker the text gives; the
fourth reported a pass and then named the hook's silence as its evidence,
which the sentence forbids. On this fixture the hook's silence on the last
edit was in fact a clean run, so the report was true and its reasoning was
not the text's.

### The empty case

The Contract Table sentence — the empty case is built from nil, not from an
empty literal — is visible in the `feed` contract tests. Both old-text
sessions wrote a `nil events` case and an `[]Event{}` case beside it; both
`tree` sessions wrote the nil case and no empty-literal case. All four
`gateway` sessions built the empty list from `NewServer(…, nil)` in every
arm. The `feed` code is 4 to 10 lines shorter in `tree` than in `1.14.0`
(33 and 38 against 43), the only size movement in the run, and at n=1 a
movement, not an effect.

## Reading

- **Correctness and lint saturate.** 8/8 golden, 8/8 lint-clean, every arm.
  Opus 5 medium had nothing to fix on these two fixtures on 2026-09-12 and
  has nothing today; the three arms cannot be told apart on the axis the
  changes were made for. The hook's case remains the Sonnet 5 pair of
  2026-09-12.
- **The hook is not a cost here.** `hook-only` and `tree` cost less a session
  than `1.14.0` (0.94 and 0.86 against 1.03), with the same edit counts. The
  2026-09-12 Opus 5 run read +8%; with two sessions an arm the sign of the
  difference is noise, and so was the +8%.
- **The text does what it says, where it can be seen.** The empty-literal
  case is gone from both `tree` `feed` tests and present in both old-text
  ones. The discard reason is on every write in every arm — the `go-http`
  bullet did not change that on this model, where the 1.14.0 Delete Pass
  already carried it. Body comments in `tree` are 2.75 a session against 3.5
  and 4.0: the 2026-09-12 reading of +1.8 comments does not repeat.
- **The step-6 sentence is followed in form and once not in substance.** 3/4
  `tree` sessions wrote `(hook)` per check; 1/4 wrote a pass and cited the
  hook's silence. The `hook-only` session that refused to do the same did so
  with the old text. One repetition each, so a form the text produced and
  not a rate.

## What this changes

- Nothing on the plugin: the arms are tied on correctness, lint, size and
  cost at n=1 on Opus 5 medium, as the 2026-09-12 Opus 5 run already was at
  n=5 on `gateway`.
- The three skill sentences have their first per-arm reading and it is
  neutral on this model. Whether they carry the Sonnet 5 golden 7/10 to 9/10
  reading, or the hook alone does, needs the same three arms on Sonnet 5 at
  n=5 or more (`feed` and `gateway`, about $9).
- The step-6 wording has a first observation: the `(hook)` form appears, the
  silence rule was broken once in four. If the rule is meant to hold, the
  sentence needs the case spelled out — a clean run after the last edit is a
  pass the hook reported by printing nothing, or it is not — because the
  model can read "silence is not a result" either way.
