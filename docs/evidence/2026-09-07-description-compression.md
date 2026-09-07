# Description compression experiment

## Working-tree result

Shortened 18 of 24 skill descriptions. Total description length decreased
from 9,978 to 6,388 Unicode characters (3,590 fewer; 36.0%). This measures
description text only, not the host's complete skill catalog or billed tokens.
Skill bodies, references, runtime permissions, and trigger expectations were
not changed. Existing exact-description test goldens were updated.

The following descriptions remain unchanged: `go-code-refactor`, `go-testing`,
`go-logging`, `go-http`, `go-security`, and `go-troubleshooting`.

**Status: provisional. Behavioral equivalence and dollar savings are not
established.** The final 18-description subset has not had a live trigger run.

## Live experiment

Claude Code 2.1.261, default configured model, existing `evalrun`, all 105
trigger cases, four concurrent cases per arm, 90-second timeout per case.
The original tree and a separate temporary candidate checkout ran with
overlapping execution. The candidate shortened all 24 descriptions, to
4,809 characters; it is different from the conservative final subset.

Commands, run from each checkout's `evals` directory:

```bash
go run ./cmd/evalrun -kind trigger -set all -j 4 -timeout 90s \
  -out /tmp/golang-skills-trigger-before.json
go run ./cmd/evalrun -kind trigger -set all -j 4 -timeout 90s \
  -out /tmp/golang-skills-trigger-after.json
```

Raw reports:

- [Original descriptions](2026-09-07-description-compression-trigger-before.json)
- [All-24 candidate](2026-09-07-description-compression-trigger-candidate.json)

The runner printed 49/105 before and 46/105 for the candidate, with 10 and 8
explicit errors respectively. **These are not valid comparative accuracy
scores.** Both runs encountered timeouts and then a session quota failure.
A direct diagnostic request returned HTTP 429, `terminal_reason: api_error`,
zero input/output tokens, and `You've hit your session limit`.
The trigger parser can record this kind of response as an empty skill set
without an error, making positive cases fail and negative cases pass without
a valid model decision. The exact affected cases cannot be recovered from
the reports because they do not retain the raw CLI transcripts.

## Conservative retention

The candidate introduced these before-pass / candidate-fail observations:

| Query | Expected skills retained at their original description |
|---|---|
| Write a test for a Go shipping-cost function | `go-testing` |
| The same Go error is logged at three layers | `go-logging` |
| This handler is a mess; make it readable | `go-code-refactor` |
| Download a file by a client-supplied name | `go-http`, `go-security` |
| Store passwords and verify API keys | `go-security` |
| Go service RSS grows despite a flat heap profile | `go-troubleshooting` |

The first three appeared before the burst of empty results; the remaining
three appeared around quota exhaustion. None could be repeated after the
quota error. All six descriptions were retained conservatively; this is not
evidence that shortening any particular one caused the observed failure.
Retaining them also does not prove that routing interactions in the final
subset are equivalent.

## Local verification

- The all-24 candidate passed the skill creator's frontmatter validator.
- Original structure and description-golden tests passed before editing.
- The final subset passed frontmatter validation for all 24 skills and
  `TestStructure`, `TestFrontmatterDescriptionsInvariant`,
  `TestSkillArchitecture`, `TestCrossRefs`, and `TestEvalsJSONSchema`.
- A direct comparison confirmed that only the description line changed in
  each edited skill. `git diff --check` passed.
- Full candidate tests reached an unrelated environment failure in
  `TestScriptFunctional/SetupLintDryRun`: installed `golangci-lint` was built
  with Go 1.26, below the repository's Go 1.27 target. The identical failure
  was reproduced on the original tree before applying description changes.
  The full final-subset run failed at the same check.

## Codex rerun of the final subset

The comparison the quota failure prevented was completed on a different host:
[GPT-5.6-Luna (medium) on the Codex CLI](2026-09-07-description-compression-codex-luna-medium.md),
all 105 trigger cases, both arms, 210 sessions, no errors. Codex renders the
descriptions into the developer prompt and a skill fires when the model reads its
`SKILL.md`, so the description is the only input to the decision.

The skills block fell 13,930 → 10,338 characters, cases passed 88/105 → 84/105,
expected skills read 98/110 → 94/110, and all 18 negative controls held case for
case (McNemar exact p = 0.42). Three of the fourteen flips turn on a description
that was never edited, so part of that movement is run-to-run noise. **Prompt
size is established; behavioral equivalence still is not** — one repetition per
case cannot rule out a small regression, and `go-functions`, `go-interfaces`,
`go-context` and `go-error-handling` each lost an expected read.

That run says nothing about the claude host, whose skill tool sees the same
descriptions through a different mechanism. To settle this one there,
rerun the final subset against original descriptions with a working session
quota. Compare expected-skill recall per case, negative cases, and additional
skill loads; the existing positive-case pass check does not reject additional
loads. Use the intended host/model for a subsequent cost comparison. Do not use
these quota-contaminated reports as a baseline for further tuning.
