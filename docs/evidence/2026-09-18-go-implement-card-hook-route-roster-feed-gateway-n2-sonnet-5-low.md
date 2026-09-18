# The idiom card through the host route at `low` effort — Sonnet 5, `roster`, `feed`, `gateway` at n=2

The [medium-effort runs](2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n3-sonnet-5-medium.md)
put the card route on record: the prompt note and the edit gate get
`go-style-core/references/CURRENT-GO.md` read in 15/15 Sonnet 5 medium
sessions at +15% cost. The Sonnet 5 prompting guide says the model "respects
effort levels strictly, especially at the low end" and at `low` "scopes its
work to what was asked rather than going above and beyond". This run asks
what that does to the route: the 1.21.0 branch head without the two hook
changes against the tree with them, one run, `-effort low`.

| Arm | Tree |
|---|---|
| `reference` (no card route) | `git worktree` of `66f06cb` (the 1.21.0 branch head): the prompt note names `go-style-core` and the owners, "All of them before the first edit"; the gate requires `go-style-core` and the owners. Plugin SHA-256 `bfcfed52ecd8273c265ff05738e65a40716a3caa9b16a78dde590c732cd14ef0` |
| `baseline` (card route) | the working tree: the note names the card by installed path in the load sentence, "All of them in one message, before the first edit"; the gate also requires one whole `Read` of the card. Plugin SHA-256 `f77b3f47783f032ee71e1f6d7a761039f29af3af628f48407f874d56d949e7c1`, the same bytes as run 2 of the medium report |

## Run

- Finished: 2026-09-18 20:19 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort **`low`**; seed `5`
- Corpus `implement`, fixtures `roster`, `feed`, `gateway`, arms `reference`, `baseline`, 2 repetitions per fixture and arm, 12 sessions, `-j 4`, 0 CLI errors
- Report: [`2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n2-sonnet-5-low.json`](2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n2-sonnet-5-low.json) (SHA-256 `6fecf89de17edeb97a5bb672164dae114822dbbec04cc5a4c01d94a2d2215fc0`); traces [`….traces.tar.gz`](2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n2-sonnet-5-low.traces.tar.gz) (SHA-256 `e043109eb1be7fa40e07d752d9e91abe6519f24585a97e662765947597818873`), one `traces/<arm>-<fixture>-r<rep>.jsonl` per session
- Cost: $3.76

```bash
go run ./cmd/abrun -corpus implement -tasks roster,feed,gateway -runner claude -model claude-sonnet-5 -effort low \
  -reference-root <worktree of 66f06cb> -arms reference,baseline \
  -n 2 -j 4 -seed 5 -timeout 10m -keep -verbose -out <json>
```

Toolchain and event columns as in the medium report; no session had a shell
tool, the edit hook ran after every `.go` edit.

## The reading

| Arm | Valid | Golden | Lint clean | Δlines | Skill calls/s | Skill turns/s | Card `Read` | …from the note | …from the gate | Gate blocks | …for `go-style-core` | Hook findings | Assistant turns/s | $/session |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` (no route) | 6/6 | 4/6 | 4/6 | +41.2 | 3.67 | 2.50 | 0/6 | — | — | 7 | 6 | 26 | 13.8 | 0.299 |
| `baseline` (route) | 6/6 | 4/6 | 4/6 | +39.8 | 4.00 | 2.00 | **6/6** | 6 | 0 | 3 | 1 | 26 | 12.0 | 0.328 |

| Arm | Fixture | Rep | Golden | Lint | Δlines | Skills loaded (in load order) | Skill turns | Card `Read` | Gate blocks (named) | Hook findings | Turns | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | 0 | pass | 0 | +40 | go-code, go-style-core, go-testing, go-error-handling | 3 | no | 2 (go-style-core; go-error-handling) | 3 | 13 | 0.332 |
| `reference` | `feed` | 1 | pass | 0 | +39 | go-code, go-style-core, go-testing, go-error-handling | 2 | no | 1 (go-style-core) | 4 | 12 | 0.314 |
| `reference` | `gateway` | 0 | **fail** (nil list → `null`) | 0 | +80 | go-code, go-http, go-error-handling, go-style-core | 3 | no | 1 (go-style-core) | 7 | 18 | 0.349 |
| `reference` | `gateway` | 1 | **fail** (`HEAD` → 200, and `null`) | 0 | +55 | go-code, go-http, go-error-handling, go-style-core | 3 | no | 1 (go-style-core) | 5 | 14 | 0.294 |
| `reference` | `roster` | 0 | pass | 5 | +15 | go-code, go-style-core, go-interfaces, go-testing | 2 | no | 1 (go-style-core) | 6 | 15 | 0.280 |
| `reference` | `roster` | 1 | pass | 5 | +18 | go-code, go-style-core | 2 | no | 1 (go-style-core) | 1 | 11 | 0.223 |
| `baseline` | `feed` | 0 | pass | 0 | +33 | go-code, go-style-core, go-testing, go-error-handling | 4 | whole, before the first edit | 2 (go-testing; go-error-handling) | 3 | 15 | 0.340 |
| `baseline` | `feed` | 1 | pass | 0 | +39 | go-code, go-style-core, go-error-handling, go-testing, go-data-structures | 2 | whole, before | 1 (go-style-core) | 1 | 10 | 0.341 |
| `baseline` | `gateway` | 0 | **fail** (`HEAD /healthz` → 200) | 0 | +66 | go-code, go-style-core, go-http, go-error-handling | 1 | whole, before | 0 | 4 | 10 | 0.308 |
| `baseline` | `gateway` | 1 | **fail** (`HEAD` → 200) | 0 | +69 | go-code, go-style-core, go-http, go-error-handling, go-testing, go-security | 2 | whole, before | 0 | 8 | 21 | 0.559 |
| `baseline` | `roster` | 0 | pass | 5 | +17 | go-code, go-style-core | 2 | whole, before | 0 | 4 | 8 | 0.197 |
| `baseline` | `roster` | 1 | pass | 5 | +15 | go-code, go-style-core, go-interfaces | 1 | whole, before | 0 | 6 | 8 | 0.224 |

**The note is followed at `low` as at `medium`: card read 6/6, all before
the first edit, none needing the gate.** Without the route the card is read
0/6. Sonnet 5 at `low` reads what the turn tells it to and nothing beyond,
which is the guide's description of the level.

**The route arm met the gate less, not more.** Seven gate blocks against
three, and the difference is one skill: without the route, 6/6 sessions
tried their first `.go` edit before loading `go-style-core` and the gate
named it every time; with the route, 1/6. The two arms differ in two things
the note says — the card, and "All of them in one message" in place of "All
of them before the first edit" — and this run cannot say which of the two
got `go-style-core` loaded in time; at `medium` the same tree met the gate
for `go-style-core` in 3/15 sessions against 3/15, so the effect is a
`low`-effort one if it is real. Skill turns 2.00 against 2.50 and assistant
turns 12.0 against 13.8 follow from the same fact: a block is a wasted edit
and two extra turns.

**Cost +10% ($0.328 against $0.299), the card's own tokens and nothing
else visible.** Per session the route arm created 48K cache tokens against
45K (the card is about 4,700 tokens on Sonnet 5's tokenizer) and re-read
449K against 407K over fewer turns — the larger context per turn outweighs
the two turns saved. Output 4.5K against 3.8K tokens.

**`low` is where `gateway` fails, in both arms.** Golden 4/6 against 4/6,
the four misses all `gateway`: the nil-list `null` once, the `HEAD` → 200
clause three times, one session both. At `medium` the same fixture was
11/13 for this tree today and 7/7 for 1.20.1; at `low` it is 0/4. The
guide's "risk of under-thinking at `low`" has a number here. `roster` is
saturated at `low` as at `medium`: no older form in 4/4, `slices.Sort` and
`slices.ContainsFunc` in every session.

## What this establishes

- The host route works at `low`: 6/6 card reads from the note alone, the
  gate never fired for the card, +10% a session (n=2 a fixture, so the
  dollar figure is the card's tokens and noise).
- At `low` the tree without the route edited before loading `go-style-core`
  in every session and the route arm in one; whether the card line or the
  "in one message" clause did that needs the two separated, at n≥3.
- `gateway` correctness at `low` is 0/4 in both arms on the `HEAD` and
  nil-list clauses; nothing in either text touches that, and the fixture
  says what `low` costs on this model.

## Next

- Separate "in one message" from the card line in the note at `low`, n=3,
  if the gate-block difference is to be credited to either.
- The `high` run that the medium report asked for is still the one that
  tells whether default-effort users need the route at all.
