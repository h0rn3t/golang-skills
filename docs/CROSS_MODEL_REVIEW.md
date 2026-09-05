# GPT-6 and Claude skill review

Reviewed on 2026-09-05 against baseline commit `c68fec7`. Scope: the shared Go skill instructions, router,
verification agent, authoring guidance, and behavioral evals. This is a prompt
and host-portability review, not a fresh audit of every Go API recommendation.

## Sources and approach

Read the current official [GPT-6 Astra prompting guidance](https://developers.openai.com/api/docs/guides/latest-model/gpt-6-astra.md#prompting-best-practices)
and [Codex skill documentation](https://learn.chatgpt.com/docs/build-skills).
The relevant model guidance concerns honoring user intent, explicit instruction
priority, useful progress reports, host-controlled delegation, and proportionate
verification. Codex supports skill discovery independently of Claude's plugin
agent and hooks. Keep one shared instruction set; isolate host-specific metadata.

## Findings and corrections

| Priority | Before | Correction and owner |
|---|---|---|
| P1 | `go-verify` interpreted "check it builds" as authorization for build, vet, modernization, lint, and race tests. | [Agent](../agents/go-verify.md) selects the requested check and names its scope. Build success cannot be reported as full-gate success. |
| P1 | The restraint reference said to ship a smaller default and make the user request the full version again, despite another section preserving explicit requirements. Quality eval 21 even required rejecting a specifically requested shared package. | [Restraint rules](../skills/go-code-refactor/references/OVER-ENGINEERING.md#ship-then-question) simplify implementation while retaining the whole requested deliverable. Eval 21 now respects the requested package and reports unavailable execution honestly. |
| P1 | Refactoring with no tests or a red baseline automatically ended the turn for approval, including work already authorized. | [Refactor](../skills/go-code-refactor/SKILL.md#resolve-baseline-and-scope) establishes characterization coverage when needed, investigates existing failures, and continues independently verifiable work. Unresolved contract or source ambiguity still warrants a question. |
| P2 | Gate and agent used competing `unavailable`/`skipped` terminology and offered only PASS/FAIL totals. Missing infrastructure was automatically called environmental based on the message. | [Gate](../skills/go-linting/SKILL.md#verification-gate) defines `INCOMPLETE` for missing required evidence; attribution requires evidence. Agent uses the same states. |
| P2 | The router required a full gate per checklist item; linting required rerunning it after every category and unioning local and default gates when uncertain. | Reuse evidence for unchanged work, honor the local gate, and rerun affected checks when something relevant changes. Required repository checks remain required. |
| P2 | Router commands assumed a checkout-local `skills/` directory and an active PostToolUse hook. It also automatically loaded an optional external `ponytail` skill. | [Router](../skills/go-code/SKILL.md#resource-routing) resolves installed resources, handles missing siblings, and counts only observed hook results. Bundled restraint guidance has no external skill dependency. |
| P2 | Agent could substitute a non-race `make test` for race coverage; its dependency detection missed staged changes. The gate's command substitution could select the current package for an empty diff. | Inspect the test recipe, include relevant staged/untracked changes, pass explicit existing package paths, and skip modernization when that set is empty. |
| P2 | Style instructions categorically banned delegated reviews, limited progress updates to plan changes, and allowed unrelated whole-file modernization. | [Style owner](../skills/go-style-core/SKILL.md) follows user and host instructions, keeps useful progress updates, and confines modernization to scope and the module's supported Go version. |
| P2 | Automated evals invoke `claude` and a Claude judge exclusively; schema tests do not test GPT-6 behavior. | Document that boundary, add portable quality prompts 22–27, and run separate application probes. The runner still has no Codex provider. |

Descriptions and skill names remain stable; this change does not retune automatic
selection across the whole pack. Claude-specific `model: sonnet` and tool grants
stay in the Claude plugin agent. No installed user skills or host configuration
are changed by editing this checkout.

## Behavioral validation

Baseline and updated probes use separate GPT-6 subagent contexts in the current
Codex environment. The baseline receives the old skill files; the updated pass
receives the current skills and scenario prompts without expected answers or
assertions. The parent assesses the returned actions against the assertions.
These are simulated decisions, not executions against the illustrative Go repos.

Baseline observations:

- The build-only request selected six gate commands. The agent noted:
  "There is no build-only branch."
- With required tools unavailable, the agent reported
  `gate: FAIL (verification incomplete)` and identified the missing aggregate
  status as an inference. It did **not** falsely report PASS.
- For the complete endpoint, it retained all explicit requirements despite the
  conflicting simplification paragraph. This was an instruction conflict found
  by inspection, not a demonstrated scope reduction in that probe.
- It recovered the installed script path using host guidance and rejected the
  assumed hook evidence. The literal checkout-relative command remained wrong.
- It reused checkpoint evidence because higher-priority host instructions
  overrode the skill's repeated-gate rule. The host already mitigated this case.

Updated results, assessed from the returned answers:

| Probe | Observed decision | Assessment |
|---|---|---|
| Quality 22 | `go build -o /dev/null ./cmd/server`; report build scope without claiming execution. | Pass; the output flag also avoids leaving a binary in the project. |
| Quality 23 | `gate: INCOMPLETE`, naming both unavailable tools. | Pass. |
| Quality 24 | Attach the same gate evidence to all three items; rerun affected checks after relevant changes or an explicit requirement. | Pass after clarifying the prompt to ask when reruns would be justified. |
| Quality 25 | Include batch processing, dry-run, pagination, and focused checks in the outline. | Pass. |
| Quality 26 | Establish a characterization baseline, then refactor without renewed approval. | Pass. |
| Quality 27 | From `/tmp/service`, invoke the absolute installed checker path with `.`; report missing siblings and do not assume hook output. | Pass. |
| Agent A | `go build ./...` only; success heading `build: PASS`. | Pass. |
| Agent B | Full gate reports `INCOMPLETE` and both unavailable tools. | Pass. |
| Delegation C | Permit an explicitly authorized, bounded independent review alongside other useful work. | Pass. |

The first answer to quality 24 correctly reused results but did not describe
future rerun conditions, which the original prompt had not asked for. Its
assertions were retained and the prompt clarified before rerunning that case.

Local checks passed: `go test -count=1 ./...` from `evals/`,
`agentskills-validate@1.0.1` for all five edited `SKILL.md` directories,
the CI relative-link check, and `git diff --check`. The Go tests include script
functional checks, eval schema, skill structure, and lint configuration validation.

## Opus verification through Claude CLI

On the same date, ran the six quality prompts and three agent/delegation probes
through Claude Code 2.1.261. The CLI's initialization and usage records confirm
the evaluated model was **`claude-opus-5`**, with `medium` effort. This was an
external Claude process, not a GPT subagent given an Opus role description.

Each prompt ran in a separate temporary directory, loading a snapshot of the
current checkout via `--plugin-dir`. Only `Skill`, `Read`, `Glob`, and `Grep`
were enabled; `--restricted`, disabled hooks, and an empty strict MCP config
kept the probes read-only. Stream output recorded actual skill calls and reads.
No project gates were executed. All nine initial processes completed successfully
with no permission denials; process success is not itself a quality score.

[Recorded answers, tool calls, and source hashes](evidence/2026-09-05-opus-5-probes.json)
retain the observations and parent assessments. Temporary paths are normalized
to `<probe-root>`; no login metadata is included.

The eight probes that loaded their instructions selected the expected core
actions: build-only scope, `INCOMPLETE` for missing evidence, checkpoint reuse,
complete requested behavior, characterization coverage, and authorized independent
review. Quality 27 initially answered without reading or invoking a skill, so
that run does not establish skill application. A separate run explicitly naming
the actual `go-code/SKILL.md` confirmed installed script resolution, project cwd,
missing-sibling handling, and no assumed hook results. It invoked the executable
checker directly rather than with `bash`, so the frozen assertion's literal
command spelling was not met; the path and target were correct.

There are still presentation issues: several answers are much longer than the
question needs; quality 23 adds unsolicited installation advice, and quality 26
asks for a real package path even though only a simulated next step was requested.
Do not interpret the matching core decisions as perfect compliance or a 9/9
first-pass benchmark. These observations do not change the Go rules in this run.

The key invocation, with the prompt and snapshot path supplied as variables:

```bash
claude -p "$probe_prompt" --model claude-opus-5 --effort medium \
  --plugin-dir "$probe_snapshot" --add-dir "$probe_snapshot" \
  --tools Skill,Read,Glob,Grep --allowed-tools Skill,Read,Glob,Grep \
  --restricted --strict-mcp-config --mcp-config '{"mcpServers":{}}' \
  --settings '{"disableAllHooks":true}' --no-session-persistence \
  --output-format stream-json --verbose
```

Quality 27's explicit-file rerun omitted `Skill` from the tools, matching that
scenario's capability boundary.

## Reproducing and interpreting the checks

Run the repository's structural and script tests with the supported Go and
golangci-lint versions on PATH:

```bash
cd evals
go test -count=1 ./...
```

The initial local baseline failed because PATH selected golangci-lint 2.12.2
built with Go 1.26. Selecting the already installed 2.13.1 built with Go 1.27
made it pass; no repository configuration change was needed for that failure.

For model probes, start a fresh session per variant with the checked-out skills.
Use the `prompt` fields of quality evals 22–27; keep their assertions out of the
evaluated model's context. Record model, host, loaded paths, prompt, actual
answer, and assertion outcomes. Also apply the agent to a build-only request
and a full gate with missing tools. Use scratch projects for execution tests.

The automated `evalrun` still measures Claude only. This review does not claim
an Opus-versus-GPT-6 benchmark, automatic trigger coverage on Codex, or a live
end-to-end test of all 78 trigger and 27 quality cases. Fresh subagents share
the host's higher-priority instructions, so these probes cannot isolate the
skill's causal contribution from the host's defaults.

## Troubleshooting follow-up

The [go-troubleshooting review](GO_TROUBLESHOOTING_REVIEW.md) records the later
ticket/data-flow expansion, new behavioral cases, actual Opus selection probes,
and remaining model limitations against baseline `e94ee92`.
