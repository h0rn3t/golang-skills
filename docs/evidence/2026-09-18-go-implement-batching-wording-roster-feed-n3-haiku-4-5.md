# Three wordings of the owner-load batching clause — Haiku 4.5, `roster` and `feed` at n=3

`go-code` step 3 asks for the owner skills to be loaded "with the `Skill`
calls for all of them in one message: an owner loaded on a turn of its own
re-reads the whole context" — a cost explanation, where the prompting guide's
`<use_parallel_tool_calls>` block and the Fable 5.1 page's one-line nudge
("First privately list what you need next; then request every item that
doesn't depend on another's result in this one response") are imperatives.
The 2026-09-13 routing runs found Sonnet 5 loading one skill per turn; the
prompt hook's owner list brought that back to one message. On 2026-09-18
Sonnet 5 medium loaded the owners in one message in 24/24 sessions of the
[card-step runs](2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md)
(2.00 Skill turns a session: `go-code`, then everything else), so the clause
has no headroom there. Haiku 4.5 still spreads the loads (2.4–2.6 Skill turns
on `roster` the same day), so the three wordings are compared on it, as an
event count.

| Arm | Wording of step 3's load clause |
|---|---|
| `baseline` (current) | "…with the `Skill` calls for all of them in one message: an owner loaded on a turn of its own re-reads the whole context." |
| `parallel` | "The `Skill` calls for them have no dependency on one another: make all of them in parallel, in this one message, and none on a turn of its own." — the `<use_parallel_tool_calls>` form |
| `fable` | "First privately list every owner this step needs; then request every one of them in this one response, since none depends on another's result." — the Fable 5.1 page's nudge |

Two runs, sharing the `baseline` arm of the first: `parallel` against
`baseline`, then `fable` alone. Every other byte of the three trees is the
working tree of the day (`bfcfed52…`).

## Runs

- Run 1 (`parallel`, `baseline`) finished 2026-09-18 17:01 UTC; run 2 (`fable`) finished 17:05 UTC; seed `1` both
- Runner: `claude` 2.1.267; model `claude-haiku-4-5-20251001`, no reasoning-effort flag
- Corpus `implement`, fixtures `roster`, `feed`, 3 repetitions per fixture and arm, 12 + 6 sessions, `-j 3`, 0 CLI errors
- Plugin SHA-256: `baseline` `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0`; `parallel` `1a1416226b739d9cfda22466c2cf7d00fbc0b288ff86888a7638b7f1807d813f`; `fable` `f9567c94e7f29e7d441f1cec83dff3338c20ae2470942bba7b0650a6a91e5e58`
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook ran in every session, no session had a shell tool
- Run 1 report: [`2026-09-18-go-implement-batching-parallel-vs-current-roster-feed-n3-haiku-4-5.json`](2026-09-18-go-implement-batching-parallel-vs-current-roster-feed-n3-haiku-4-5.json) (SHA-256 `8958bcf5000c87d027d6b53e9a5167212147b2b53f8da0b5790c98bd1109e737`); traces [`….traces.tar.gz`](2026-09-18-go-implement-batching-parallel-vs-current-roster-feed-n3-haiku-4-5.traces.tar.gz) (SHA-256 `9ec15ead0ab03085125fc23bd95de647d520d059eab7df1e52639238e039b624`); $2.55
- Run 2 report: [`2026-09-18-go-implement-batching-fable-roster-feed-n3-haiku-4-5.json`](2026-09-18-go-implement-batching-fable-roster-feed-n3-haiku-4-5.json) (SHA-256 `ab32021d5d28ec641fd7b5637069e21a2e1acde4427cda47027468f28dafb4a1`); traces [`….traces.tar.gz`](2026-09-18-go-implement-batching-fable-roster-feed-n3-haiku-4-5.traces.tar.gz) (SHA-256 `456b7852e8ea79b7edfe8424824de10eb72f44617ad34c108f81724384f952f4`); $1.06

```bash
go run ./cmd/abrun -corpus implement -tasks roster,feed -runner claude -model claude-haiku-4-5-20251001 \
  -reference-root <copy of the tree with the parallel wording> -arms reference,baseline \
  -n 3 -j 3 -seed 1 -timeout 10m -keep -verbose -out <run 1 json>
go run ./cmd/abrun -corpus implement -tasks roster,feed -runner claude -model claude-haiku-4-5-20251001 \
  -reference-root <copy of the tree with the fable wording> -arms reference \
  -n 3 -j 3 -seed 1 -timeout 10m -keep -verbose -out <run 2 json>
```

The `reference` arm of each JSON is the named variant. Event columns are
counted from the traces as in the card-step reports; a Skill turn is one
assistant message with one or more `Skill` calls, so the floor is 2
(`go-code`, then the owners) and every turn above it is an owner loaded
alone.

## The reading

| Wording | Sessions | Golden | Lint clean | Skill calls/s | Skill turns/s | Sessions at the 2-turn floor | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|
| current | 6 | 6/6 | 3/6 | 3.67 | 3.33 | 1/6 | 5 | 26 | 18.2 | 0.170 |
| `parallel` | 6 | 4/6 | 1/6 | 4.50 | 3.33 | 2/6 | 5 | 50 | 26.3 | 0.256 |
| `fable` | 6 | 6/6 | 3/6 | 3.67 | 3.00 | 2/6 | 3 | 29 | 17.7 | 0.176 |

Per session:

| Wording | Fixture | Rep | Golden | Skills loaded (in load order) | Skill turns | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|
| current | `feed` | 0 | pass | go-code, go-style-core, go-testing, go-error-handling | 4 | 1 | 5 | 22 | 0.210 |
| current | `feed` | 1 | pass | go-code, go-style-core, go-testing, go-error-handling | 4 | 2 | 7 | 23 | 0.201 |
| current | `feed` | 2 | pass | go-code, go-style-core, go-error-handling | 3 | 1 | 2 | 18 | 0.192 |
| current | `roster` | 0 | pass | go-code, go-style-core, go-testing, go-data-structures | 4 | 0 | 5 | 16 | 0.139 |
| current | `roster` | 1 | pass | go-code, go-style-core, go-testing, go-data-structures | 2 | 0 | 5 | 15 | 0.140 |
| current | `roster` | 2 | pass | go-code, go-style-core, go-testing | 3 | 1 | 2 | 15 | 0.137 |
| `parallel` | `feed` | 0 | pass | go-code, go-style-core, go-testing, go-error-handling | 3 | 1 | 8 | 27 | 0.280 |
| `parallel` | `feed` | 1 | pass | go-code, go-style-core, go-error-handling, go-interfaces, go-linting | 5 | 2 | 12 | 40 | 0.377 |
| `parallel` | `feed` | 2 | **fail** | go-code, go-style-core, go-testing, go-defensive, go-error-handling | 4 | 1 | 14 | 37 | 0.391 |
| `parallel` | `roster` | 0 | pass | go-code, go-testing, go-data-structures, go-style-core, go-linting | 4 | 1 | 6 | 20 | 0.204 |
| `parallel` | `roster` | 1 | pass | go-code, go-style-core, go-testing, go-data-structures | 2 | 0 | 5 | 16 | 0.135 |
| `parallel` | `roster` | 2 | **fail** | go-code, go-style-core, go-testing, go-data-structures | 2 | 0 | 5 | 18 | 0.146 |
| `fable` | `feed` | 0 | pass | go-code, go-style-core, go-error-handling, go-testing | 3 | 1 | 6 | 18 | 0.239 |
| `fable` | `feed` | 1 | pass | go-code, go-style-core, go-error-handling | 3 | 1 | 5 | 21 | 0.183 |
| `fable` | `feed` | 2 | pass | go-code, go-style-core, go-error-handling, go-testing | 4 | 1 | 7 | 23 | 0.223 |
| `fable` | `roster` | 0 | pass | go-code, go-style-core, go-testing, go-data-structures, go-linting | 4 | 0 | 5 | 18 | 0.161 |
| `fable` | `roster` | 1 | pass | go-code, go-style-core, go-testing, go-data-structures | 2 | 0 | 5 | 17 | 0.138 |
| `fable` | `roster` | 2 | pass | go-code, go-style-core | 2 | 0 | 1 | 9 | 0.113 |

**No wording brings Haiku to the floor.** One to two sessions in six load
everything after `go-code` in one message; the rest load an owner or two on
a turn of their own, and the gate names the missing one at the first edit in
about half the sessions. The `parallel` form is indistinguishable from the
current one on the batching columns (3.33 turns, five blocks each) and worse
on everything else in this sample — two golden failures, 50 hook findings
against 26, +51% a session — which at n=3 per fixture on a model with a
recorded `feed` miss rate is the fixture's variance, not the sentence's.

**The Fable 5.1 form leans the right way by one session's worth.** 3.00
Skill turns against 3.33, three gate blocks against five, golden 6/6, cost
level ($0.176 against $0.170). The Fable 5.1 page frames its nudge as a
per-turn system message, re-sent after every tool result; a `SKILL.md` line
is read once, and this is what once buys.

**The wording is not the lever.** The three arms differ by one sentence and
the batching columns move by at most one turn in three; the 2026-09-13
change that did move Sonnet 5 from one load per turn to one message was the
prompt hook's owner list, delivered by the host before the first edit.

## What this establishes

- On Sonnet 5 medium the clause is saturated (24/24 sessions at the 2-turn
  floor); measuring wordings there would measure noise.
- On Haiku 4.5 none of the three wordings reaches the floor; the Fable 5.1
  form is the only one that leans toward it, by one session in six, and the
  `<use_parallel_tool_calls>` form does not. The current sentence stays.
- The per-turn nudge the Fable 5.1 page describes is a hook's job
  (PostToolUse `additionalContext` after each `Skill` result), not a
  sentence's; that is the arm still to run.

## Next

- `fable` against current on Haiku at n≥5 per fixture if the one-turn lean
  is worth $3 to confirm; otherwise leave the sentence and test the hook
  route.
