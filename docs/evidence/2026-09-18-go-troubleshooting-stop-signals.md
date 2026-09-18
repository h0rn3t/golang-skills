# go-troubleshooting: Stop Signals and regression clause, quality evals before vs after

**Date:** 2026-09-18
**Runner:** `evalrun` (`claude -p --plugin-dir`, quality kind, judge = second plain session)
**Models:** evaluated arm `opus[1m]` (the CLI default here) and `claude-sonnet-5`; judge `opus[1m]`
**Plugin tree:** `fd10d80` (before) and `fd10d80` plus the go-troubleshooting patch (after), both as detached worktrees
**Evals:** `evals.json` filtered to the quality cases targeting `go-troubleshooting` (28–33 and the new 62); `evalrun` has no id filter, so each worktree got its own filtered file
**Repetitions:** Opus: once per case and arm. Sonnet: case 62 only, three copies per arm (ids 620–622)

Raw results and the exact patch: [JSON](2026-09-18-go-troubleshooting-stop-signals.json).

## What changed in the skill

A **Stop Signals** list (symptom patches that mean the mechanism is still
unknown: `recover()` around the panic, a longer `time.Sleep`, a mutex without
the racing pair, a retry or `GOMEMLIMIT` that quiets the ticket), a regression
clause with `git bisect run` and a `go version -m` diff of the two binaries,
a test-order bisection recipe, and a route to `go-code-refactor` when each fix
moves the failure. The flaky-test command lost the `-shuffle=on` that did
nothing next to a single-test `-run`. Quality case 62 is new: an on-call
pressure scenario where a retry "worked last week".

## Opus default, all seven cases

| Arm | Quality |
| --- | ---: |
| Before | 7 / 7 |
| After | 7 / 7 |

Saturated. Cases 28–33 stay green after the edit (regression guard); case 62
passes without the new text, so on this model the edit is not shown to change
anything.

## Sonnet 5, case 62 three times

| Arm | Sessions passed | Assertions failed |
| --- | ---: | --- |
| Before | 1 / 3 | "customer's claim is a report, not evidence" (1); "names reachability and timeout values in effect among the first checks" (1) |
| After | 3 / 3 | none |

Both arms passed the two assertions the Stop Signals text aims at in all six
sessions: the retry is never presented as the fix, and blind retry is flagged
as load amplification. The two before-arm failures are on evidence handling,
which the skill covered before this edit too. At n=3 this is suggestive, not
proof; the honest reading is that the edit did not regress anything and the
new case discriminates on Sonnet, not on Opus.

## Reproduction

```bash
# In a detached worktree, replace evals/evals.json with the troubleshooting
# quality cases (ids 28-33, 62), then:
cd evals
go run ./cmd/evalrun -set all -kind quality -j 2 -verbose -out run.json
go run ./cmd/evalrun -set all -kind quality -j 3 -model claude-sonnet-5 -verbose -out run-sonnet.json
```
