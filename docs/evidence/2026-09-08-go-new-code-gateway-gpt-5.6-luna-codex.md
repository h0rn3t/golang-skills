# `gateway` Helper-Count Follow-Up — GPT-5.6-Luna (medium) on Codex CLI

This run was commissioned to settle one question left open by the
[corpus run earlier the same day](2026-09-08-go-new-code-implement-gpt-5.6-luna-codex.md):
whether the skill tree makes this model add helper functions on `gateway` that
it does not write unaided, and whether editing `go-code`'s new-code section
moves that. It answers a different question than expected. **The `gateway`
helper count is too unstable for an n=5 arm pair to measure, and re-measuring
one byte-identical arm overturns the earlier run's central claim.**

## Run

- Finished: 2026-09-08 16:53 EEST (started 16:44, 8.3 minutes wall clock)
- Runner: `codex` 0.153.4
- Model: `gpt-5.6-luna`, reasoning effort `medium`
- Seed: `1`; corpus `implement`, `-tasks gateway`
- Arms: `no-skill`, `reference`, `baseline`; 5 repetitions each, 15 sessions
- `reference`: `git archive d515147` — the tree with the `## Writing New Code`
  section as committed. SHA-256
  `4223809986c69102d4de921698352b6811cc3c1b01eddf978d2c9a77611e88bb`, byte-identical
  to the arm measured as `baseline` in the corpus run.
- `baseline`: working tree at `d515147` plus uncommitted edits — the rewritten
  helper-extraction and interface paragraphs plus a new closing-inspection
  paragraph in `go-code`, edits to `go-functions`, `go-interfaces`,
  `go-packages`, and the `go-code-refactor` PLAYBOOK split into
  `POLICY-TABLES.md`. SHA-256
  `e044fd6bb36bdfdf39dba916b750afdac2577c7141d3c793dd53fd5f83fcfb70`
- Raw report: [`2026-09-08-go-new-code-gateway-gpt-5.6-luna-codex.json`](2026-09-08-go-new-code-gateway-gpt-5.6-luna-codex.json)
- Raw report SHA-256: `2bd65dcde6e7b8a0ad3f2dda4797db4ecaed860193cc3014e85f64664c01dc6c`

```bash
go run ./cmd/abrun -corpus implement -tasks gateway -runner codex \
  -model gpt-5.6-luna -effort medium \
  -reference-root <git archive d515147 checkout> \
  -arms no-skill,reference,baseline -n 5 -j 4 -seed 1 \
  -out ../docs/evidence/2026-09-08-go-new-code-gateway-gpt-5.6-luna-codex.json
```

15/15 valid: every session built, passed the hidden golden test, changed the
fixture and stayed out of the repository checkout.

## This run in isolation

| Arm | Lines | Helpers | Types | Branches | Golden |
|---|---:|---:|---:|---:|---:|
| No skill | +97.0 ± 8.3 | 2.00 ± 1.41 | 0.00 | 13.20 | 5/5 |
| `d515147` | +103.2 ± 15.0 | 4.40 ± 2.06 | 0.60 | 12.80 | 5/5 |
| Working tree | +91.8 ± 16.0 | 2.40 ± 1.85 | 0.20 | 13.20 | 5/5 |

| Comparison | Lines | Helpers |
|---|---|---|
| `d515147` → working tree | −11.4 (−34.0 to +11.2, p = 0.30) | −2.00 (−4.9 to +0.9, p = 0.23) |
| No skill → working tree | −5.2 (−23.8 to +13.4, p = 0.57) | +0.40 (−2.0 to +2.8, p = 0.87) |
| No skill → `d515147` | +6.2 (−11.4 to +23.8, p = 0.54) | +2.40 (−0.2 to +5.0, p = 0.13) |

Read alone, this run says the uncommitted edit is the best of the three: fewest
lines, fewer helpers than the committed tree, and level with the unaided model
on helpers. Nothing separates at n = 5, and the previous run's `no-skill`
→ skilled helper finding does not reproduce as significant here.

## The reason that reading is not enough

The `reference` arm is byte-identical to the corpus run's `baseline` arm — same
digest, same seed, same model, same prompt, roughly two hours apart. It did not
reproduce:

| Same tree `d515147`, `gateway` | Sessions | Mean |
|---|---|---:|
| Corpus run, 15:40 | 104, 107, 76, 88, 90 | 93.0 lines |
| This run, 16:44 | 125, 112, 106, 87, 86 | 103.2 lines |
| Corpus run helpers | 0, 3, 0, 2, 4 | 1.80 |
| This run helpers | 7, 6, 2, 5, 2 | 4.40 |

The two samples are statistically compatible (lines p = 0.33, helpers p = 0.12),
so this is one wide distribution, not a changed model. Pooled at n = 10 the tree
writes **98.1 ± 14.2 lines** and **3.10 ± 2.26 helpers**, with helpers ranging
0 to 7 per session. The `no-skill` arm drifted the same way, 0.80 → 2.00 helpers.

## What the pooled data says

| `gateway` tree | n | Lines | Helpers/session |
|---|---:|---:|---:|
| No skill | 10 | 95.4 ± 7.3 | **1.40 ± 1.43** |
| `0aca2fe`, before `## Writing New Code` | 5 | 93.2 | **3.20** |
| `d515147`, with it | 10 | 98.1 ± 14.2 | **3.10 ± 2.26** |
| Working tree, today's edit | 5 | 91.8 | **2.40** |

Two corrections to the corpus run's conclusions follow from this table, and both
matter more than anything this run measured on its own.

**The `## Writing New Code` section did not reduce helper count.** The corpus
run put that tree at 1.80 helpers per session against the pre-edit tree's 3.20
and called it directional. With the second sample pooled in, the same tree sits
at 3.10 — indistinguishable from the 3.20 it was supposed to improve on. The
1.80 was the low half of a distribution that reaches 7.

**The pre-edit tree's helper cost survives as a direction, not a result.** That
finding was `no-skill` 0.80 → `0aca2fe` 3.20, interval +0.7 to +4.1, p = 0.040 —
the only interval excluding zero in that run. Re-estimated against the pooled
control it is +1.80, interval +0.3 to +3.3, p = 0.066. The gap is still there and
still points the same way — every skilled tree measured on `gateway` writes more
helpers than the unaided model, 2.40 to 3.20 against 1.40 — but the p < 0.05 in
the earlier file came from a control arm that happened to draw low, and should
not be cited as an established effect.

Today's edit at 2.40 is the lowest skilled number on the board and the only one
whose lines fall below the unaided model. At n = 5, against a per-session range
of 0 to 7, that is a hint and nothing more.

## Practical conclusion

The uncommitted edit is safe to keep: 15/15 golden, no build failures, no
interfaces, no unrequested exports, and every point estimate at least as good as
the committed tree's. It is not measured, and on this fixture at this sample size
it cannot be.

For anything stronger, `gateway` needs roughly n = 20 per arm rather than 5 — the
per-session helper spread of ±2.26 puts the detectable difference at about two
helpers, which is the entire size of the effect being chased. That is 40+ codex
sessions per comparison. The cheaper alternative is to stop measuring helper
counts on this fixture and this model altogether: the control arm passes every
trap, so the corpus can only score scaffold, and the scaffold metric here is
noisier than the differences being tested.
