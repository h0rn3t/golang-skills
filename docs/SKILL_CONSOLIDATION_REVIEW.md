# Skill Consolidation and Validation

Reviewed on 2026-09-05 against the clean baseline recorded in the
[probe evidence](evidence/2026-09-05-skill-consolidation-probes.json).

## Result and Migration

The pack now has 24 skills, 61 references, 10 scripts, and 5 assets.

| Retired skill | Owner |
|---|---|
| `go-functional-options` | [go-functions](../skills/go-functions/SKILL.md) |
| `go-control-flow` | [go-style-core](../skills/go-style-core/SKILL.md) |
| `go-declarations` | [go-style-core](../skills/go-style-core/SKILL.md) |

Update explicit invocations and remove retired directories from manually
copied installations after replacing the owner directories and references.
No alias skills remain. Installed user copies were not changed by this work.
`go-defensive` and the distinct workflow/domain skills remain separate.

The always-loaded style entrypoint shrank from 220 to 129 lines, or 1,243 to
1,007 whitespace-delimited words. Function design and constructor configuration
share a 137-line entrypoint. These are file-size measurements, not measured
token or latency savings. Detailed syntax loads only when a decision needs it.

Three overlapping initialization, literal, and struct references became one.
Loop semantics and iterator termination have a dedicated reference. Routing,
ownership, README inventories, manifests, and the existing conformance checks
use the current owners. Function-body work alone no longer triggers API design.

Constructor guidance now chooses config structs or functional options by caller
needs, preserves repository conventions, and checks zero/nil/default contracts.
The moved reference also corrects blanket comparability claims using the
[Go comparison rules](https://go.dev/ref/spec#Comparison_operators).

## Behavioral Checks

Claude Code 2.1.261 ran fresh read-only sessions with `claude-opus-5`, medium
effort, isolated plugin snapshots, no hooks, no MCP servers, and no persistent
sessions. The evidence records prompts, observed skill calls, reference reads,
snapshot hashes, answers, and manual assertion assessments.

| Probe | Observed result |
|---|---|
| Baseline native selection | 8/11 against the previous expected skill sets |
| Consolidated native selection | 13/13, including translation controls and shadowing |
| Initial native application | 2/4 against the final semantic criteria |
| Final application with explicit owner-file reading | 4/4 |
| Final native constructor and refactor reruns | 2/2 |

Selection expectations were updated to the new ownership model before running
the consolidated probes. Function parameter grouping selects `go-functions`;
refactoring selects its workflow; a routine new function selects the router.
These scores are not an unchanged-rubric A/B demonstration of model improvement.
Positive selection checks require expected owners and permit additional skills.

The initial constructor answer contradicted its own non-nil metadata guarantee;
the refactor accepted NaN prices after changing `> 0` to `<= 0`. Narrow rules
now require resolving constructor contract mismatches and preserving predicate
semantics. Both final application modes met the relevant criteria. Iterator
checks accept either `break` or `return` for early consumer termination.

The native refactor rerun still showed no explicit read of `go-style-core`,
despite its strengthened route. Its output retained NaN rejection, but that
does not establish reliable reference-following or causality. The explicit
owner-file probes establish application separately from automatic discovery.

## Repository and Example Checks

- `go test -count=1 ./...` from `evals/`, with the installed Go 1.27-compatible
  lint binary on PATH: passed.
- `agentskills-validate@1.0.1`: all 24 skills passed; final changed entrypoints
  also passed focused validation.
- Skill Creator quick validation for both merged owners: passed.
- golangci-lint configuration validation, relative Markdown links, and
  `git diff --check`: passed.
- The final explicit constructor package and iterator/caller compiled and
  passed tests in an isolated module declaring Go 1.23. Config tests came from
  the answer; iterator checks cover filtering, order, early exit, and no result.

The repository now contains 92 trigger and 41 quality scenarios; schema tests
do not execute them against a model. This was a focused Claude probe, with
manual grading, not a full suite or a new GPT-6 behavioral run. Application
passes cover the listed criteria, not every possible generated-code defect.
