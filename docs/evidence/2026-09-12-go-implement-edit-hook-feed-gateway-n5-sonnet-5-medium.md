# Edit hook with tests and lint — `feed` and `gateway`, reference against baseline, Sonnet 5 medium (n=5)

The Sonnet 5 half of the measurement behind the 2026-09-12 afternoon edits.
The `reference` arm is the committed tree at `d1b7124` (release 1.14.0);
`baseline` is the working tree with four changes measured together:

- `hooks/go-vet-on-edit.sh` runs the package's own tests (`go test -short
  -count=1`) and `golangci-lint` — the repository's configuration when it has
  one, the bundled `skills/go-linting/assets/golangci.yml` otherwise — after
  every edit of a `.go` file, once the package type-checks, and prints the
  failures back to the session. The hook is the host's process, so it runs in
  a session whose tool set has no shell, which every `claude` arm here is.
- `go-http` gains the open-discard bullet: `_, _ = w.Write(body)` with its
  reason, never bare, with `errcheck` and `gosec` G104 named. Until now the
  idiom lived only in `go-code`'s Delete Pass.
- `go-code` Contract Table: the empty case is built from nil, not from an
  empty literal, because `[]T{}` already has the shape the case is meant to
  prove and `slices.Clone` of nil is nil.
- `go-code` Plain Code: the standard-library bullet gives only the safe form
  for a map's keys (`AppendSeq` into `make`, then `Sort`) and points at
  `go-data-structures` for the collectors whose empty result is nil, instead
  of naming `slices.Sorted(maps.Keys(m))` as the trap. One Sonnet 5 `feed`
  session in five had written the named trap on the previous tree.

The [Opus 5 pair](2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.md)
measured the same tree on `gateway`. The four changes are one arm; nothing
below attributes a reading to one of them alone.

## Run

- Finished: 2026-09-12 13:34 UTC
- Runner: `claude` 2.1.267
- Model: `claude-sonnet-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`; fixtures: `feed`, `gateway`
- Arms: `reference`, `baseline`; 5 repetitions per fixture and arm, 20 sessions
- `reference`: a `git worktree` of `d1b7124`, plugin SHA-256
  `0d360ceb91aaddfc327a72e0c79bf2b911ba5b0e1fd759ceadf8c227a84e52fe`
- `baseline`: plugin SHA-256 `c59720d3ad45f9b4603cbece408d8ec9d19015e0d245f8053f026ade960e174d`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2, inside the hook and
  for the harness's lint line
- Raw report: [`2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.json`](2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.json)
  (SHA-256 `ca155138dd3c3e5c76624d22ecd5ec88e08b201da5f0784dfa031adca6717084`)
- Session transcripts: [`2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.traces.tar.gz`](2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.traces.tar.gz)
  (SHA-256 `51eeafd1e8f25088654dce89b05664aabd9216f83a83986d0aa869bfe3879fb2`),
  one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $3.6257 reference, $3.8228 baseline — $0.363 against $0.382 a session

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of d1b7124> -arms reference,baseline \
  -tasks feed,gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.json
```

All 20 sessions completed without a CLI error, changed the fixture, and stayed
out of the repository checkout. The tool set had no shell. `go-code` fired
first in every session, 4.6 `Skill` messages a session on `feed` and 6.4 on
`gateway` in the reference arm, 4.8 and 6.2 in baseline.

Two of the harness's own `lint before` readings (one per arm) came back
`n/a` with `parallel golangci-lint is running`: the hook's linter inside a
session and the harness's linter outside it share golangci-lint's global
lock. Both now pass `--allow-parallel-runners`; the `lint after` readings
were taken in all 20 sessions.

## Results

Lines are over valid runs; p is an exact permutation test on the mean.

| Fixture | golden reference | golden baseline | Δlines reference | Δlines baseline | perm p | lint after, reference | lint after, baseline | $ / run |
|---|---|---|---|---|---:|---|---|---|
| `feed` | 4/5 | **5/5** | 40, 37, 41, 41 → 39.8 | 41, 42, 43, 41, 33 → 40.0 | 0.94 | 0, 1, 1, 0, 0 · clean 3/5 | 0, 0, 0, 0, 0 · clean 5/5 | 0.253 / 0.248 |
| `gateway` | 3/5 | **4/5** | 86, 72, 80 → 79.3 | 76, 81, 88, 75 → 80.0 | 0.91 | 0 in every session | 0 in every session | 0.472 / 0.516 |

### What the hook said, per baseline session

| Fixture | rep | Golden | edits | `go test` sections | `golangci-lint` sections |
|---|---|---|---:|---:|---:|
| `feed` | 0 | 1/1 | 7 | 2 | 0 |
| `feed` | 1 | 1/1 | 6 | 1 | 1 |
| `feed` | 2 | 1/1 | 3 | 0 | 0 |
| `feed` | 3 | 1/1 | 5 | 1 | 0 |
| `feed` | 4 | 1/1 | 5 | 0 | 1 |
| `gateway` | 0 | 1/1 | 10 | 3 | 4 |
| `gateway` | 1 | **0/1**, `HEAD` 200 | 20 | 15 | 13 |
| `gateway` | 2 | 1/1 | 8 | 0 | 2 |
| `gateway` | 3 | 1/1 | 7 | 0 | 1 |
| `gateway` | 4 | 1/1 | 10 | 1 | 4 |

A section is one hook message carrying a failing `go test` or a lint finding
in the edited file; the reference arm's hook has neither and printed none.

### The failures

- `reference` `feed` #4: `TestRenderEmptyKeepsEveryMemberTyped/no_events` —
  a `null` where the contract says `[]`, the fixture's recorded miss of about
  one Sonnet 5 session in five. The session's own contract test failed on the
  same clause and, with no shell and the old hook, nothing ran it.
- `reference` `gateway` #0: `HEAD /healthz` answered 200.
- `reference` `gateway` #2: `TestEmptyListIsJSONArray/nil` — `GET /accounts`
  on nil input is `null`, the clause the new Contract Table sentence names.
- `baseline` `gateway` #1: `HEAD /healthz` answered 200, after 20 edits and 15
  hook test reports, at $0.66. The session's own contract test passed at the
  end: it carried no `HEAD` case, so the hook had nothing to run against the
  clause. The hook runs the tests the model wrote; it does not write them.

## Reading

- **Lint went to zero in every baseline session.** The two reference `feed`
  sessions with a finding after (an unchecked write, a `prealloc`) have no
  baseline counterpart; on `gateway` both arms were already clean at this
  effort. On the previous tree every Sonnet 5 `gateway` session carried two
  `errcheck` findings; between the `go-http` bullet and the hook, none does.
- **Golden 7/10 to 9/10.** The `feed` miss and the nil-`accounts` miss are
  the two clauses the skill text now names, and neither occurred in five
  baseline sessions of either fixture; at n=5 against base rates of one in
  five this is consistent with the edits and does not establish them. The one
  baseline failure is the `HEAD` clause, whose Sonnet rate has moved on every
  consecutive day and which the hook cannot reach when the model's own test
  has no case for it.
- **Size is level:** +0.2 lines on `feed`, +0.7 on `gateway`, p > 0.9 on both.
  Body comments 0.0 to 0.4 on `feed` and 2.0 to 2.8 on `gateway`; the extra
  ones are the discard reasons the `go-http` bullet asks for, one per write.
- **The hook is cheap on this model:** $0.363 to $0.382 a session (+5%),
  with `feed` slightly cheaper and `gateway` 9% dearer, where the sessions
  that received the most hook messages also made the most edits.
- **The sessions read the hook as evidence.** Several baseline reports wrote
  `lint pass … test pass (verified via editor hooks; no direct shell access)`
  on the checks line where the skill text said `unavailable (no shell)`. The
  step-6 wording was changed after this run to name hook output as the
  observed result for the checks it ran; that sentence is unmeasured.

## What this changes

- The hook ships as measured: every baseline session lint-clean, correctness
  held or better, +5% cost on Sonnet 5.
- The three skill sentences ship with it; their separate effect is not
  measured here and the `feed` and nil readings are one-in-five base rates
  observed at n=5.
- `abrun` and the hook pass `--allow-parallel-runners` to golangci-lint so a
  harness reading cannot fail on a session's lock.
