# `encoding/json/v2` as the default for new JSON code — Sonnet 5 `low` on `feed` and `gateway`, n=1

A smoke, not an effect: one session per arm and fixture. It reads whether the
2026-09-19 edit set changes the shapes it was written against, on the effort
level where Sonnet 5 scopes tightest to the ask.

The edit set, against release 1.21.1: `go-code` Plain Code says new JSON in a
package with no `encoding/json` import is `encoding/json/v2`, names the filter
loop as a call (`slices.DeleteFunc` on a clone), and gains a Delete Pass bullet
for a value set to the library's default (`MaxHeaderBytes: 1 << 20`,
`Content-Type: text/plain` before a text write); the Declaration Budget names
the `type server struct` plus method-per-route shape; the idiom card's v2 row
names `Marshal`, `MarshalWrite`, `UnmarshalRead` and scopes the `[]`-must-encode
forms to v1; `go-http`'s Handler Shape is the one-decode `UnmarshalRead` form
with `MarshalWrite` for the response, and its Server Construction row says the
header caps are set only to change the defaults; `go-code-refactor` ties an
added helper to the budget's three rules. The 17 `gateway` and `feed` sessions
of 2026-09-12..18 that motivated it had `encoding/json/v2` in 0/17,
`make([]T, 0, n)` or `slices.AppendSeq(make…)` for the wire in 14/17, and a
`writeJSON` helper in 11/14 `gateway` sessions.

| Arm | Tree |
|---|---|
| `reference` | `git worktree` of `904793d` (release 1.21.1); plugin SHA-256 `1747042b38be8026315c9495e2f2e3ee875eed2de929c0717b9de79942bbe26e` |
| `baseline` | the working tree with the edit set; plugin SHA-256 `3aac7f9eff0d22f7d433670a0f97bbcd7f8e10dd0049bbc87ffbf2cae35367db` |

## Run

- Finished: 2026-09-19 08:52 UTC
- Runner: `claude` 2.1.267; model `claude-sonnet-5`, effort `low`; seed `8`
- Corpus `implement`, fixtures `feed`, `gateway`, arms `reference`, `baseline`, one repetition per fixture and arm, 4 sessions, `-j 4`, 0 CLI errors
- Toolchain Go 1.27.1 linux/amd64, golangci-lint 2.13.2 on PATH; the edit hook ran in every session, no session had a shell tool
- Cost: $1.57
- Report: [`2026-09-19-go-implement-json-v2-feed-gateway-n1-sonnet-5-low.json`](2026-09-19-go-implement-json-v2-feed-gateway-n1-sonnet-5-low.json) (SHA-256 `ec709199729cdf8761967c9744c636fd8a04b700b5a54cfd05dd3098f968ceaf`); traces [`….traces.tar.gz`](2026-09-19-go-implement-json-v2-feed-gateway-n1-sonnet-5-low.traces.tar.gz) (SHA-256 `ad48f3baa686b21007514ae93b13037117fd1021249f60542218e8c5c12e3040`), one `traces/<arm>-<fixture>-r0.jsonl` per session

```bash
go run ./cmd/abrun -corpus implement -tasks feed,gateway -runner claude -model claude-sonnet-5 -effort low \
  -reference-root ../../golang-skills-1.21.1 -arms reference,baseline -n 1 -j 4 -seed 8 -timeout 10m -keep -verbose \
  -out ../docs/evidence/2026-09-19-go-implement-json-v2-feed-gateway-n1-sonnet-5-low.json
```

Besides the corpus columns, each session's production file is read for the
forms the edit names; the turn and load counts come from the traces.

## The reading

| Arm | Fixture | Golden | Own test | `json/v2` | wire `make`/`AppendSeq` | `writeJSON` | `MaxHeaderBytes` = default | `text/plain` on `ok` | Δlines | Δtypes | Δclos | Turns | Skill loads | $ |
|---|---|---|---|---|---|---|---|---|---|---|---|---|---|---|
| `reference` | `feed` | pass | none | no | `make([]eventDoc, 0, n)`, `AppendSeq(make…)` + `Sort` | — | — | — | +38 | 0 | 0 | 21 | 4 | 0.316 |
| `baseline` | `feed` | pass | none | yes | none | — | — | — | +43 | **2** | 0 | 23 | 3 | 0.270 |
| `reference` | `gateway` | pass | yes | no | `make` + `copy`, `make([]Account, 0, n)`, a filter loop | closure, `json.NewEncoder(w).Encode` | yes | yes | +79 | 0 | 2 | 35 | 5 | 0.534 |
| `baseline` | `gateway` | pass | yes | yes | none | none — `json.MarshalWrite(w, v)` at both sites | no | no | +64 | 0 | 1 | 26 | 5 | 0.446 |

**`gateway`: every named form is gone, −15 lines, golden held.** The baseline
session imported `encoding/json/v2`, wrote `_ = json.MarshalWrite(w, list)`
with its reason at the two response sites instead of a `writeJSON` closure,
filtered with `slices.DeleteFunc(slices.Clone(sorted), …)` — the form the new
Plain Code bullet shows — where the reference wrote `make([]Account, 0, n)`
and a loop, copied with `slices.Clone` where the reference wrote `make` plus
`copy`, and left `MaxHeaderBytes` and the `text/plain` header out. Three
`HEAD` patterns beside three `GET` patterns in both arms; both wrote a contract
test with `HEAD` and `[]` cases.

**`feed`: v2 landed, the document's types did not stay local.** The baseline
session imported v2, passed `slices.Sorted(maps.Keys(counts))` straight into
the document, and wrote no `make` for the wire (the `make(map[string]int)` is
written to, so it stays). It also declared `document` and `eventDoc` at
package level, +5 lines against the reference's function-local `eventDoc` and
anonymous document, and its budget line gave the reason back — "both required
by `Render` to shape the JSON wire document" — which is the reason-slot
pattern `docs/SKILL_AUTHORING_TEMPLATE.md` describes. One session; the
reference arm has written the same shape on other days. Neither `feed` session
wrote a test: at `low` the Contract Table step is skipped in about half the
sessions, as the 2026-09-18 runs recorded.

**Cost and turns lean with the edit, within n=1 noise.** $0.358 against
$0.425 a session, 49 assistant turns against 56, the idiom card read 4/4
through the hook route, lint clean 4/4, `go fix` with nothing to propose 4/4.

## What this establishes

- On the fixture the edit was written against, one `low` session took every
  form the edit names: `encoding/json/v2` with `MarshalWrite`, no
  `writeJSON`, no wire `make`, no default restated, the filter as a call.
- On `feed` the v2 half held and the shape half did not; whether Plain Code's
  "Smallest scope that works" needs the wire document's types named as a
  case ("declared inside the function that marshals them") is a question for
  n=5, not for this session.
- Nothing here is a line-count claim; the corpus README's admission rule
  still applies.

## Next

- `feed,gateway` at `medium` n=5 and `low` n=3 against the same reference,
  reading `Δtypes` on `feed` beside `Δlines`.
- If package-level document types repeat on `feed`, name the case in Plain
  Code and measure that sentence alone.
