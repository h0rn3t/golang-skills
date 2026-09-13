# Edit hook with tests and lint — `gateway`, reference against baseline, Opus 5 medium (n=5)

The Opus 5 half of the measurement behind the 2026-09-12 afternoon edits: the
edit hook that runs the package's tests and golangci-lint after every `.go`
edit, the `go-http` open-discard bullet, the nil-built empty case in the
Contract Table, and the safe-form-only standard-library bullet. The
[Sonnet 5 pair](2026-09-12-go-implement-edit-hook-feed-gateway-n5-sonnet-5-medium.md)
lists the four changes and measured them on `feed` and `gateway`. `gateway`
is here because it is the fixture where the Opus 5 discard regressed and
recovered on 2026-09-12 morning, and the one with the most writes for the
hook to find.

## Run

- Finished: 2026-09-12 13:35 UTC
- Runner: `claude` 2.1.267
- Model: `claude-opus-5`, reasoning effort `medium`
- Seed: `1`; `-j 4`
- Corpus: `implement`; fixture: `gateway`
- Arms: `reference`, `baseline`; 5 repetitions each, 10 sessions
- `reference`: a `git worktree` of `d1b7124` (release 1.14.0), plugin SHA-256
  `0d360ceb91aaddfc327a72e0c79bf2b911ba5b0e1fd759ceadf8c227a84e52fe`
- `baseline`: plugin SHA-256 `c59720d3ad45f9b4603cbece408d8ec9d19015e0d245f8053f026ade960e174d`
- Toolchain: Go 1.27.1 linux/amd64; golangci-lint 2.13.2
- Raw report: [`2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.json`](2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.json)
  (SHA-256 `0e1e7856d367c2e0865a94c9453d85afb37ba7098c143189033450a570cf120f`)
- Session transcripts: [`2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.traces.tar.gz`](2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.traces.tar.gz)
  (SHA-256 `9de99e8fa3529782fcf556ccac5edc95c2fb0f2c6e5eb1e941ec30fa344dde4a`)
- Cost: $4.9063 reference, $5.3123 baseline — $0.98 against $1.06 a session

```bash
go run ./cmd/abrun -corpus implement -runner claude -model claude-opus-5 -effort medium \
  -reference-root <worktree of d1b7124> -arms reference,baseline \
  -tasks gateway -n 5 -j 4 -seed 1 -keep -verbose \
  -out ../docs/evidence/2026-09-12-go-implement-edit-hook-gateway-n5-opus-5-medium.json
```

Every session completed without a CLI error, changed the fixture, stayed out
of the repository checkout, and passed the hidden golden test. The tool set
had no shell. `go-code` fired first in every session; 7.0 `Skill` messages a
session in reference, 6.8 in baseline.

## Results

| Arm | rep | Golden | Δlines | Δfuncs | Δbcom | discard reasons | edits | `go test` sections | `golangci-lint` sections | lint |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|
| `reference` | 0 | 1/1 | 72 | 2 | 4 | 0 | 3 | — | — | 0 |
| `reference` | 1 | 1/1 | 75 | 2 | 4 | 0 | 7 | — | — | 0 |
| `reference` | 2 | 1/1 | 76 | 2 | 5 | 2 | 4 | — | — | 0 |
| `reference` | 3 | 1/1 | 72 | 2 | 4 | 2 | 6 | — | — | 0 |
| `reference` | 4 | 1/1 | 77 | 2 | 5 | 0 | 5 | — | — | 0 |
| `baseline` | 0 | 1/1 | 68 | 2 | 5 | 2 | 7 | 1 | 4 | 0 |
| `baseline` | 1 | 1/1 | 79 | 2 | 7 | 2 | 6 | 2 | 2 | 0 |
| `baseline` | 2 | 1/1 | 78 | 2 | 5 | 0 | 7 | 4 | 1 | 0 |
| `baseline` | 3 | 1/1 | 80 | 2 | 6 | 2 | 7 | 4 | 4 | 0 |
| `baseline` | 4 | 1/1 | 73 | 2 | 8 | 2 | 7 | 3 | 1 | 0 |

| Arm | golden | valid | Δlines | Δfuncs | Δclos | Δbcom | lint after | lint clean | `Skill` msgs | edits | $ / run |
|---|---|---|---:|---:|---:|---:|---:|---:|---:|---:|---:|
| `reference` | 5/5 | 5 | 75, 77, 72, 76, 72 → 74.4 | 2.00 | 0.00 | 4.4 | 0.00 | 5/5 | 7.0 | 5.0 | 0.981 |
| `baseline` | 5/5 | 5 | 78, 73, 68, 80, 79 → 75.6 | 2.00 | 0.00 | 6.2 | 0.00 | 5/5 | 6.8 | 6.8 | 1.062 |

Lines +1.2 (exact permutation p = 0.70); body comments +1.8 (p = 0.048).
Every session in both arms wrote both writes as `_, _ = w.Write(body)` or
`io.WriteString`; no session wrote a bare one.

## Reading

- **Nothing to fix, so the hook found nothing to fix.** Reference was already
  golden 5/5 and lint-clean 5/5 after the morning's Tree B; baseline is the
  same. The hook reported a failing `go test` in every baseline session (1–4
  times) and a lint finding in every one (1–4 times), and each was gone by
  the last edit: 6.8 edits a session against 5.0.
- **What the hook costs here: +8%,** $0.98 to $1.06 a session, all of it the
  extra turns the hook messages bought. On a fixture the model already gets
  right, that is the price of the evidence and not of a fix.
- **Body comments +1.8, of which +1.2 is the discard reason.** Baseline wrote
  the reason on both writes in 5/5 sessions, reference in 2/5 (0.8 against
  2.0 a session). That is the `go-http` bullet's example being followed to
  the letter, one comment per write, which the Delete Pass allows. The other
  +0.6 is within the day's range.
- **Size is level:** 74.4 to 75.6, arms overlapping (68–80 against 72–77).
  Helpers stayed at two package-level functions a session, `Δclos` 0.

## What this changes

- On Opus 5 medium `gateway` the hook is a cost, not a correction: the
  morning's tree left it nothing to catch. Its case is the Sonnet 5 pair,
  where lint went from three findings in five sessions to none and the
  fixture's two one-in-five misses did not appear.
- The `go-http` bullet's example comment is copied per write. If a later
  Delete Pass reading counts it as prose, the bullet's example is where to
  shorten it, not the rule.
