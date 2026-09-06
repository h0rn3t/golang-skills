# Refactor Skill Review Fixes Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Correct unsafe refactoring guidance and make `abrun` reject or expose measurements that cannot support a code-quality claim.

**Architecture:** Keep `abrun` as one command, but extract small pure helpers for validation, job construction, subprocess result handling, recursive fixture analysis, and live status. Add an optional complete plugin-tree reference arm instead of pretending the no-skill control is a before/after comparison. Pin every reviewed documentation correction with repository regression tests.

**Tech Stack:** Go 1.27 standard library, Markdown skill references, existing `evals` test module.

---

### Task 1: Pin runner failure modes

**Files:**
- Create: `evals/cmd/abrun/main_test.go`
- Modify: `evals/cmd/abrun/main.go`

- [ ] **Step 1: Write failing subprocess and result-status tests**

Add a helper-process test that writes `partial output` and exits non-zero, then assert the command-output helper returns both bytes and an error. Add table tests for `resultStatus` so `Err`, build failure, and golden failure all produce `ERR`.

```go
func TestCommandOutputReturnsPartialOutputAndError(t *testing.T) {
    cmd := exec.Command(os.Args[0], "-test.run=TestHelperProcess", "--", "partial-error")
    cmd.Env = append(os.Environ(), "GO_WANT_HELPER_PROCESS=1")
    got, err := commandOutput(cmd)
    if err == nil {
        t.Fatal("commandOutput(partial-error) error = nil, want non-nil")
    }
    if string(got) != "partial output\n" {
        t.Fatalf("commandOutput(partial-error) output = %q, want %q", got, "partial output\n")
    }
}
```

- [ ] **Step 2: Run the focused test and verify RED**

Run: `go test ./cmd/abrun -run 'TestCommandOutput|TestResultStatus' -count=1`

Expected: compilation failure because `commandOutput` and `resultStatus` do not exist.

- [ ] **Step 3: Implement minimal subprocess and status helpers**

`commandOutput` must always wrap non-zero exits, even when stdout is non-empty. `claude` delegates to it and preserves the timeout-specific message. `printResult` obtains its prefix from `resultStatus`.

```go
func commandOutput(cmd *exec.Cmd) ([]byte, error) {
    var stderr strings.Builder
    cmd.Stderr = &stderr
    out, err := cmd.Output()
    if err != nil {
        return out, fmt.Errorf("%w: %s", err, strings.TrimSpace(stderr.String()))
    }
    return out, nil
}

func resultStatus(r result) string {
    if r.Err != "" || !r.Build || !r.Golden {
        return "ERR"
    }
    return "ok "
}
```

- [ ] **Step 4: Run the focused test and verify GREEN**

Run: `go test ./cmd/abrun -run 'TestCommandOutput|TestResultStatus' -count=1`

Expected: PASS.

### Task 2: Make fixture analysis recursive and fixture contracts mandatory

**Files:**
- Modify: `evals/cmd/abrun/main_test.go`
- Modify: `evals/cmd/abrun/main.go`

- [ ] **Step 1: Write failing recursive fixture tests**

Create nested production and test files under `t.TempDir()`. Assert `analyze` counts nested lines/types/interfaces/functions, including a final line without a trailing newline. Assert `hideTestFiles` renames nested `_test.go` files. Add a validation test where a task exists without its `_golden/task/*.go` companion.

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `go test ./cmd/abrun -run 'TestAnalyze|TestHideTestFiles|TestValidateFixtures' -count=1`

Expected: nested metrics and nested test hiding fail; fixture validation is undefined.

- [ ] **Step 3: Implement recursive walking and validation**

Replace both top-level `os.ReadDir` loops with `filepath.WalkDir`. Count every production `.go` file beneath the fixture root and every `_test.go` file separately. Add one for a non-newline-terminated final line. Validate every selected fixture and golden directory before locating `claude` or launching work.

```go
func lineCount(data []byte) int {
    if len(data) == 0 {
        return 0
    }
    lines := bytes.Count(data, []byte{'\n'})
    if data[len(data)-1] != '\n' {
        lines++
    }
    return lines
}
```

- [ ] **Step 4: Run the focused tests and verify GREEN**

Run: `go test ./cmd/abrun -run 'TestAnalyze|TestHideTestFiles|TestValidateFixtures' -count=1`

Expected: PASS.

### Task 3: Add a true reference arm and unbiased job order

**Files:**
- Modify: `evals/cmd/abrun/main_test.go`
- Modify: `evals/cmd/abrun/main.go`
- Modify: `evals/ab/README.md`

- [ ] **Step 1: Write failing arm and job-order tests**

Build two temporary plugin roots whose `go-code-refactor/SKILL.md` files have distinct markers. Assert `buildArms` materializes `reference` from `-reference-root` and `baseline` from the current root. Assert `buildJobs` is deterministic for one seed and differs from arm-major construction.

- [ ] **Step 2: Run the focused tests and verify RED**

Run: `go test ./cmd/abrun -run 'TestBuildArms|TestBuildJobs|TestValidateOptions' -count=1`

Expected: compilation failure because reference-root, seeded jobs, and option validation are absent.

- [ ] **Step 3: Implement reference-root, validation, and seeded shuffling**

Add `-reference-root` and `-seed`. A supplied reference root contributes a `reference` arm containing that entire plugin tree. Keep `baseline` as the current checkout and variants as current plus a Markdown splice. Validate positive `-n`, `-j`, and timeout values. Build all `(arm, task, repetition)` jobs, shuffle with `math/rand.NewSource(seed)`, and record the seed in JSON.

- [ ] **Step 4: Document evidence rules**

Document separate uses:

```bash
# Discover whether the complete current plugin beats no skill.
go run ./cmd/abrun -arms no-skill,baseline \
  -model claude-opus-4-1-20250805 -out control.json

# Compare a previous checkout with the current complete plugin tree.
go run ./cmd/abrun -reference-root ../golang-skills-before \
  -arms reference,baseline -model claude-opus-4-1-20250805 \
  -seed 1 -out before-after.json
```

Remove the uncommitted historical result tables. State that published claims require the raw JSON, exact model, seed, and both compared arms.

- [ ] **Step 5: Run arm tests and verify GREEN**

Run: `go test ./cmd/abrun -run 'TestBuildArms|TestBuildJobs|TestValidateOptions' -count=1`

Expected: PASS.

### Task 4: Pin and correct unsafe skill guidance

**Files:**
- Modify: `evals/eval_test.go`
- Modify: `skills/go-code-refactor/SKILL.md`
- Modify: `skills/go-code-refactor/references/CATALOG.md`
- Modify: `skills/go-code-refactor/references/MECHANICAL.md`
- Modify: `skills/go-code-refactor/references/SAFETY-NET.md`
- Modify: `skills/go-code-refactor/references/STRUCTURAL.md`

- [ ] **Step 1: Add failing documentation regression checks**

Extend `TestKnownReferenceRegressions` with positive required phrases and forbidden unsafe phrases. Pin these contracts:

- `deadcode` is evidence for investigation, never deletion proof;
- rename may introduce dynamic errors;
- Replace Temp with Query requires pure, stable, cheap expressions;
- mutable variables have no cross-package alias;
- interface methods and exported fields are not unconditionally additive;
- recursive `gofmt` never uses the invalid `./...` path;
- `-coverpkg` expands instrumentation rather than making no-test packages visible.

- [ ] **Step 2: Run documentation tests and verify RED**

Run: `go test . -run 'TestKnownReferenceRegressions|TestSkillArchitecture|TestRuleOwnershipMap' -count=1`

Expected: FAIL on the current unsafe or missing wording.

- [ ] **Step 3: Correct the main skill gates**

Make only genuinely purposeless churn a stop condition. Critical untested code requires characterization first; a minimal-change request receives the minimal safe change and reports a separate refactor. Describe gopls rename as compilation-aware rather than behavior-preserving, and keep it low-risk only for unexported, non-dynamic names with focused checks.

- [ ] **Step 4: Correct routed reference guidance**

Apply the exact safety qualifications from the approved design. Replace `gofmt ... -w ./...` with a valid single-package example plus guidance to enumerate repository files through a checked script for recursive work. Keep public API deletion and breaking changes outside a behavior-preserving refactor.

- [ ] **Step 5: Run documentation tests and verify GREEN**

Run: `go test . -run 'TestKnownReferenceRegressions|TestSkillArchitecture|TestRuleOwnershipMap' -count=1`

Expected: PASS.

### Task 5: Verify the complete change

**Files:**
- Verify all files changed by Tasks 1–4.

- [ ] **Step 1: Format and inspect the diff**

Run:

```bash
gofmt -w evals/cmd/abrun/main.go evals/cmd/abrun/main_test.go evals/eval_test.go
git diff --check
```

Expected: no formatting or whitespace errors.

- [ ] **Step 2: Run focused runner and golden tests**

Run:

```bash
cd evals
go test ./cmd/abrun -count=1
go test -race ./cmd/abrun -count=1
```

Copy each golden test into a temporary module beside its fixture and run `go test ./...` there.

Expected: PASS for runner tests and all four golden packages.

- [ ] **Step 3: Run repository checks**

Run:

```bash
cd evals
go vet ./...
go test ./... -count=1
go test -race ./... -count=1
```

Expected: PASS, except a locally stale `golangci-lint` binary may keep the known `SetupLintDryRun` case unavailable; report the exact output rather than treating it as clean.

- [ ] **Step 4: Review final staged and unstaged state**

Run: `git status --short && git diff --stat && git diff --cached --stat`

Expected: only scoped implementation changes plus the user's pre-existing staged change remain; no paid live A/B report is claimed.
