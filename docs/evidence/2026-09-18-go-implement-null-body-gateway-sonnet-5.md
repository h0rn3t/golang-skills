# The empty-list case compares body text, not a decoded value — Sonnet 5 on `gateway`, `medium` at n=5 and `low` at n=3

After the [`HEAD` edit](2026-09-18-go-implement-head-case-gateway-sonnet-5.md)
every remaining Sonnet 5 golden miss on `gateway` was one clause: a nil
account list encoded as `null`. The four sessions that missed it today all
had a contract test, and all decoded the response body into a slice with
`json.Unmarshal` or `Decode` — which reads `null` as an empty slice, so the
model's own case passed on the very body the contract forbids and the edit
hook had nothing to print. The `go-code` Contract Table paragraph already
says the empty case is built from nil; it did not say how the case reads the
answer. One sentence added after that rule:

> A case for a list that may come back empty compares the body text with
> `[]` (`strings.TrimSpace(rec.Body.String())`), never a decoded value:
> `json.Unmarshal` reads `null` into an empty slice, and the case passes on
> the one body the contract forbids.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `66f06cb` with the card-route hooks and the `HEAD` edit copied in — the state the `HEAD` runs measured as `baseline`; plugin SHA-256 `b3d1240d7660af4731d2e8069840baee39278c3a74235f057708e66af6d3ad00` |
| `baseline` | the working tree: the same plus the sentence above; plugin SHA-256 `f0669b9c197f7461c78d1f7b29a0dae8fd3b52cc5003bd94d8869132bbca6188` |

## Runs

- Runner: `claude` 2.1.267; model `claude-sonnet-5`; corpus `implement`, fixture `gateway` alone; seed `7`; `-keep -verbose`; 0 CLI errors in both
- `medium`: 5 repetitions per arm, 10 sessions, `-j 4`, finished 2026-09-18 20:51 UTC, $5.40. Report: [`2026-09-18-go-implement-null-body-gateway-n5-sonnet-5-medium.json`](2026-09-18-go-implement-null-body-gateway-n5-sonnet-5-medium.json) (SHA-256 `4a581078241b451237d8474ada8a73ada33325816be9d1f5f5f3e5ecc0196ad8`); traces [`….traces.tar.gz`](2026-09-18-go-implement-null-body-gateway-n5-sonnet-5-medium.traces.tar.gz) (SHA-256 `fc42efaac1f86cbe00ad3d1478f3a8552259057ee4f859b60220ce1065863c11`)
- `low`: 3 repetitions per arm, 6 sessions, `-j 3`, finished 20:48 UTC, $2.25. Report: [`2026-09-18-go-implement-null-body-gateway-n3-sonnet-5-low.json`](2026-09-18-go-implement-null-body-gateway-n3-sonnet-5-low.json) (SHA-256 `a60137ea5636317b37d42201602a3ee8c38f330d20d68459467eddd7cd626052`); traces [`….traces.tar.gz`](2026-09-18-go-implement-null-body-gateway-n3-sonnet-5-low.traces.tar.gz) (SHA-256 `a4a06a468aea6a360fae63af16cfff125c25d5db015ec30ad87c88d7da5d4cee`)

```bash
go run ./cmd/abrun -corpus implement -tasks gateway -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 66f06cb + hooks + HEAD edit> -arms reference,baseline -n 5 -j 4 -seed 7 -timeout 10m -keep -verbose -out <json>
# -effort low -n 3 -j 3 for the second run
```

Columns as in the `HEAD` report, read from each session's own test file and
`gateway.go`.

## The reading

`medium`, n=5:

| Arm | Golden | Own test | Body text `[]` | nil case | `HEAD` case | `HEAD` patterns / `GET` | Hook findings | Turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|
| `reference` | 5/5 | 5/5 | 4/5 | 5/5 | 5/5 | 3/3 in 5/5 | 45 | — | 0.612 |
| `baseline` | 4/5 | 5/5 | 5/5 | 5/5 | 5/5 | 3/3 in 4/5 | 23 | — | 0.469 |

`low`, n=3:

| Arm | Golden | Own test | Body text `[]` | nil case | `HEAD` case | `HEAD` patterns / `GET` | Hook findings | $/session |
|---|---|---|---|---|---|---|---|---|
| `reference` | 2/3 | 1/3 | 1/3 | 1/3 | 1/3 | 3/3 in 3/3 | 10 | 0.360 |
| `baseline` | 2/3 | 3/3 | 2/3 | 2/3 | 3/3 | 3/3 in 3/3 | 16 | 0.390 |

Per session:

| Effort | Arm | Rep | Golden | Failing clause | Own test | Body `[]` | nil case | `HEAD` patterns | Hook findings | $ |
|---|---|---|---|---|---|---|---|---|---|---|
| `medium` | `reference` | 0 | pass | | yes | yes | yes | 3 | 12 | 0.786 |
| `medium` | `reference` | 1 | pass | | yes | no | yes | 3 | 12 | 0.596 |
| `medium` | `reference` | 2 | pass | | yes | yes | yes | 3 | 10 | 0.646 |
| `medium` | `reference` | 3 | pass | | yes | yes | yes | 3 | 6 | 0.553 |
| `medium` | `reference` | 4 | pass | | yes | yes | yes | 3 | 5 | 0.477 |
| `medium` | `baseline` | 0 | pass | | yes | yes | yes | 3 | 3 | 0.416 |
| `medium` | `baseline` | 1 | pass | | yes | yes | yes | 3 | 2 | 0.394 |
| `medium` | `baseline` | 2 | pass | | yes | yes | yes | 3 | 6 | 0.508 |
| `medium` | `baseline` | 3 | pass | | yes | yes | yes | 3 | 8 | 0.518 |
| `medium` | `baseline` | 4 | **fail** | `HEAD` → 200 ×3 | yes | yes | yes | **0** | 4 | 0.509 |
| `low` | `reference` | 0 | pass | | no | no | no | 3 | 2 | 0.296 |
| `low` | `reference` | 1 | **fail** | nil list → `null` | no | no | no | 3 | 4 | 0.339 |
| `low` | `reference` | 2 | pass | | yes | yes | yes | 3 | 4 | 0.446 |
| `low` | `baseline` | 0 | **fail** | nil list → `null` | yes | no | no | 3 | 7 | 0.438 |
| `low` | `baseline` | 1 | pass | | yes | yes | yes | 3 | 2 | 0.358 |
| `low` | `baseline` | 2 | pass | | yes | yes | yes | 3 | 7 | 0.373 |

**The test-shape column moved a little; the clause did not fail where the
column held.** Body-text comparison with `[]` in 7/8 tests against 5/8, and
no session whose test compared the body text shipped `null`. The two `null`
misses of this pair are the two sessions at `low` whose test had no
empty-list case of any shape: one wrote no test at all (`reference` r1), one
wrote a test with `HEAD` cases and nothing for the empty list (`baseline`
r0). The sentence can shape a case the model writes; it does not make the
model write the case, and at `low` the Contract Table step itself is
skipped or shortened in 3/6 sessions here and 3/4 in the earlier `low` runs.

**`HEAD` held at 15/16 across the pair, and the miss is a broken test.**
Both arms carry the `HEAD` edit; `HEAD` patterns beside every `GET` pattern
in 15/16 sessions, including the two `low` sessions that wrote no test. The
one miss (`medium` `baseline` r4) wrote its test after the code, built its
requests with a relative URL, and the hook printed every case failing with
`unsupported protocol scheme`; the session edited the test twice, read both
files, loaded `go-linting`, stopped, and reported `test pass (hook)` — a
check the hook printed as failing and no later edit cleared, which "The Edit
Hook Record" says is not a pass. The `HEAD` wording is identical in both
arms, so the miss is the corpus's coin; the false `pass (hook)` is a
finding about the report rule, recorded here for the next reading of that
rule.

**Cost is noise at this n.** $0.469 against $0.612 at `medium` — the
`baseline` sessions met the hook 23 times against 45 — and $0.390 against
$0.360 at `low`. Neither is a reading of one sentence.

## What this establishes

- With the `HEAD` edit in both arms, Sonnet 5 on `gateway` is 13/16 across
  this pair (5/5 and 4/5 at `medium`, 2/3 and 2/3 at `low`), against 4/8
  for the tree of the afternoon at the same two efforts (`HEAD` runs'
  `reference`: 4/5 and 0/3).
- The body-text sentence is followed where a case is written (7/8) and
  changes nothing where the case is not; the remaining `null` misses are
  sessions without an empty-list case, both at `low`.
- One `medium` session claimed `test pass (hook)` after the hook had
  printed a failing test and nothing cleared it. The record rule's
  "cleared by a later edit" condition needs its own count on the day's
  traces before it is reworded.

## Next

- Count `pass (hook)` claims against the last hook print on every
  shell-less session of the day; if the false claim repeats, the rule gets
  the form "the last hook print in the session decides".
- `gateway` at `medium` n≥10 for a golden claim on the two edits together;
  at `low`, the Contract Table step is the thing to measure, since the
  sentence has nothing to shape when the case is not written.
