# DeepSeek V4.1-Flash on OpenCode — all three corpora against no skills, n=1

A firing smoke on a model the pack had never been run against. The point is
not an effect size: at one repetition per fixture nothing here is a claim
about any wording. The point is whether this model is a usable carrier at
all — whether the routers load, whether the sessions complete, and whether
each corpus has room on it — before anything inside a skill is measured on
it (see [`docs/RULE_OWNERSHIP.md`](../RULE_OWNERSHIP.md) for what the routers
own, and the note below on the model this one is easily confused with).

| Arm | Tree |
|---|---|
| `no-skill` | control: an isolated `HOME` with no `skills/` directory at all |
| `baseline` | the working tree at the 1.21.0 branch head, plugin SHA-256 `f0669b9c…`, copied into the arm's `HOME` |

## Run

- Finished: 2026-09-19 00:07 (refactor), 00:11 (implement), 00:17 (review),
  00:19 (the one retry), UTC+03:00
- Runner: `opencode` 1.18.31; model `opencode-go/deepseek-v4.1-flash`; seed `1`;
  `-j 4`; no `-effort` (the opencode CLI cannot ask for one and `abrun`
  rejects the flag rather than recording a level it did not get)
- Corpora `refactor` (4 fixtures), `implement` (7), `review` (6), arms
  `no-skill`, `baseline`, 1 repetition per fixture and arm — 34 sessions
- `baseline` plugin SHA-256
  `f0669b9c197f7461c78d1f7b29a0dae8fd3b52cc5003bd94d8869132bbca6188`
- Toolchain Go 1.27.1 darwin/arm64, golangci-lint 2.13.2 on PATH
- 1 CLI error: `refactor`/`pricing`/`no-skill` died on `database is locked`.
  Four concurrent sessions share one arm `HOME`, and opencode serializes its
  own state through a SQLite file there, so `-j 4` races it. Re-run alone at
  `-j 1` and merged into the arm below; keep `-j` at 1 or 2 for this runner,
  or expect to repair cells.
- Cost: $0.34 over the 34 sessions
- Reports: [`…-refactor-control-n1-deepseek-v4.1-flash.json`](2026-09-19-go-refactor-control-n1-deepseek-v4.1-flash.json)
  (SHA-256 `edf5869e52f34037c24ce33f7e35669f20613a47eba1f4523436db5c550a38cc`),
  its [retry cell](2026-09-19-go-refactor-control-n1-deepseek-v4.1-flash-pricing-retry.json)
  (`67d4f59ea9130adf03dfed0e71c9431899128de152b0c4c2770612ae9ed39155`),
  [`…-implement-…json`](2026-09-19-go-implement-control-n1-deepseek-v4.1-flash.json)
  (`c2d214fd603e04b4a6a069bf8098d178a9d8ca3beea56ab7e94022cdafa0b55f`),
  [`…-review-…json`](2026-09-19-go-review-control-n1-deepseek-v4.1-flash.json)
  (`c3030c79549af30b5b2b37a5859f630abd26526f9fad0110303b2dd4d300f46b`)
- Traces, one `traces/<arm>-<fixture>-r0.jsonl` per session:
  [refactor](2026-09-19-go-refactor-control-n1-deepseek-v4.1-flash.traces.tar.gz)
  (`f55fd5690c0e8d5b91766a3f864c55b238182595eabbe31e0a532a240d261b71`, 9 files —
  the failed cell and its retry are both in there),
  [implement](2026-09-19-go-implement-control-n1-deepseek-v4.1-flash.traces.tar.gz)
  (`26cbe2fbef5224eda4c0bcc03442f4d301caca0c170f1639233c061c92c4b17b`),
  [review](2026-09-19-go-review-control-n1-deepseek-v4.1-flash.traces.tar.gz)
  (`4f39b875c4247eca952c0f7920e887d93f18fe94a99fffcddc98c1cb33727ff1`)

```bash
for c in refactor implement review; do
  go run ./cmd/abrun -runner opencode -model opencode-go/deepseek-v4.1-flash \
    -corpus "$c" -arms no-skill,baseline -n 1 -j 4 -seed 1 -keep -out "$c.json"
done

go run ./cmd/abrun -runner opencode -model opencode-go/deepseek-v4.1-flash \
  -corpus refactor -tasks pricing -arms no-skill -n 1 -j 1 -seed 1 -keep \
  -out refactor-pricing-noskill.json
```

## Routers fire on this model

This is the result the run was for. `deepseek-v4-flash` — a different model,
one letter apart in the listing — never touched `go-code` at all on
`implement`/`gateway`, which is why anything measured inside the router on it
measured nothing. `deepseek-v4.1-flash` loads the routers:

| Corpus | Router | Sessions it fired in | Silent on |
|---|---|---|---|
| refactor | `go-code-refactor` | 4/4 | — |
| implement | `go-code` | 5/7 | `catalog`, `ledger` |
| review | `go-code-review` | 5/6 | `partner` |

The `no-skill` arm loaded nothing in 22/22 sessions, so the control is a
control. Where a router fired it pulled owners with it — up to eight skills
in one session (`fetch`: `go-context`, `go-error-handling`, `go-http`,
`go-linting`, `go-resilience`, `go-style-core`, `go-testing`) — which is the
shape the routing model is supposed to produce.

## Refactor — 4 fixtures, 8 valid sessions

| Arm | Δlines | Δfuncs | Δbranches | golden | lint after | skill | $/session |
|---|---|---|---|---|---|---|---|
| `no-skill` | −4.8 | +1.25 | −6.25 | 4/4 | 0.50 | 0/4 | 0.0046 |
| `baseline` | **−18.2** | +0.25 | −7.00 | 4/4 | 0.50 | 4/4 | 0.0164 |

Lint findings before the session were 3.25 a run in both arms, so the
`0.50` after is the same clean-up on both sides.

| Fixture | `no-skill` Δlines | `baseline` Δlines |
|---|---|---|
| `dispatch` | −10 | −10 |
| `store` | −15 | −15 |
| `pricing` | −25 | **−47** |
| `report` | **+31** | −1 |

Two fixtures are identical in both arms; the corpus difference is `pricing`
and `report`. `report` is the fixture recorded as bimodal on Sonnet 5 — a
flat renderer or three to four write helpers — and here the unaided session
took the second road, adding five functions and 31 lines to a refactoring
task. One session a side is exactly one draw from that distribution, so the
−18.2 is a direction to re-measure, not a size.

## Implement — 7 fixtures, 14 sessions

| Arm | Δlines | Δfuncs | golden | ran its own tests | skill | $/session |
|---|---|---|---|---|---|---|
| `no-skill` | +74.0 | +1.86 | 7/7 | 0/7 | 0/7 | 0.0050 |
| `baseline` | +56.0 | +0.43 | 7/7 | 5/7 | 4/7 | 0.0151 |

**The corpus has no room on this model.** The hidden golden test passes in
every session of both arms, so there is nothing here for a wording to fix,
and any `implement` run on this model measures size and cost only.

| Fixture | `no-skill` Δlines | `baseline` Δlines |
|---|---|---|
| `catalog` | +28 | +28 |
| `roster` | +22 | +21 |
| `feed` | +44 | +41 |
| `ledger` | +45 | +41 |
| `pool` | +64 | +48 |
| `gateway` | +95 | +75 |
| `fetch` | +220 | **+138** |

The gap grows with the task's size and is almost all `fetch`, where the
unaided session wrote 10 functions and 40 branches for a two-function
package. The skilled sessions that ran `go test` on their own work (4/7,
all of them router sessions) are the ones the opencode tool set makes
possible; the claude arm has no shell.

## Review — 6 fixtures, 12 sessions

| Arm | recall | must | must filed as Must Fix | read-only | baits hit | citations | unkeyed | verified markers | $/session |
|---|---|---|---|---|---|---|---|---|---|
| `no-skill` | 65/76 (0.86) | 29/31 | 22 | 57/68 | 4/13 | 106 | 27 | 5 | 0.0078 |
| `baseline` | **69/76 (0.91)** | **31/31** | **27** | **61/68** | 6/13 | 99 | **13** | 47 (+18 plausible) | 0.0123 |

Every column moves the skill's way except baits: the skilled arm flagged six
correct lines as defects against four. Off-key citations halved (27 → 13) and
evidence markers appear where the unaided arm had almost none, which is the
`go-code-review` filing discipline showing up in the output.

Defects only one arm found:

| Found only by `baseline` | Found only by `no-skill` |
|---|---|
| `books/self-transfer-mints` | `partner/throttle-burns-attempt` |
| `orders/id-with-commit-error` | `vault/partial-file-visible` |
| `partner/test-any-error` | |
| `partner/ticker-panics-on-zero` | |
| `vault/uploader-content-type-served` | |
| `worker/test-background` | |

Two of the six are tests that pass for the wrong reason
(`partner/test-any-error`, `worker/test-background`) — the defect class Opus 5
misses 0/4 unaided — so this corpus has room on this model where `implement`
has none.

## What this does and does not support

- It supports: the routers load on `deepseek-v4.1-flash`, so this model can
  carry a measurement of text inside `go-code`, `go-code-refactor` or
  `go-code-review`. Its cheapness (a 34-session sweep of all three corpora
  for $0.34) makes an n=10 run on it affordable where Sonnet 5 is not.
- It supports: `review` and `refactor` have room on it; `implement` does not,
  at 7/7 golden in both arms.
- It does not support: any statement of the form "the skills make this model
  write N fewer lines". One repetition per fixture, no repetition-level
  variance, and `report` alone carries the refactor gap.
- It does not support: any before/after claim about the 1.21.0 edits. That
  needs a `reference` arm from a second checkout, which this run did not have.
