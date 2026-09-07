# Description compression, trigger recognition — GPT-5.6-Luna (medium) on Codex CLI

## Run

- Finished: 2026-09-07 20:36 EEST
- Runner: `codex` 0.153.4
- Model: `gpt-5.6-luna`, reasoning effort `medium`
- Cases: all 105 trigger evals from `evals/evals.json` (87 positive, 18 negative)
- Arms: `before` = descriptions at `8337ba3`; `after` = the working tree with
  18 of 24 descriptions shortened
- 1 repetition per case and arm, 210 sessions, 5 concurrent, 240-second timeout
- `before` skills SHA-256: `546f221e3e350d742cddf2a673a3f8ded1953d170a900f785182496ccf5cc176`
- `after` skills SHA-256: `9f6fba3ef9f4c9cbeec0b0dff5db73204e6617ce54bd971d99c241f1d7280d15`
- Raw report: [`2026-09-07-description-compression-codex-luna-medium.json`](2026-09-07-description-compression-codex-luna-medium.json)
- Raw report SHA-256: `f1952d7084f943e3a61655f9274a516328a2651278ef53c640aac9fbe730d08b`
- Harness: [`2026-09-07-description-compression-codex-luna-medium.py`](2026-09-07-description-compression-codex-luna-medium.py)

`evals/cmd/evalrun` drives the claude CLI only, so this run used a purpose-built
harness with the arm-isolation and transcript-parsing rules the codex runner in
`evals/cmd/abrun` already establishes: a private `HOME` and `CODEX_HOME` per arm,
and skill firing read from the *command* of a shell call, never from its output.

Codex is the sharper instrument for this particular question. It has no skill
tool: the skills reach the model as a catalog of names and descriptions in the
developer prompt, and a skill fires when the model decides to read its `SKILL.md`
through the shell. The description is therefore the only thing the model sees
before deciding, and it is the only thing this experiment changed.

Two failure modes of the earlier claude-side attempt are closed here. Jobs were
interleaved arm by arm, so any mid-run degradation would land on both arms at the
same point; and a session that exited non-zero or never completed a turn is
recorded as an error rather than as an empty skill set. **All 210 sessions
completed with a turn and a zero exit; there were no errors and no quota
failures.**

## What the model actually sees

`codex debug prompt-input` renders the developer prompt without spending a
request, so the catalog size is measured, not estimated.

| | before | after | delta |
|---|---:|---:|---:|
| Skills block | 13,930 chars | 10,338 chars | −3,592 (−25.8%) |
| Whole developer prompt | 23,128 chars | 19,536 chars | −3,592 (−15.5%) |

Both arms offer the same 24 go-* skills plus the 5 skills Codex ships itself, and
neither home exposed anything installed on the host.

## Recognition

| | before | after |
|---|---:|---:|
| Cases passed | 88/105 | 84/105 |
| Positive cases | 75/87 | 71/87 |
| Negative cases (no go-* skill may fire) | 13/18 | 13/18 |
| Expected skills read | 98/110 | 94/110 |
| Skill reads beyond expectation | 1.71/case | 1.33/case |
| Errors | 0 | 0 |

Fourteen cases flipped: 9 lost, 5 gained. McNemar exact p = 0.42 on case pass and
p = 0.45 on expected-skill reads. **Neither difference is distinguishable from
noise at one repetition per case**, and all 18 negative controls returned the
same verdict in both arms, case for case.

The flips carry their own noise estimate. Three of the fourteen turn on a skill
whose description was never touched — `go-code-refactor` lost *Modernize this Go
service* and gained *List the shortcuts we deliberately left in this codebase*,
and `go-troubleshooting` lost the Ukrainian root-cause case. Those three cannot
have been caused by the edit, which puts the run-to-run flip rate at roughly a
fifth of the observed movement before any effect of the shortening is counted.

Recall per expected skill moved on eight of the twenty-four:

| Skill | Cases expecting it | Read before | Read after |
|---|---:|---:|---:|
| `go-defensive` | 5 | 3 | 5 |
| `go-style-core` | 6 | 4 | 6 |
| `go-context` | 5 | 5 | 4 |
| `go-error-handling` | 6 | 6 | 5 |
| `go-troubleshooting` | 8 | 8 | 7 |
| `go-functions` | 4 | 4 | 2 |
| `go-interfaces` | 6 | 6 | 4 |
| `go-packages` | 3 | 1 | 0 |

`go-functions` and `go-interfaces` are the two worth rerunning: both lost two
cases, and in two of those four the model read `go-style-core` instead — the
generalist neither shortened description now fences off as sharply ("Should I use
named return values", "type switch versus if-else chain"). The other two went to
a plausible neighbour rather than nowhere: `go-logging` for the Printf-style
helper, `go-generics` for the two-constraint type parameter.
`go-packages` was already the weakest skill in the catalog at 1/3 before the
edit, so its 0/3 is a fixture problem more than a description problem.

## Cost

Codex reports tokens rather than dollars.

| Per session | before | after | delta |
|---|---:|---:|---:|
| Input tokens (mean) | 90,406 | 75,832 | −16.1% |
| of which cached | 74,225 | 60,782 | |
| Output tokens (mean) | 1,479 | 1,324 | −10.5% |
| Wall clock (mean) | 36.7 s | 33.1 s | −3.6 s |

The 16% is not all catalog. The `after` arm also *read fewer skills* — 2.23
against 2.65 per session — and each `SKILL.md` it opens is worth roughly 13,000
input tokens across the turns that follow. On the 50 cases where both arms read
the identical skill set, the paired median drop is 1,924 input tokens per
session, which is about the ~900-token catalog delta re-sent across a two-turn
session. That is the part attributable to the descriptions themselves; the rest
is the model opening less.

Across the 105-case suite: 9,492,609 → 7,962,404 input tokens (−16.1%).

## Interpretation

On this model and runner, shortening 18 of 24 descriptions by 36% of their text
**did not measurably damage recognition and did measurably shrink the prompt**.
The catalog costs ~900 fewer tokens every turn of every session, negative
controls are untouched, and the case-pass difference of four is well inside what
this harness produces from run-to-run variance alone.

What the run does not establish is equivalence. One repetition per case has no
power to rule out a real regression of a few points, and the direction of every
aggregate — 88 → 84, 98 → 94, six of the eight moved skills losing rather than
gaining — is consistently, if insignificantly, downward. The narrowing of extra
reads (1.71 → 1.33) is the same observation seen from the other side: the shorter
text pulls the model in less often, which is a saving where it was
over-triggering and a loss where it was not.

Before treating the compression as free, rerun `go-functions`, `go-interfaces`,
`go-context` and `go-error-handling` — the four skills that lost expected reads
and whose descriptions did change — at 5 repetitions per case on both arms. That
is 4 skills × their ~21 cases × 2 arms × 5 reps, which is affordable where the
full suite at that depth is not.
