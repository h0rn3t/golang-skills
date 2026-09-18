# A/B fixtures for skill wording

`go run ./cmd/abrun` runs one refactoring prompt against these fixtures under
two or more versions of `go-code-refactor/SKILL.md` and reports what changed in
the code the model wrote — not in the prose it produced. Two more corpora
live beside it: [`_implement`](_implement/README.md), where the model fills in
documented stubs and the hidden golden test is the specification, and
[`_review`](_review/README.md), where the model reviews a package with seeded
defects and the review is scored against a hidden key by the source line each
defect sits on.

## The `HEAD` case named, the empty-list case read as body text — Sonnet 5 on `gateway` (2026-09-18)

[`HEAD` report](../../docs/evidence/2026-09-18-go-implement-head-case-gateway-sonnet-5.md),
[`null` report](../../docs/evidence/2026-09-18-go-implement-null-body-gateway-sonnet-5.md):
twenty `gateway` sessions of the evening read against the model's own test
file gave the mechanism behind every Sonnet 5 golden miss — a `HEAD` case
in the contract test means the edit hook prints the failure and the model
fixes it; no case, no fix; every `null` miss decoded the body and could not
see it. Two edits, each measured on `gateway` alone against the tree before
it, `medium` n=5 and `low` n=3: the `HEAD` case named with its scope and the
`go-http` rule made a count — `HEAD` in the code 8/8 against 5/8 (`low` 3/3
against 0/3), golden 5/5 against 4/5 and 2/3 against 0/3, cost level at
`medium`; the body-text sentence — 7/8 against 5/8 tests compare the body
with `[]`, the two remaining `null` misses in `low` sessions that wrote no
empty-list case. One `medium` session reported `test pass (hook)` after the
hook had printed a failing test and nothing cleared it.

## The idiom card through the host route, Sonnet 5 medium at n=3 plus n=2 (2026-09-18)

[Report](../../docs/evidence/2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n3-sonnet-5-medium.md):
the 1.21.0 branch head against the same tree with `go-prompt-routing.sh`
naming the card's installed path in its load sentence and
`go-code-routing.sh` requiring one whole `Read` of it before the first `.go`
edit, on `roster`, `feed`, `gateway`, 18 then 12 sessions. Card read 15/15
against 1/15 — 14 from the note, 1 from the gate; golden 14/15 against
15/15 (the `gateway` `HEAD` clause); +15% cost ($0.424 against $0.367,
p ≈ 0.30), about half of it the card's own tokens; the note's first wording,
a sentence of its own, spread the loads (3.22 Skill turns against 2.44),
folded into the load sentence 2.67 against 2.33. `roster` is saturated on
this model (no older form in 10/10 either way), so the read is the result,
not a code effect. At [`low` effort](../../docs/evidence/2026-09-18-go-implement-card-hook-route-roster-feed-gateway-n2-sonnet-5-low.md)
(n=2) the note is followed the same way, 6/6 against 0/6, at +10%; the
route arm met the gate for `go-style-core` in 1/6 sessions against 6/6, and
`gateway` failed 0/4 in both arms. The [third seed of the 1.20.1 comparison](../../docs/evidence/2026-09-18-go-implement-card-in-step2-seed3-roster-feed-gateway-n3-sonnet-5-medium.md)
the same evening turned the afternoon's +8% cost lean to −6%; pooled over
21 sessions a side the 1.21.0 tree is +2% (p ≈ 0.86), golden 19/21 against
21/21 on two `gateway` clauses.

## The review checklist split, Sonnet 5 medium at n=2 (2026-09-18)

[Report](../../docs/evidence/2026-09-18-go-review-cuts-only-vs-restated-n2-sonnet-5-medium.md):
the eleven cuts alone against the cuts with the seven restatements, six
fixtures, 24 sessions. Recall 117 against 116 of 152, must 57 against 55 of
62, cost $0.258 in both arms; the afternoon's 0.72 → 0.80 was the cuts. The
same bytes scored 0.80 at seed 1 and 0.76 at seed 2: ±0.04 recall is this
corpus's run-to-run movement at n=2. One row to watch: `invoice/middle-man`
found 2/2 under "No premature interfaces" and 0/2 under "Interfaces where
they are consumed".

## The Orient move on the refactor corpus, Sonnet 5 medium at n=3 plus `report` at n=5 (2026-09-18)

[Report](../../docs/evidence/2026-09-18-go-refactor-orient-move-n3-sonnet-5-medium.md):
release 1.20.1 against the 1.21.0 branch head — the `REFACTOR_SKILL_DIR`
paragraph moved into Orient, the `loc-diff`, bugs-found and hook-record
sentences restated — on the four fixtures at n=3 (24 sessions) and on
`report` alone at n=5 (10 sessions). Golden 17/17 and lint identical in both
arms, one load of the refactor skill per session in 34/34, cost −10% and +4%
in the two runs. Lines −10.8 against −15.5 at n=3 with the whole gap in
`report` (four helpers in 3/3 sessions against 1/3), which at n=5 reversed to
+7.2 against +9.0 (p = 0.75): the fixture is bimodal on this model, a flat
`Render` or three to four write helpers, and eight sessions a side (p = 0.30)
do not settle it. Shell-less sessions stated a line count anyway in 7/12
against 6/12, so the `loc-diff` restatement is neutral there; the idiom card
was read in 1/17 and 0/17 refactor sessions.

## The idiom card as a workflow step, Sonnet 5 medium and Haiku 4.5 (2026-09-18)

[Sonnet report](../../docs/evidence/2026-09-18-go-implement-card-step-roster-feed-gateway-n2-sonnet-5-medium.md),
[Haiku report](../../docs/evidence/2026-09-18-go-implement-card-step-roster-n5-haiku-4-5.md):
`reference` (1.20.1) against the working tree, twice — the card read first as
a numbered step of its own, then folded into step 2 of `go-code` — on
`roster`, `feed`, `gateway` at n=2 (Sonnet 5 medium, 24 sessions) and on
`roster` at n=5 (Haiku 4.5, 20 sessions). Sonnet 5 read the card in 0/24
sessions under every wording; the separate step made it skip the load steps
as a block in 2/6 sessions (six gate blocks against one) and the step-2 form
restored the 1.20.1 shape (2.00 Skill turns a session, no gate block). Haiku
read it in 5/10 reference sessions, 3/5 with the separate step, 4/5 with
step 2; every session that read it wrote no older form (12/12), two of eight
that did not wrote `sort.Strings`. Golden 11/12 against 12/12 on Sonnet (the
`gateway` nil-list `null`), 10/10 both arms on Haiku. The same tree carried
`go-style-core` "The Edit Hook Record"; shell-less Sonnet sessions marked
every claimed check `(hook)` in 8/11 reports against 1/11. Cost leaned
against the tree on Sonnet (+8% over both runs) within the reference's own
run-to-run movement. [Opus 5 medium](../../docs/evidence/2026-09-18-go-implement-hook-record-feed-gateway-n2-opus-5-medium.md)
on `feed` and `gateway` at n=2: neutral on every axis, $0.989 against
$1.037 a session, card read 8/8 in both arms, `(hook)` marking 4/4 against
3/4.

## The load step and the three-router gate, three arms at n=1 (2026-09-13)

[Report](../../docs/evidence/2026-09-13-go-refactor-routing-load-n1-sonnet-5-medium.md):
Sonnet 5 medium, `no-skill`, `reference` (1.17.0) and `baseline` (the working
tree where `go-code-refactor` and `go-code-review` say to load `go-style-core`
and the owners before the first edit and `go-code-routing.sh` gates the first
`.go` edit after any of the three routers), one repetition per fixture and
arm, 12/12 valid. `go-style-core` before the first edit 3/4 from the text and
4/4 with the gate, against 0/4; an owner skill 4/4 against 1/4; five gate
fires in three sessions, every retry landed. Golden 4/4 and lint-clean 3/4 in
both skilled arms; −13.2 against −16.0 lines, the gap the `store` session
that kept the nil-map guards again. The loads arrived one per turn, not in
one message, and cost rose +87% ($0.348 against $0.186 a session; +62%
without the `pricing` session that wrote a 167-line characterization test).
A firing smoke, not a line claim. Follow-up the same night
([report](../../docs/evidence/2026-09-13-go-refactor-routing-cost-n1-sonnet-5-medium.md)):
with the no-shell paragraph in `go-code-refactor`, the one-message clause
gone, and the prompt hook naming `go-style-core` and the owners the target's
code points at, `reference` (1.17.0) against `baseline`, one repetition per
fixture, 8/8 valid: $0.206 against $0.228 a session (the morning's baseline
was $0.348), the refactor skill loaded once per session against six loads in
four reference sessions, no post-edit loads, two gate blocks against five in
the morning; golden 4/4 and lint-clean 3/4 in both arms, −11.2 against −13.2
lines. The dollar gap is n=1 noise — the reference arm itself moved from
$0.186 to $0.228 between the runs — the counts are the result.

## The idiom card on `roster`, Haiku 4.5 at n=5 (2026-09-13)

[Report](../../docs/evidence/2026-09-13-go-implement-roster-current-go-card-n5-haiku-4-5.md):
Haiku 4.5, `no-skill`, `reference` (1.16.0) and `baseline` (the working tree
with `go-style-core/references/CURRENT-GO.md`, the write-time idiom card),
five repetitions on `roster`, 15 sessions, 14/15 valid. The fixture
discriminates here where it could not on Sonnet 5 medium: the unaided arm
copied `sort.Strings` from the pre-1.21 neighbor in 5/5 sessions (one without
the import, so it did not build), the normative rule alone left it in 2/5,
the card in 0/5 — `slices.Sort` 5/5, `slices.ContainsFunc` 4/5,
`slices.Compact` 2/5 — with golden 5/5 in both skilled arms and +27% cost
over the reference. The card reached two sessions that loaded only `go-code`,
through its routing line. One card session modernized `legacy.go`, the file
the fixture declares off limits; behavior held, so only the `go fix` column
reading zero shows it. `map[string]bool` as a set moved nowhere (5/5, 5/5,
3/5).

## Architecture reference and the current-Go rule, three arms at n=1 (2026-09-13)

[Report](../../docs/evidence/2026-09-13-go-arch-current-go-three-arm-n1-sonnet-5-medium.md):
Sonnet 5 medium, `no-skill`, `reference` (1.15.0) and `baseline` (the working
tree with `go-code-refactor/references/ARCHITECTURE.md` and the normative
"Write Current Go"), one repetition per fixture and arm, 12/12 valid. Golden
4/4 and lint-clean 3/4 in both skilled arms, −14.2 against −11.5 lines, the
gap one `store` session that kept two nil-map guards the reference session
deleted; cost +3%. Every skilled session read all ten references, so
`ARCHITECTURE.md` (~7.3K tokens) was read 4/4 on single-package fixtures. A
smoke, not an effect. Repeated the same afternoon at n=5 on `report` and
`store` after the reference was rewritten and split into three files
([report](../../docs/evidence/2026-09-13-go-arch-current-go-n5-report-store-roster-n3-sonnet-5-medium.md)): baseline −3.0 lines against reference (p = 0.50),
golden and lint-clean 10/10 in both arms, −17% cost, all three files read in
10/10 baseline sessions — the morning's gap was noise.

## Sonnet 5 medium control, both corpora at n=5 (2026-09-10)

Refactor [report](../../docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium-n5.md),
implement [report](../../docs/evidence/2026-09-10-go-implement-control-sonnet-5-medium.md):
the n=1 medium run below repeated at five repetitions per cell on release
1.7.0, plus the implementation corpus under identical conditions and the same
arm digest. 80 sessions, no CLI errors.

The refactor corpus is the strongest result recorded on a Claude model here.
Corpus difference −9.7 lines, reproducing the n=1 reading of −9.6, with
correctness tied at 20/20 build and golden in both arms. **Three of four
fixtures separate**: `dispatch` −15.8 with arms that do not overlap
(p = 0.00794, 0.032 after Bonferroni), `report` −9.3 (p = 0.00794) and
`pricing` −9.0 (p = 0.024, exploratory after correction); `store` is a tie.
`dispatch` shows the mechanism — four of five control sessions extract
`put`/`del`/`eventKey` helpers and grow the package, while the skilled arm
folds the branches into one path and adds 0.20 functions. New functions fall
1.45 → 0.58 per session and the concision gate passes 16/20 against 10/20.
Cost 4.38x.

The implementation corpus separates nothing: −1.78 lines across the three
evaluable fixtures, correctness 13/15 against 14/15, no fixture under p = 0.05,
and `Δiface`/`Δpattern` zero in all 40 sessions. `gateway` is still unusable on
this model at this effort — 8 of 10 sessions fail the HEAD/405 contract in both
arms, including 5/5 skilled sessions in which `go-http` loaded and which carries
the rule since `f12c73b`. Three of those skilled sessions also rendered the
empty account list as `null`. Cost 4.54x. The pair is the same statement the
`gpt-5.6-luna` runs made: the plugin pays for removing structure from working
code, not for filling an empty body.

## Sonnet 5 high control (2026-09-10)

[Report](../../docs/evidence/2026-09-10-go-refactor-control-sonnet-5-high.md):
the medium run below with one flag changed, `-effort high`. 8/8 valid,
correctness tied at 4/4, routing 4/4 and deeper — eleven skill loads against
five. The corpus difference collapses from −9.6 to −1.8 because both arms
moved toward each other: the unaided control gained 3.3 lines of concision and
the skilled arm lost 4.6. On `report` the medium skilled arm held the package
at its original size while the control grew 14; at high **both arms grew 12
lines and three functions**. Effort is therefore a condition a control has to
fix and name, not a detail — the same model and tree on the same day give −9.6
and −1.8. Cost is 5.60x, the highest recorded here. n=1, so read the three
2026-09-10 runs as a range to test at n=5, never as effects: they disagree with
each other by more than any of them can measure.

## Sonnet 5 medium control (2026-09-10)

[Report](../../docs/evidence/2026-09-10-go-refactor-control-sonnet-5-medium.md):
Claude CLI with `-effort medium`, n=1 per fixture and arm, 8/8 valid, and the
same-conditions pair to the Haiku run below — same seed, tree and digest, one
day, only the model differs. `go-code-refactor` fired 4/4 against Haiku's 1/4,
which is what makes Haiku's number a property of the tier and not of the arm.
One repetition supports no effect, but the direction is clean: −15.8 lines
against −6.2, three fixtures moving the same way and none against, the gate at
4/4 against 2/4, correctness tied at 4/4, and `report` — the over-engineering
trap — rewritten to the same 50 lines while the control grew 14. Cost is 3.94x.

## Haiku 4.5 control (2026-09-10)

[Report](../../docs/evidence/2026-09-10-go-refactor-control-haiku-4-5.md):
Claude CLI, n=1 per fixture and arm, 8 sessions, 7/8 valid. One repetition
supports no structural claim; what it establishes is that `go-code-refactor`
reached **1 of 4** baseline sessions on this tier, against 19/20 on Opus 5 and
20/20 on codex, with the arm loading correctly in every one of them. A wording
comparison on Haiku is measuring the router until that changes. Behavior held
4/4 in the skilled arm against 3/4 in the control, and this is the first run
carrying `fix_hunks`: `dispatch`'s one pending modernization was removed by both
arms and no session introduced another. Its control pair is the Sonnet 5 run
above.

## Opus 5 medium control (2026-09-08)

[Report](../../docs/evidence/2026-09-08-go-refactor-control-opus-5-medium.md):
Claude CLI with `-effort medium`, n=5 per fixture and arm, 39/40 valid — one
baseline `report` run failed the test it wrote itself. Mean production-line
difference is −3.95; new functions total 26 → 10, types 12 → 9. `pricing`
carries it at −9.2 lines (p = 0.0556 unadjusted, 0.22 with Bonferroni), while
`store` and `dispatch` are flat or slightly against the skill — so the fixture
producing the corpus effect is not the one that produced it on sonnet 5. Cost is
2.94x, and the two-arm run does not isolate a wording change.

## Sonnet 5 control (2026-09-08)

[Report](../../docs/evidence/2026-09-08-go-refactor-control-sonnet-5.md): Claude
CLI, n=5 per fixture and arm, 40/40 valid. Mean production-line difference is
−3.20; new functions total 34 → 20, types 6 → 10. `store` improves by 8.2 lines,
while `report` grows 2.8 more. The `store` permutation p = 0.03968 is unadjusted
(0.15873 for four comparisons with Bonferroni); investigate it as an exploratory
signal. Cost is 3.28x, and the two-arm run does not isolate a wording change.

## Arms

`no-skill` loads no plugin — the discovery control. `baseline` is the complete
skill tree currently checked out. Supplying `-reference-root` adds a
`reference` arm from another complete plugin checkout, which is the arm to use
for a before/after comparison. Every `variants/*.md` file becomes one more arm,
spliced into the current `SKILL.md` in front of the `## Workflow` heading.
`-arms no-skill,baseline` runs a subset.

A fixture earns its place only when `no-skill` is measurably worse than
`baseline` — that gap is the trap the fixture exists to catch, and it is what
any wording change has room to move. Name the trap in the table below before
adding one.

| Variant | What it tests |
| --- | --- |
| `pattern-sentence.md` | "Apply a suitable design pattern, but only if it improves the solution." — an instruction whose condition the model itself judges |
| `pattern-gate.md` | The same intent as a gate with externally checkable conditions: three existing duplicate sites, a shorter diff, no call site that reads worse |
| `selection-once.md` | A completion criterion instead of a permission: each literal once, each selection over the same key once, each condition ladder once; shape chosen by the final code, table not to be serviced. Measured on `pricing` + `report`, 10 reps per arm: opencode +0.1 lines vs baseline, interval −7.7 … +7.9, where baseline had no gap ([report](../../docs/evidence/2026-09-07-selection-once-luna-opencode.uk.md)); copilot **−11.6**, interval −18.9 … −4.3, where baseline stopped at `switch` in 4/10 sessions ([report](../../docs/evidence/2026-09-07-selection-once-luna-copilot.uk.md)). `report` stayed at −1 on both. Ported into `go-code-refactor/SKILL.md` on 2026-09-08 as «Remove Duplication to the End», with the «once, not once under a name» clause added; the arm is now redundant against the current baseline and stays for `-reference-root` comparisons against `1e98819`. Codex with the ported text as baseline: −12.4 vs no-skill, interval −18.0 … −6.8, tables 9/10; the double-text arm was +5.7 vs baseline, interval crossing zero ([report](../../docs/evidence/2026-09-08-selection-once-luna-codex.uk.md)) |
| `whole-transformation.md` | Evaluate a helper or data table together with the duplication it removes; the [pricing experiment](../../docs/evidence/2026-09-07-pricing-whole-transformation-luna-opencode.uk.md) did not establish an improvement |

## Runners

`-runner claude` is the default and loads the arm's plugin through
`--plugin-dir`. `-runner opencode` drives the `opencode` CLI, which has no
plugin flag: it discovers skills from HOME, so every arm gets its own HOME with
that arm's `skills/` copied in. `-runner copilot` drives the GitHub Copilot CLI,
which discovers personal skills from `COPILOT_HOME`, so every arm gets its own
one of those instead — HOME itself is left alone there, because the credential
store the CLI authenticates against is keyed to it and a redirected HOME fails
every run unauthenticated. The `no-skill` arm needs an isolated home just as much
as the others — run under the operator's own it would load whatever `go-*` skills
they have installed globally and stop being a control. `-runner codex` drives the
OpenAI Codex CLI, which needs HOME redirected rather than `CODEX_HOME`: Codex
also discovers skills from host locations outside its own home, so a control arm
with only `CODEX_HOME` moved still sees the operator's globally installed
skills. `abrun` checks what each arm home actually loads before the first
session and refuses to start if an arm sees the wrong set.

### Arm preconditions

Two guards run before the first session, and both refuse to start the run rather
than report a difference that turns out to be a missing skill.

**Every SKILL.md must parse.** Each materialized arm is scanned for the failure
mode that costs the most and shows the least: a frontmatter value with an
unquoted colon-space, which YAML reads as a nested mapping and rejects. Every
loader drops such a skill, most of them quietly. This guard runs for all four
runners, including claude, which has no listing command to check an arm against,
and it costs nothing — no CLI is involved.

**Every skill in the tree must actually load.** For the three runners that
discover skills from a home, `abrun` asks the CLI what it loaded and compares the
result against the arm's own `skills/` directory, naming any skill that is
missing. The earlier version of this check only rejected an *empty* listing,
which is how a broken `go-code` passed it in a 24-skill tree; the set is what
matters, not the count. The control arm must load none, or skill discovery is not
isolated. For copilot the CLI's stderr is captured and quoted into the error,
because that is where it explains what it could not parse while still exiting
zero. For codex the listing is `codex debug prompt-input`, which renders the
prompt the model would really see without spending a request.

Five differences between the runners are not cosmetic and belong in any claim
that spans them.

**How a skill fires.** claude, copilot and opencode all give the model a skill
tool, and `abrun` records the calls. Codex has none: skills arrive as a listing
in the system prompt and a skill fires when the model reads its `SKILL.md`
through the shell, so the codex runner scores a skill from the command that
opened it — and only from the command, never from its output, because a loaded
`SKILL.md` names other skills and scanning the output would credit every one it
mentions.

**Tools.** The claude arm is restricted to `Skill,Read,Glob,Grep,Edit,Write` and
the copilot arm to the same surface under copilot's names
(`skill,view,create,edit,grep,glob`), while opencode and codex keep their own
tool sets, so only those sessions have a shell and can run `go test` on their
own work.

**Plugin parts.** The repository's PostToolUse gofmt/vet hook and its
`go-verify` subagent apply only under claude; the copilot, opencode and codex
homes carry skills alone.

**Cost.** Only claude and opencode report a session cost in dollars. Copilot
bills in premium requests and AI credits and codex reports token counts, so
their `$/run` column is empty rather than converted.

**Step ceiling.** `maxSteps` bounds a session only where the CLI can express one
(`--max-turns` for claude, the build agent's step count for opencode); copilot
and codex sessions are bounded by `-timeout` alone.

Compare arms within one runner; compare runners only with the differences
stated.

`-effort` sets the reasoning effort level and is accepted only by the runners
whose CLI can ask for one, codex and copilot. It is rejected rather than ignored
elsewhere, because a report that records `xhigh` for a run served at the model's
default is a report that lies.

## Fixtures

Each is a package where the honest refactor **removes** structure and the
tempting one **adds** it. Fixture files carry no hints — the model reads them,
so the intent is documented here instead.

| Fixture | Honest refactor | The temptation |
| --- | --- | --- |
| `dispatch` | One switch plus a shared put helper | A handler registry or per-kind Strategy for three branches that never leave the file |
| `report` | Early returns plus one row builder | A `Formatter` interface with `TextFormatter`/`CSVFormatter` for two formats no caller extends |
| `store` | Flatten the nesting, delete the duplicate miss branch | A `Repository` interface plus adapters over two plain maps |
| `pricing` | One plan table the three lookups share | Five plan types behind a `PricingStrategy` interface and a registry — the hardest bait: the plan list is duplicated across two files and a `TODO` promises a sixth plan next quarter |

## What gets measured

`_golden/<fixture>/golden_test.go` is copied in **after** the model finishes and
pins the observable behavior: exported signatures, error texts, key formats,
column widths. It never reaches the model, so it cannot be adjusted to match a
changed implementation — a failing golden test means behavior moved.

Before installing the golden tests, the harness runs `go test -count=1 ./...`
on the model's tree when it contains test files. The JSON records `model_tests`
as `pass`, `fail`, or `skipped` (no test files), with failure output in
`model_test_failure`. Older reports omit these fields: that is unmeasured,
not a historical pass. Each `_test.go` is then renamed to `_test.go.model`
before the independent golden run, even if the model's tests failed. This
avoids helper-name collisions and prevents model tests from making the golden
run pass. A passing golden result cannot override a model-test failure.

A failing golden run has two possible causes and they must never be averaged
together. `behavior_failure` is an overlay that compiled and then failed an
assertion: the observable contract moved. `harness_failure` is an overlay that
never reached an assertion, because a name it declares is also declared by the
model's production code — the compiler says `redeclared in this block`, and the
run carries no verdict about behavior at all. It prints as `HRN` rather than
`ERR`, stays out of every mean, and is a harness fault to fix, not a regression
to attribute. Every package-level name a golden declares therefore contains
`golden`, and `TestGoldenHelpersAreCollisionResistant` fails the build if a new
one does not. Neither field is set when the package failed to build on its own:
that is already `build`, and the golden result means nothing after it.

### The repair loop

`-repair` grants one extra turn, and only when the first one missed: after the
session the harness measures the package, replays the golden, and if either the
line gate or the golden failed it returns the exact numbers and the assertion
text, then lets the model act once more. `repair_fired` records that the second
turn happened, `pre_repair` and `pre_repair_golden` the readings that triggered
it, and `repair_feedback` the text the model was actually handed. Every recorded
number describes the tree after the last turn.

Two properties make the loop an experiment rather than a leak. The probe runs
against a throwaway copy of the whole module, so the golden never enters the
tree the model can read, and the model's own tests stay where it left them —
`TestProbeGoldenLeavesTreeUntouched` fails if either stops holding. And the
failure reaches the model as assertion text with file positions stripped: it is
told what broke, not which file holds the test.

The feedback names its counting convention on purpose. Stage 2 measured a
session that read the same file as 101 lines where the harness read 138 —
non-blank non-comment against physical — declared its gate met and changed
nothing. The disagreement was about the definition, not the code, so the
definition travels with the number.

### Production LOC

`lines` is the number every published line claim has to mean: **physical lines
in the fixture package's non-test `*.go` files**, blank lines and comments
included, a trailing line without a newline counted once. No file named
`*_test.go` contributes, so a session can neither shrink the number by moving
code into a test nor grow it by writing one. `Δlines` is the after-count minus
the before-count under that definition, and `line_gate_pass` is that delta being
at most zero — the concision gate's own criterion, recorded whether or not the
run was valid, because a gate met while behavior broke has to stay visible.

Per run: recursive line delta, declared types, interfaces, functions,
pattern-flavored identifiers, modernizations `go fix` still proposes,
whether the package still builds, whether the
golden test passes, session cost, and which `go-*` skills fired. Two guards sit
beside them. A run is invalid if the fixture files are byte-identical
afterwards, because a session that wrote nothing scores a zero delta on every
metric and would otherwise average in as a behavior-preserving tie. `empty_diff`
narrows that to the case worth reading: a session that ran to completion without
an error and deliberately left every file alone, which is what a gate permitting
an empty diff is supposed to produce and is not the same as a session that
failed before it wrote anything. A run is
also invalid if its transcript mentions the repository path, since the hidden
golden test lives there and a session that found its way back to the checkout
is measuring nothing. The summary averages structural deltas only over runs
that build, pass model tests when present, pass the hidden golden test, and
clear both guards; everything excluded stays visible in its own count.

A wording that helps shows up as fewer lines with the golden test still green.
A wording that licenses growth shows up as `Δtypes`, `Δiface` and `Δpattern`
rising — that is the number the "only if it improves the solution" phrasing was
added to move.

### Pending modernizations

`fix_hunks` is the number of unified-diff hunks `go fix -diff` still proposes
for the fixture package after the last turn, and `fix_hunks_before` the same
count on the fixture as shipped. It is the one reading of "reaches for what the
toolchain already ships" that costs nothing and needs no judgment: the analyzers
decide, not a rubric. Hunks in a `*_test.go` file are skipped, so the number
follows the same production-only rule as `lines`. Zero means no analyzer has
anything left to say about the code in the tree; it says nothing about what no
analyzer covers — `cmp.Or`, `errors.Join`, `iter.Seq` — so read it as a floor
and not as a modernity score.

`dispatch` ships with one pending hunk (`for i := 0; i < len(events); i++`);
the other three refactor fixtures ship with none. So the corpus asks two
questions of every arm: does a session that touches that loop leave it modern,
and does any session introduce a construct the toolchain would undo?

Both fields are absent, and `fix_hunks_unmeasured` says which reading failed and
why, when the diff could not be taken. `go fix -diff` exits non-zero exactly
when the diff is not empty, so the exit status carries no error information: a
package that does not type-check produces an empty diff on stdout with the type
errors on stderr, and recording that as zero would read as a fully modern tree.
The summary averages the pair over the valid runs that could be read at both
ends and counts the ones left with nothing to propose; a run that could not be
read stays out of the mean.

## Evidence contract

The repository carries no result table without its raw JSON report. A published
comparison records the exact runner, model, seed, prompt, repetitions, both
complete plugin roots, and the revisions or content hashes they came from.
`abrun` records a SHA-256 digest of every materialized plugin arm in the JSON
report; record the source revisions beside it when publishing the file.

Use `-keep` for comparisons: it now retains **every** run, including successful
ones. The JSON `workdir` and console `source:` line locate the scratch tree;
production source remains in place, model tests have the `.model` suffix,
and the golden files are added separately. `trace.jsonl` at the root of that
tree is the raw session transcript, reported as `trace_path`. It is what
separates a measurement from a claim about one: `reported_counts` says the final
message stated a line count, `commands` says how many shell calls the session
actually made (codex only — the claude arms are granted no shell), and the trace
says which ones. A report that cites a model's stated delta without the trace
behind it is citing prose. Archive these trees alongside the
report before temporary-directory cleanup. Without `-keep`, scratch trees are
removed. Historical reports made with failure-only `-keep` cannot recover
successful source retroactively.

Use `no-skill` versus `baseline` only to decide whether a fixture contains a
trap the plugin can catch. It does not measure whether a skill edit improved the
previous skill. For that claim, compare `reference` versus `baseline` with the
same fixtures and save the report.

Runs with a CLI/session error, build failure, model-test failure, or golden
failure are not evidence of a shorter behavior-preserving refactor. Keep them
in the report for diagnosis and exclude their structural deltas from the arm mean.

## Current validated results

The 2026-09-07 Opus 5 control completed 40/40 valid runs. Across the corpus the
skill produced 4.15 fewer production lines per run and 60% fewer new types. On
the only strong over-engineering trap, `report`, it reduced mean growth from
+33.4 to +16.8 lines (49.7%). Correctness was tied at 20/20 build and golden
passes in both arms. See the [full analysis and raw report](../../docs/evidence/2026-09-07-go-refactor-control-opus5.md).

The implementation corpus has one admitted fixture. On Opus 5 with `n=5`,
`gateway` separates the arms completely on what a working implementation costs:
152.6 lines without the skill against 99.8 with it, control at 109/124/169/170/191
against 94/97/100/101/107, with functions down 44% and branches down 38% and
correctness tied at 5/5. The skilled arm's spread is also seven times smaller.
See [`_implement/README.md`](_implement/README.md) and the
[full analysis](../../docs/evidence/2026-09-07-go-implement-gateway-opus5.md).

The same refactor corpus under `-runner opencode` with `opencode-go/minimax-m3`, also
2026-09-07, completed 40/40 valid runs. It reproduces the direction and, on
`report`, the size of the effect: +27.4 to +11.0 lines (59.9%), and −16.4 lines
against Opus 5's −16.6. Function growth fell 48.4%; type growth did not
reproduce. Correctness was tied again at 20/20 in both arms. See the
[analysis and raw report](../../docs/evidence/2026-09-07-go-refactor-control-minimax-m3.md),
including the two tool-set differences that keep the two files from being a
clean model-to-model comparison.

A harness lesson came out of a copilot re-run, and the guard it produced is
described under [Arm preconditions](#arm-preconditions). The first attempt at
that run was invalid and looked normal: an unquoted colon-space in the
rewritten `go-code` description made its YAML frontmatter unparseable, copilot
dropped the skill from its listing and reported it on stderr with a zero exit
status, and the arm pre-check passed because it only asserted that a baseline arm
loads *at least one* `go-*` skill. Eighty sessions ran without the router. A 0%
firing rate has two causes, model routing and a skill that never loaded; the
pre-check now separates them before the first session, but comparing the firing
rate against the previous run is still the cheapest way to notice that something
moved.

`gpt-5.6-luna` at `-effort medium` under `-runner codex` is the strongest
refactor result recorded and the only one on which two fixtures separate the
arms. On `report` the skill does not reduce growth, it removes it — the
control writes +18.8 lines, the skilled arm +1.2, and four of its five sessions
return a *smaller* package with the golden test still green — and `dispatch`
moves the right way for the first time on any model, −5.8 with an interval that
excludes zero. Helper growth falls 71.0%, correctness is tied at 20/20, and
`go-code-refactor` reached every one of the 20 baseline sessions
([analysis](../../docs/evidence/2026-09-07-go-refactor-control-gpt-5.6-luna-medium.md)).
Its implementation run is the opposite and belongs beside it: every trap
saturated at 20/20 in both arms and no fixture separating, despite the best
skill routing the corpus has recorded
([analysis](../../docs/evidence/2026-09-07-go-implement-control-gpt-5.6-luna-medium.md)).
The pair is the clearest statement of where the plugin pays — removing structure
from code that already works, not adding code to an empty body.

That implementation run was repeated on 2026-09-08 against the `go-code` rewrite
and the extraction rule moved into `go-code-refactor`, same prompt, seed,
fixtures, runner and model: 40/40 valid, golden 20/20 in both arms again, corpus
size −0.95 lines with every interval including zero
([analysis](../../docs/evidence/2026-09-08-go-implement-control-gpt-5.6-luna-medium.md)).
The only thing that moved is routing — `go-http` reached in 4 of 5 `gateway`
sessions against 2 of 5, and four other owner skills in half again as many
sessions — with no change in what the sessions produced. Firing rate and
measured outcome are separate results, and improving the first does not deliver
the second on a model whose control arm already passes every trap. The control
itself drifted −1.3 lines across the corpus between the two days, small enough
here to reproduce rather than replace the earlier file; a control is a property
of the served model version, so date every one of them and re-measure before
comparing a skill edit across runs rather than inheriting the number.

## Running it

```bash
# Discover whether the current plugin beats no skill on these fixtures.
go run ./cmd/abrun -arms no-skill,baseline -n 3 -j 4 \
  -model claude-opus-4-1-20250805 -seed 1 -out control.json

# Compare a previous complete checkout with the current plugin tree.
go run ./cmd/abrun -reference-root ../golang-skills-before \
  -arms reference,baseline -n 5 -j 4 \
  -model claude-opus-4-1-20250805 -seed 1 -out before-after.json

# Replay the same corpus under a different agent and model.
go run ./cmd/abrun -runner opencode -model opencode-go/minimax-m3 \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -out opencode.json

go run ./cmd/abrun -runner copilot -model <copilot-model> \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -out copilot.json

go run ./cmd/abrun -runner codex -model gpt-5.6-luna -effort medium \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -out codex.json

go run ./cmd/abrun -tasks report -n 1 -verbose     # one fixture, one pass
```

Needs the `go` toolchain and the agent CLI named by `-runner` on PATH, signed
in to a provider that serves `-model`. Each run is a full headless session, so
cost scales with fixtures x arms x `-n`; the summary prints the mean cost per
run for the arms that reported one.
