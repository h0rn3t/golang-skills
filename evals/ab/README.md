# A/B fixtures for skill wording

`go run ./cmd/abrun` runs one refactoring prompt against these fixtures under
two or more versions of `go-code-refactor/SKILL.md` and reports what changed in
the code the model wrote — not in the prose it produced.

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
skills. `abrun` checks what each arm home actually loads before the first session
and refuses to start if an arm sees the wrong set; for codex the check is
`codex debug prompt-input`, which renders the prompt the model would really see
without spending a request.

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

Any `_test.go` the model wrote is renamed out of the build first. A refactor is
supposed to leave characterization tests behind, and the helper types in them
collide with the golden file's by name; without this step the harness scores its
own collision as a behavior break. It also keeps the model's tests from being
the reason the golden run passes.

Per run: recursive line delta, declared types, interfaces, functions,
pattern-flavored identifiers, whether the package still builds, whether the
golden test passes, session cost, and which `go-*` skills fired. Two guards sit
beside them. A run is invalid if the fixture files are byte-identical
afterwards, because a session that wrote nothing scores a zero delta on every
metric and would otherwise average in as a behavior-preserving tie. A run is
also invalid if its transcript mentions the repository path, since the hidden
golden test lives there and a session that found its way back to the checkout
is measuring nothing. The summary averages structural deltas only over runs
that build, pass the hidden golden test, and clear both guards; everything
excluded stays visible in its own count.

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

Use `no-skill` versus `baseline` only to decide whether a fixture contains a
trap the plugin can catch. It does not measure whether a skill edit improved the
previous skill. For that claim, compare `reference` versus `baseline` with the
same fixtures and save the report.

Runs with a CLI/session error, build failure, or golden failure are not evidence
of a shorter behavior-preserving refactor. Keep them in the report for diagnosis
and exclude their structural deltas from the arm mean.

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

A third replication under `-runner copilot` with `mai-code-1.1-flash` does not
reproduce it, and the reason is the control rather than the skill: unaided, that
model grows `report` by only 16.2 lines against Opus 5's +33.4 and MiniMax M3's
+27.4, declares one type across 20 runs and no interfaces at all. With the bait
untaken there is nothing to remove, every fixture's interval includes zero, and
the corpus-wide direction is marginally against the skill at +1.1 lines per run.
See the [analysis and raw report](../../docs/evidence/2026-09-07-go-refactor-control-mai-code-1.1-flash.md).
The same model is the first to give the implementation corpus a live trap:
`gateway` golden passes went 1/5 without the skill against 3/5 with it, and both
sessions that reached `go-http` set every timeout —
[analysis and raw report](../../docs/evidence/2026-09-07-go-implement-control-mai-code-1.1-flash.md).

A fourth replication under `-runner codex` with `gpt-5.3-codex-spark -effort
xhigh` is the first genuine negative result for the wording. The trap is live
there — unaided, the model grows `report` by 30.6 lines, between Opus 5's +33.4
and MiniMax M3's +27.4 — and the skill did not move it (+31.8), while `dispatch`
got clearly worse. The skill was read in 18 of 20 sessions, so it is not a
triggering failure. Correctness moved the other way: 15 of 20 control sessions
produced a usable refactor against 18 of 20 skilled, with four sessions across
both arms leaving a Go file that does not parse. See the
[analysis and raw report](../../docs/evidence/2026-09-07-go-refactor-control-codex-spark-xhigh.md).
The implementation corpus has not been run on codex; the account's quota ran out
first, and the command to finish it is recorded in that file.

A fifth model, `opencode-go/mimo-v2.5-pro` under `-runner opencode`, gives the
cleanest replication yet and the most useful implementation run to date. On the
refactor corpus it cuts `report` growth 68.0% (+16.4 to +5.2, the only fixture
interval here that excludes zero) and is the first run where both structural
mechanisms move at once, types −50% and functions −60%
([analysis](../../docs/evidence/2026-09-07-go-refactor-control-mimo-v2.5-pro.md)).
On the implementation corpus it is the first model to make three traps live at
once — `gateway`, `feed` and `catalog` all fail unaided, and every failure in
the run is a trap rather than a compile error. `gateway` goes 1/5 to 3/5, the
same numbers as MAI-Code-1.1-Flash; `feed` and `catalog` are tied because the
owning skill never fired
([analysis](../../docs/evidence/2026-09-07-go-implement-control-mimo-v2.5-pro.md)).

A sixth model, `gpt-5.6-luna` at `-effort medium` under `-runner codex`, is the
strongest refactor result recorded and the first on which two fixtures separate
the arms. On `report` the skill does not reduce growth, it removes it — the
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

go run ./cmd/abrun -runner copilot -model mai-code-1.1-flash \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -out copilot.json

go run ./cmd/abrun -runner codex -model gpt-5.3-codex-spark -effort xhigh \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 -out codex.json

go run ./cmd/abrun -tasks report -n 1 -verbose     # one fixture, one pass
```

Needs the `go` toolchain and the agent CLI named by `-runner` on PATH, signed
in to a provider that serves `-model`. Each run is a full headless session, so
cost scales with fixtures x arms x `-n`; the summary prints the mean cost per
run for the arms that reported one.
