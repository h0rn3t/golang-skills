# A/B fixtures for skill wording

`go run ./cmd/abrun` runs one refactoring prompt against these fixtures under
two or more versions of `go-code-refactor/SKILL.md` and reports what changed in
the code the model wrote — not in the prose it produced.

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
pattern-flavored identifiers, whether the package still builds, whether the
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
