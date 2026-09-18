# The idiom card through the host route — prompt note and edit gate, Sonnet 5 medium, two runs (n=3, n=2)

The [card-step runs of the afternoon](2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md)
closed the text route: Sonnet 5 medium read
`go-style-core/references/CURRENT-GO.md` in 0/24 sessions under three
wordings of `go-code` and `go-style-core`, and their "Next" named the host
routes to try. These two runs measure both at once, in one arm, against the
1.21.0 branch head:

- `go-prompt-routing.sh` names the card by its installed path in the note it
  already prints for `go-code` and `go-code-refactor` prompts, for a whole
  `Read` in the same message as the loads;
- `go-code-routing.sh` records a `Read` of the card (whole: no `offset` past
  line 1, no `limit` shorter than the file) and, in a session that loaded a
  router, refuses the first `.go` edit until one has happened, naming the
  path — once per session, like every other gate item.

The two are separable in the traces by order: a card `Read` before any gate
block is the note's, one after a block naming the card is the gate's. Run 1
carried the card in a sentence of its own in the note ("In the same message
as the `go-style-core` load, Read the idiom card whole …"); run 2 folded it
into the sentence that lists the loads ("Load `go-style-core` with it, and
the owners …; `go-testing` if you write or edit a test; and Read the idiom
card whole …: <path>. All of them in one message, before the first edit.").

| Arm | Tree |
|---|---|
| `baseline` | `git worktree` of `66f06cb` (the 1.21.0 branch head), plugin SHA-256 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0` |
| `reference`, run 1 | the same tree with the two hooks changed, the card in a note sentence of its own; plugin SHA-256 `a5cdeb420950c754d5a0814c37dc4acd43fc3e8348c200ccd03963db48cdcc96` |
| `reference`, run 2 | the same, the card folded into the note's load sentence; plugin SHA-256 `f77b3f47783f032ee71e1f6d7a761039f29af3af628f48407f874d56d949e7c1` — the hooks now in the tree |

## Runs

Run 1 (card sentence of its own):

- Finished: 2026-09-18 19:15 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `medium`; seed `3`
- Corpus `implement`, fixtures `roster`, `feed`, `gateway`, arms `reference`, `baseline`, 3 repetitions per fixture and arm, 18 sessions, `-j 4`, 0 CLI errors
- Report: [`2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n3-sonnet-5-medium.json`](2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n3-sonnet-5-medium.json) (SHA-256 `cd624ddb054b3d5998acb0d78fb3aa1553fa58f52da1510795b39cd142ea32b6`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n3-sonnet-5-medium.traces.tar.gz) (SHA-256 `39bde1db718138dfc6dcd0e0686308bd8162c118cfd72cca10e758d3c27e0f18`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $7.13

Run 2 (card in the load sentence):

- Finished: 2026-09-18 19:25 UTC; seed `4`; 2 repetitions per fixture and arm, 12 sessions; otherwise as run 1
- Report: [`2026-09-18-go-implement-card-hook-one-sentence-roster-feed-gateway-n2-sonnet-5-medium.json`](2026-09-18-go-implement-card-hook-one-sentence-roster-feed-gateway-n2-sonnet-5-medium.json) (SHA-256 `35c951f040a549f803bd0f83620db6dc8164d74efa29f4105bfafde0fe2cd5ab`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-hook-one-sentence-roster-feed-gateway-n2-sonnet-5-medium.traces.tar.gz) (SHA-256 `cc9b04089862d1775966900f7329ca471f9acf1b06dfd789e8630fe434ee84bf`)
- Cost: $4.73

```bash
go run ./cmd/abrun -corpus implement -tasks roster,feed,gateway -runner claude -model claude-sonnet-5 -effort medium \
  -reference-root <worktree of 66f06cb with the two hooks changed> -arms reference,baseline \
  -n 3 -j 4 -seed 3 -timeout 10m -keep -verbose -out <run 1 json>     # -n 2 -seed 4 for run 2
```

Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook
ran after every `.go` edit; no session had a shell tool. The event columns
are counted from the traces as in the afternoon's report; "note named card"
is the prompt hook's `UserPromptSubmit` output carrying the card's path.

## The reading

Run 1:

| Arm | Valid | Golden | Lint clean | Δlines | Δfuncs | Skill calls/s | Skill turns/s | Card `Read` | …from the note | …from the gate | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` (route) | 9/9 | 8/9 | 6/9 | +42.8 | +1.00 | 4.33 | 3.22 | **9/9** | 9 | 0 | 4 | 42 | 15.3 | 0.424 |
| `baseline` | 9/9 | 9/9 | 6/9 | +46.7 | +0.67 | 4.44 | 2.44 | 1/9 | — | — | 4 | 40 | 15.6 | 0.368 |

Run 2:

| Arm | Valid | Golden | Lint clean | Δlines | Δfuncs | Skill calls/s | Skill turns/s | Card `Read` | …from the note | …from the gate | Gate blocks | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` (route) | 6/6 | 6/6 | 4/6 | +42.3 | +0.33 | 4.83 | 2.67 | **6/6** | 5 | 1 | 3 | 24 | 14.0 | 0.423 |
| `baseline` | 6/6 | 6/6 | 4/6 | +49.0 | +1.67 | 4.50 | 2.33 | 0/6 | — | — | 3 | 26 | 13.7 | 0.365 |

Per session, run 1:

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +34 | go-code, go-style-core, go-testing, go-error-handling | 4 | whole, before the first edit | 1 | 3 | 18 | 0.402 |
| `reference` | `feed` | 1 | pass | 0 | +36 | go-code, go-style-core, go-testing, go-documentation, go-error-handling | 4 | whole, before | 1 | 3 | 19 | 0.479 |
| `reference` | `feed` | 2 | pass | 0 | +41 | go-code, go-style-core, go-error-handling, go-testing | 3 | whole, before | 1 | 1 | 11 | 0.281 |
| `reference` | `gateway` | 0 | **fail** | 0 | +63 | go-code, go-style-core, go-http, go-error-handling, go-testing | 3 | whole, before | 0 | 19 | 25 | 0.754 |
| `reference` | `gateway` | 1 | pass | 0 | +80 | go-code, go-style-core, go-http, go-error-handling, go-testing | 4 | whole, before | 0 | 4 | 14 | 0.499 |
| `reference` | `gateway` | 2 | pass | 0 | +81 | go-code, go-style-core, go-http, go-error-handling, go-testing | 3 | whole, before | 0 | 6 | 19 | 0.590 |
| `reference` | `roster` | 0 | pass | 5 | +17 | go-code, go-style-core, go-interfaces, go-testing | 3 | whole, before | 0 | 2 | 12 | 0.276 |
| `reference` | `roster` | 1 | pass | 5 | +18 | go-code, go-style-core, go-testing | 3 | whole, before | 0 | 2 | 9 | 0.240 |
| `reference` | `roster` | 2 | pass | 5 | +15 | go-code, go-style-core, go-testing, go-interfaces | 2 | whole, before | 1 | 2 | 11 | 0.295 |
| `baseline` | `feed` | 0 | pass | 0 | +38 | go-code, go-style-core, go-testing, go-error-handling | 3 | whole, before | 1 | 2 | 19 | 0.398 |
| `baseline` | `feed` | 1 | pass | 0 | +37 | go-code, go-style-core, go-data-structures, go-error-handling, go-testing | 2 | no | 0 | 6 | 14 | 0.318 |
| `baseline` | `feed` | 2 | pass | 0 | +43 | go-code, go-style-core, go-testing, go-error-handling | 3 | no | 1 | 3 | 15 | 0.318 |
| `baseline` | `gateway` | 0 | pass | 0 | +73 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 3 | 13 | 0.383 |
| `baseline` | `gateway` | 1 | pass | 0 | +83 | go-code, go-http, go-error-handling, go-testing, go-style-core | 2 | no | 0 | 6 | 17 | 0.418 |
| `baseline` | `gateway` | 2 | pass | 0 | +94 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 9 | 25 | 0.688 |
| `baseline` | `roster` | 0 | pass | 5 | +17 | go-code, go-style-core, go-testing, go-interfaces | 3 | no | 0 | 4 | 9 | 0.215 |
| `baseline` | `roster` | 1 | pass | 5 | +15 | go-code, go-style-core, go-interfaces, go-testing | 2 | no | 1 | 5 | 15 | 0.284 |
| `baseline` | `roster` | 2 | pass | 5 | +20 | go-code, go-style-core, go-testing, go-interfaces | 3 | no | 1 | 2 | 13 | 0.290 |

Per session, run 2:

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +38 | go-code, go-style-core, go-error-handling, go-data-structures, go-testing | 4 | whole, before | 0 | 3 | 13 | 0.404 |
| `reference` | `feed` | 1 | pass | 0 | +34 | go-code, go-style-core, go-error-handling, go-testing, go-linting | 4 | whole, before | 1 | 1 | 12 | 0.358 |
| `reference` | `gateway` | 0 | pass | 0 | +77 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | whole, before | 0 | 3 | 14 | 0.468 |
| `reference` | `gateway` | 1 | pass | 0 | +73 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | whole, before | 0 | 10 | 23 | 0.742 |
| `reference` | `roster` | 0 | pass | 5 | +17 | go-code, go-style-core, go-interfaces, go-testing | 2 | whole, before | 0 | 2 | 9 | 0.258 |
| `reference` | `roster` | 1 | pass | 5 | +15 | go-code, go-style-core, go-interfaces, go-testing | 2 | whole, **after the gate named it** | 2 | 5 | 13 | 0.310 |
| `baseline` | `feed` | 0 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-data-structures, go-error-handling | 2 | no | 0 | 2 | 12 | 0.302 |
| `baseline` | `feed` | 1 | pass | 0 | +41 | go-code, go-style-core, go-testing, go-error-handling | 2 | no | 0 | 2 | 10 | 0.276 |
| `baseline` | `gateway` | 0 | pass | 0 | +81 | go-code, go-style-core, go-http, go-error-handling, go-testing | 2 | no | 0 | 3 | 12 | 0.411 |
| `baseline` | `gateway` | 1 | pass | 0 | +95 | go-code, go-http, go-error-handling, go-testing, go-security, go-style-core | 3 | no | 1 | 10 | 23 | 0.627 |
| `baseline` | `roster` | 0 | pass | 5 | +17 | go-code, go-style-core, go-testing | 3 | no | 1 | 5 | 12 | 0.277 |
| `baseline` | `roster` | 1 | pass | 5 | +20 | go-code, go-style-core, go-testing, go-interfaces | 2 | no | 1 | 4 | 13 | 0.293 |

**The route reads the card: 15/15 against 1/15.** The note alone did it in
14/15 sessions — a whole `Read`, before the first edit, with no gate block
naming the card; the gate caught the fifteenth (`roster` r1 of run 2: the
session loaded `go-code`, `go-style-core` and `go-interfaces` in one message,
read the code and edited; the gate refused the edit for the card, the model
read it and retried). Every `Read` was whole; none used `offset` or `limit`.
The `baseline` read it once in 15, on its own, in the `feed` r0 of run 1.

**Where the card is named decides how the loads bunch.** With the card in a
note sentence of its own (run 1), Sonnet 5 loaded `go-code`, then
`go-style-core` with the card on a turn of its own, then the owners later,
often `go-testing` alone: 3.22 Skill turns a session against 2.44. Folded into
the sentence that lists the loads (run 2), `gateway` and `roster` sessions
loaded `go-style-core`, the owners and the card in one message (2.00 Skill
turns in 4/6) and the two `feed` sessions still spread them: 2.67 against
2.33. Run 2's wording is the one in the tree.

**Cost: +15% a session, of which the card is about half.** Pooled over
both runs, $0.424 against $0.367 (+15.6%, n=15 a side, exact permutation
p ≈ 0.30). The token accounting says where it goes: per session +7,000
cache-creation tokens and +80,000 cache-read tokens (a larger context
re-read over the same fourteen to fifteen turns), +1,300 output tokens. The
card is 9.4K characters — about 3,600 tokens on the Opus 5 tokenizer and
about 4,700 on Sonnet 5's, which the Sonnet 5 prompting guide says produces
roughly 30% more tokens for the same text — so its creation and its re-read
over the remaining turns come to about $0.03 of the $0.057; the rest is the
turns the loads spread over. `roster` sessions cost the same in both arms ($0.276 against
$0.272); `feed` (+19%) and `gateway` (+21%) carry the difference.

**Correctness and size did not move in a direction this n can read.** Golden
14/15 against 15/15: the miss is `gateway` r0 of run 1 answering `HEAD` with
200, the clause whose rate the 2026-09-11 and 2026-09-12 runs recorded
flipping day to day, in a session with 19 hook findings on its test file.
Lint-clean 10/15 in both arms (the `roster` findings are the neighbor's
`legacy.go`). Lines +42.6 against +47.6 and new functions +0.73 against
+1.07 lean toward the route; at n=15 with `gateway` spanning +63 to +98 that
is a lean, not a reading. `roster` cannot show the card's purpose on this
model: both arms wrote `slices.Sort` and no older form in 10/10 sessions, as
in every Sonnet 5 medium run since 2026-09-13.

## What this establishes

- The read the "Write Current Go" rule requires now happens on Sonnet 5
  medium, from the host: 15/15 through the note (14) and the gate (1),
  against 0/24 from three wordings of the skill text and 1/15 here.
- The gate is the backstop, not the mechanism: it fired for the card once
  in fifteen sessions, and a note that names the card in the same sentence
  as the loads keeps the one-message shape in 4/6 sessions.
- The route costs about $0.06 a session on this model at this effort,
  about half of it the card's own tokens; whether the read changes the code
  Sonnet 5 writes is not something `roster` can show, since Sonnet 5 writes
  the current form there without it. Haiku 4.5 is the model on which the
  card's effect is on record (0/5 older forms with it against 5/5 without,
  2026-09-13), and it read the card in 4/5 from the text alone; the route
  is unmeasured there.
- A refactor prompt gets the same note and gate; the refactor corpus has
  not been run with them.

## Next

- Haiku 4.5 on `roster` at n=5 with the route: the read rate should reach
  5/5 and the older-form count stay 0/5; that run is what credits the route
  with a code effect rather than a read.
- The refactor corpus with the route against this baseline at n≥3.
- If the +15% is to come down, the place is the `feed`-shaped session that
  still loads `go-style-core` and the card on a turn before the owners; a
  wording that makes the model list every load first is the Fable 5.1 form
  the batching run compared on Haiku.
