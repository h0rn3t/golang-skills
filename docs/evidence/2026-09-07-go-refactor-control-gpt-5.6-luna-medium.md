# Go Refactor Skill Control — GPT-5.6-Luna (medium) on Codex CLI

The strongest result the refactor corpus has produced, and the first on which
more than one fixture separates the arms.

## Run

- Finished: 2026-09-07 16:40 EEST
- Runner: `codex` 0.153.4
- Model: `gpt-5.6-luna`, reasoning effort `medium` (the model's own default)
- Seed: `1`
- Fixtures: `dispatch`, `pricing`, `report`, `store`
- Arms: `no-skill`, `baseline`
- Repetitions: 5 per fixture and arm, 40 sessions total
- Baseline plugin SHA-256: `6443a658f74cd465b98f0065cdb7c117bcde52c64af1acc1a5ef7f3a67369d3c`
- Plugin source: working tree at `5dcb55d`
- Raw report: [`2026-09-07-go-refactor-control-gpt-5.6-luna-medium.json`](2026-09-07-go-refactor-control-gpt-5.6-luna-medium.json)
- Raw report SHA-256: `62a8fb56b1383f6243d94087984e753b35519af58ac0d715d9fc13e3e2c7b30f`

```bash
go run ./cmd/abrun -runner codex -model gpt-5.6-luna -effort medium \
  -arms no-skill,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-07-go-refactor-control-gpt-5.6-luna-medium.json
```

All 40 sessions completed, built, passed their hidden golden tests, changed the
fixture and stayed out of the repository checkout, so all 40 count as evidence.
`go-code-refactor` fired in all 20 baseline sessions and `go-style-core` in 19;
`go-testing` fired in 6, `go-code` in 4 and `go-code-review` in 1. This is the
first run on any runner where the owning skill reached every single session.

Two runner conditions matter for reading this file. Codex has no skill tool, so
a skill is scored from the shell command that opened its `SKILL.md`; Codex keeps
its own tool set, so these sessions have a shell and could run `go test` on their
own work. The plugin's PostToolUse hook and `go-verify` subagent do not apply.

## Results

Line values are mean delta with sample standard deviation over five runs.
Negative is less production code; the effect column is baseline minus no-skill,
so negative favors the skill.

| Fixture | No skill | Skill | Effect | 95% interval |
|---|---:|---:|---:|---|
| `dispatch` | −2.2 ± 4.3 | −8.0 ± 1.0 | **−5.8** | −10.4 to −1.2 |
| `pricing` | −33.8 ± 4.2 | −31.2 ± 2.8 | +2.6 | −2.6 to +7.8 |
| `report` | +18.8 ± 4.3 | +1.2 ± 4.9 | **−17.6** | −24.4 to −10.8 |
| `store` | −2.6 ± 3.8 | −6.2 ± 5.9 | −3.6 | −10.8 to +3.6 |
| All runs | −4.95 | −11.05 | **−6.10** | |

Two fixtures separate the arms here, where every previous run managed at most
one. On `report` the skill did not merely reduce growth, it removed it: the
control's five runs were +13, +16, +20, +21, +24 and the skilled arm's were −1,
−1, −1, −1, +10. Four of five skilled sessions returned a package *smaller*
than the one they were handed while the hidden golden test still passed. The
effect is −17.6 lines, the largest recorded, and its interval is the only one in
the corpus that clears zero by more than a rounding error.

`dispatch` is the surprise. It has gone the wrong way on every model measured
before this one — Opus 5 +2.0, MiniMax M3 +0.8, MiMo v2.5 Pro +3.0,
MAI-Code-1.1-Flash +5.8 — and here it is −5.8 with an interval that excludes
zero. The skilled arm is also far more consistent on it,
±1.0 against the control's ±4.3.

| Structural additions across 20 runs | No skill | Skill | Change |
|---|---:|---:|---:|
| Types | 1 | 1 | no difference |
| Interfaces | 0 | 0 | no difference |
| Functions | 31 | 9 | **−71.0%** |
| Pattern-name hits | 0 | 0 | no difference |

Function growth falls by more than on any other model; the previous best was
MiMo v2.5 Pro at −60.0%. Type growth has nowhere to move, because the unaided
model declared exactly one new type across 20 runs.

The skill produced a test file in 7 of 20 runs against 0 of 20 without it, which
lines up with `go-testing` firing in 6 sessions. `abrun` hides model-authored
tests before applying the independent golden test, so this is evidence of test
creation, not evidence that those tests assert the right behavior.

Codex reports token counts rather than dollars, so the `$/run` column is empty
for both arms.

## Interpretation

The corpus's central claim reproduces here more cleanly than anywhere else. The
trap is live — the unaided model grows `report` by 18.8 lines — the skill
removes all of it, correctness is untouched at 20/20 in both arms, and the
mechanism is unambiguous: 22 fewer helper functions across the arm, with no
change in types or interfaces. This model's failure mode is helper sprawl, the
same one MiniMax M3 has, and the skill suppresses it hardest.

`dispatch` moving for the first time is worth a second look rather than a
conclusion. Four models put it on the wrong side and one puts it clearly on the
right side; that is a fixture whose result depends on the model more than the
wording does, and `n=5` per cell with four fixtures tested at once makes these
descriptive intervals, not a preregistered result.
