# Delete Pass — `feed` and `gateway`, reference against baseline, Sonnet 5 medium (n=5)

The Sonnet 5 half of the measurement behind the Delete Pass edits. The
`reference` arm is the committed tree at `b25ae01` (release 1.13.0);
`baseline` is the working tree with the edits the
[Opus 5 `gateway` pair](2026-09-12-go-implement-delete-pass-gateway-n5-opus-5-medium.md)
describes: the one-comment-per-overridden-default rule, the
standard-library bullet with the `slices.Sorted(maps.Keys(m))` nil trap named,
the Delete Pass after the Contract Table, and the one-line-comment rule for
unexported helpers. `feed` is here because its 35-to-52-line spread on
2026-09-12 was the case for the pass; `gateway` because it is where Sonnet's
correctness has moved between runs.

This run measured the tree *before* the two lines added the same hour — the
explicit-discard sentence in the Delete Pass and the handler-registered-once
sentence in the Declaration Budget; its `baseline` plugin digest is
`782bf572c34526c87d27bab2e982785a0e55d3f4326f02c3497f9394c42e7d72`. Neither
line addresses anything below that changed between the arms.

## Run

- Finished: 2026-09-12 10:43 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`; fixtures: `feed`, `gateway`
- Arms: `reference`, `baseline`; 5 repetitions per fixture and arm, 20 sessions
- `reference`: a `git worktree` of `b25ae01`, plugin SHA-256
  `c5b603e94104ae6f7badca3c918dda9e258d6a163cb2eca5f80379d494065865`
- `baseline`: plugin SHA-256 `782bf572c34526c87d27bab2e982785a0e55d3f4326f02c3497f9394c42e7d72`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2 for the lint line
- Raw report: [`2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.json`](2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.json)
  (SHA-256 `88e0422ac9aaa7a2fcb7873b396227435734d0d018b13838135f264f2b6ac45b`)
- Session transcripts: [`2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.traces.tar.gz`](2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `ff2bc91e2e7bc64b703f74402f03d1eb4d3db039122f7a76681caf11309b6844`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $3.1288 reference, $3.4203 baseline — $0.313 against $0.342 a session

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of b25ae01> -arms reference,baseline \
  -tasks feed,gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-delete-pass-feed-gateway-n5-sonnet-5-medium.json
```

All 20 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell. `go-code` fired
first in every skilled session, 4.2 `Skill` messages a session in the
reference arm and 4.1 in baseline.

## Results

Lines are over valid runs; p is an exact permutation test on the mean, Fisher's
exact test for golden.

| Fixture | golden reference | golden baseline | Fisher p | Δlines reference | Δlines baseline | diff | perm p | Δfuncs | Δbcom | lint after, mean · clean | $ / run |
|---|---|---|---:|---|---|---:|---:|---|---|---|---|
| `feed` | 5/5 | **4/5** | 1.00 | 43, 38, 46, 43, 39 → 41.8 | 37, 41, 36, 39 → 38.2 | −3.5 | 0.135 | 0 / 0 | 0 / 0 | 0.40 · 3/5 / 0.25 · 3/4 | 0.273 / 0.292 |
| `gateway` | **1/5** | 4/5 | 0.21 | 74 (the one pass; 69, 71, 63, 69 failed) | 87, 69, 85, 78 → 79.8 | — | — | 1.0 / 3.8 | 1.4 / 0.8 | 2.00 · 0/1 / 2.00 · 0/4 | 0.353 / 0.392 |

### `feed`, how `kinds` was built

| Arm | `Kinds: []string{}` then append | hand-rolled key loop | `slices.AppendSeq(make(...), maps.Keys(m))` | `slices.Sorted(maps.Keys(m))` | `fmt.Errorf` around `json.Marshal` |
|---|---:|---:|---:|---:|---:|
| `reference` | 2 | 3 | 0 | 0 | 1 |
| `baseline` | 1 | 0 | 3 | **1** | 0 |

### `gateway`, per session

| Arm | rep | Golden | failure | Δlines | helpers | budget line names call sites | lint |
|---|---|---|---|---:|---|---|---:|
| `reference` | 0 | 0/1 | `HEAD` 200 | 69 | `writeJSON` | yes | 0 |
| `reference` | 1 | 0/1 | `GET /accounts` on nil input is `null` | 71 | `methodNotAllowed` | yes | 3 |
| `reference` | 2 | 0/1 | `HEAD` 200 | 63 | — | — | 3 |
| `reference` | 3 | 0/1 | `HEAD` 200 | 69 | `writeJSON` | yes | 2 |
| `reference` | 4 | 1/1 | | 74 | `filterActive`, `writeJSON`, one closure | yes | 2 |
| `baseline` | 0 | 1/1 | | 87 | `server` type with three handler methods, `methodNotAllowed`, `filterActive`, `writeJSON` | partly: "holds route state, used by 3 handlers" | 2 |
| `baseline` | 1 | 1/1 | | 85 | `handleMethodNotAllowed`, `handleHealthz`, `handleAccounts`, `handleAccount`, `writeJSON` | **no**: "each is the handler body for one of the four registered routes" | 2 |
| `baseline` | 2 | 1/1 | | 78 | `methodNotAllowed`, `filterActive`, `writeJSON` | yes | 2 |
| `baseline` | 3 | 1/1 | | 69 | `methodNotAllowed`, `writeJSON` | yes | 2 |
| `baseline` | 4 | 0/1 | `GET /accounts` on nil input is `null` | 74 | `methodNotAllowed`, `filterActive`, `writeJSON` | yes | 2 |

Every baseline `gateway` session registered the three `HEAD` patterns; two
reference sessions did. The two lint findings in every baseline session are
`errcheck` on the bare `w.Write([]byte("ok"))` and on the bare
`json.NewEncoder(w).Encode(v)`; the reference sessions carry the same two
plus, twice, a third.

## Reading

- **`feed`: the standard-library bullet was read, and the trap it names was
  written once anyway.** Three baseline sessions took the `AppendSeq` form
  the bullet gives, none hand-rolled the key loop that three reference
  sessions wrote, and none wrapped `json.Marshal`; the valid sessions are
  3.5 lines shorter (p = 0.135). The fifth wrote
  `kinds := slices.Sorted(maps.Keys(counts))` — the exact form the same
  bullet calls out as nil for an empty map — rendered `kinds` as `null`, and
  failed its own contract test, which carried the case, with no shell to run
  it. That is the recorded Sonnet 5 miss on this fixture, about one session
  in five since 2026-09-10, now made through the form the skill names. One
  session cannot say whether naming the trap primed it; a Sonnet `feed` n=10
  on this tree against the same tree with the bullet reworded to give only
  the safe form would.
- **`gateway`: correctness moved the way it has moved before, and helpers
  moved out.** The 1.13.0 tree answered `HEAD` with 200 in 3/5 sessions and
  passed 1/5, the working tree registered `HEAD` in 5/5 and passed 4/5; the
  remaining miss in each arm is the other clause, `slices.Clone` of a nil
  `accounts` encoding as `null`. Sonnet's `HEAD` rate on skilled trees has
  been 5/5, 4/10 and 4/5 on consecutive days; nothing in these edits speaks
  to `HEAD`, so the 1/5 to 4/5 is read as the fixture's variance, not the
  edit's effect. What the edit did change is where the helpers went: two
  baseline sessions turned every route handler into a package-level function
  or method, one of them behind a `server` struct, and wrote a budget line
  that names no second call site — the form the closure rule forbids and the
  budget's first rule rejects. That is what the handler-registered-once
  sentence added after this run is for; it is unmeasured on Sonnet.
- **Comments and lines.** Sonnet narrates little on either tree: `Δbcom` 0
  on `feed` in both arms, 1.4 against 0.8 on `gateway`. With one valid
  reference session, `gateway` lines do not compare.
- **Lint is level and dirty in both arms:** the bare `w.Write` and `Encode`
  in every baseline session and in four of five reference sessions. The
  explicit-discard sentence added after this run is aimed at the Opus 5
  regression on the same line; on Sonnet 5 there was nothing to regress.
- **Cost** $0.31 against $0.34 a session; `Skill` messages level.

## What this changes

- The `feed` half of the edit ships as measured: the safe form displaces
  the hand-rolled loop and the needless wrap, and the one `null` is the
  fixture's base rate through a named form; a rewording that gives only the
  safe form is the open question, not a regression.
- The two sentences added after this run go to the Opus 5 `gateway` rerun;
  their Sonnet reading is a later pair.
