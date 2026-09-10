#!/usr/bin/env bash
set -uo pipefail

VERSION="1.2.0"
SCRIPT_NAME="$(basename "$0")"

usage() {
    cat <<EOF
$SCRIPT_NAME v$VERSION — Behavior-preservation harness for Go refactors

USAGE
    bash $SCRIPT_NAME [options] <mode> [path]

DESCRIPTION
    Captures build, vet, test, and race results, with informational lint findings.
    Matching records alone do not prove that observable behavior is unchanged.

    Modes:
      baseline      Record the state before any edit
      after         Record the state after a refactor step
      diff          Compare recorded check results, including failures and skips
      leaks         Run tests; leak verification stays incomplete without a profile
      loc-baseline  Record the starting production LOC, before the first edit
      loc-diff      Recount and compare against that record

    Results are written under .refactor-verify/ in the working directory.

    The check modes say nothing about size and the loc modes say nothing about
    behavior. Physical LOC counts every line of the non-test *.go files, blank
    lines and comments included; code LOC counts only the lines holding at
    least one Go token, so a deleted comment cannot pay for a line of code.

    Exits 0 if all checks pass (or the diff is empty, or neither LOC count
    grew), 1 if a check failed, the diff is non-empty, or a count grew,
    2 on usage or environment error.
    The leaks mode exits 3 when tests pass: this harness does not collect or
    inspect in-process leak profiles, so it cannot certify leak freedom. Its
    2 keeps the environment meaning (a toolchain older than Go 1.26).

OPTIONS
    -h, --help       Show this help message
    -v, --version    Show version
    --json           Output results as JSON
    --limit N        Max lines reported per section (0 = unlimited, default: 0)
    --out DIR        Result directory (default: .refactor-verify)

ARGUMENTS
    mode             baseline | after | diff | leaks | loc-baseline | loc-diff
    path             Package pattern (default: ./...); the loc modes count the
                     directory it names, recursively

EXAMPLES
    bash $SCRIPT_NAME baseline ./...
    bash $SCRIPT_NAME after ./internal/...
    bash $SCRIPT_NAME diff
    bash $SCRIPT_NAME --json after ./...
    bash $SCRIPT_NAME leaks ./...
    bash $SCRIPT_NAME loc-baseline ./internal/gateway
    bash $SCRIPT_NAME --json loc-diff ./internal/gateway
EOF
}

json_escape() {
    local s="$1"
    s="${s//\\/\\\\}"
    s="${s//\"/\\\"}"
    s="${s//$'\t'/\\t}"
    s="${s//$'\r'/}"
    s="${s//$'\n'/\\n}"
    printf '%s' "$s"
}

# Trim a blob to at most LIMIT lines. The result and the truncation flag stay
# in the current shell: a command substitution would discard TRUNCATED with the
# subshell that set it.
LIMITED_TEXT=""
TRUNCATED=false
apply_limit() {
    LIMITED_TEXT="$1"
    TRUNCATED=false
    if [[ "$LIMIT" -le 0 || -z "$LIMITED_TEXT" ]]; then
        return
    fi
    # grep -c, not `printf '%s\n' | wc -l`: that form appends a newline the text
    # may already end with, counting one line too many and reporting a complete
    # blob as truncated.
    local total
    total=$(printf '%s' "$LIMITED_TEXT" | grep -c '' | tr -d ' ')
    if [[ "$total" -le "$LIMIT" ]]; then
        return
    fi
    TRUNCATED=true
    LIMITED_TEXT="$(printf '%s\n' "$LIMITED_TEXT" | head -n "$LIMIT")"
}

JSON_OUTPUT=false
LIMIT=0
OUT_DIR=".refactor-verify"
MODE=""
TARGET=""

while [[ $# -gt 0 ]]; do
    case "$1" in
        -h|--help)    usage; exit 0 ;;
        -v|--version) echo "$SCRIPT_NAME v$VERSION"; exit 0 ;;
        --json)       JSON_OUTPUT=true; shift ;;
        --limit)
            if [[ $# -lt 2 ]]; then
                echo "error: --limit requires a number" >&2
                exit 2
            fi
            LIMIT="$2"; shift 2 ;;
        --out)
            if [[ $# -lt 2 ]]; then
                echo "error: --out requires a directory" >&2
                exit 2
            fi
            OUT_DIR="$2"; shift 2 ;;
        -*)           echo "error: unknown option: $1" >&2; usage >&2; exit 2 ;;
        *)
            if [[ -z "$MODE" ]]; then MODE="$1"; else TARGET="$1"; fi
            shift ;;
    esac
done

if ! [[ "$LIMIT" =~ ^[0-9]+$ ]]; then
    echo "error: --limit must be a non-negative integer, got: $LIMIT" >&2
    exit 2
fi

case "$MODE" in
    baseline|after|diff|leaks|loc-baseline|loc-diff) ;;
    "") echo "error: missing mode" >&2; usage >&2; exit 2 ;;
    *)  echo "error: unknown mode: $MODE" >&2; usage >&2; exit 2 ;;
esac

TARGET="${TARGET:-./...}"
GO_TEST_TIMEOUT="${GO_TEST_TIMEOUT:-5m}"

if ! command -v go &>/dev/null; then
    echo "error: go is not installed or not in PATH" >&2
    exit 2
fi

mkdir -p "$OUT_DIR" || { echo "error: cannot create $OUT_DIR" >&2; exit 2; }
OUT_ABS="$(cd "$OUT_DIR" && pwd)" || { echo "error: cannot resolve $OUT_DIR" >&2; exit 2; }

# ------------------------------------------------------------------ loc modes
# The counter is a Go program because the second number needs the Go scanner:
# a nonblank-line count reads a multiline string as code it is not, and a
# hand-rolled comment stripper reads a comment inside a raw string as a comment.
# It is built rather than `go run` so the exit status reaching the caller is the
# gate verdict and not "exit status 1" from the toolchain.
if [[ "$MODE" == "loc-baseline" || "$MODE" == "loc-diff" ]]; then
    LOC_DIR="${TARGET%%/...}"
    LOC_DIR="${LOC_DIR:-.}"
    if [[ ! -d "$LOC_DIR" ]]; then
        echo "error: not a directory: $LOC_DIR" >&2
        exit 2
    fi
    LOC_ABS="$(cd "$LOC_DIR" && pwd)" || { echo "error: cannot resolve $LOC_DIR" >&2; exit 2; }
    LOC_RECORD="$OUT_ABS/loc.baseline.json"
    if [[ "$MODE" == "loc-diff" && ! -f "$LOC_RECORD" ]]; then
        echo "error: no starting count recorded; run loc-baseline before the first edit" >&2
        exit 2
    fi
    LOC_WORK="$(mktemp -d)" || { echo "error: cannot create a temp directory" >&2; exit 2; }
    trap 'rm -rf "$LOC_WORK"' EXIT
    printf 'module refactorloc\n\ngo 1.21\n' >"$LOC_WORK/go.mod"
    cat >"$LOC_WORK/main.go" <<'GOEOF'
// Counts the production LOC of one directory two ways: every physical line of
// its non-test *.go files, and the subset of those lines holding at least one
// Go token. Both numbers gate a concision pass, because a physical count alone
// lets a deleted doc comment pay for added code.
package main

import (
	"crypto/sha256"
	"encoding/json"
	"flag"
	"fmt"
	"go/scanner"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// convention travels with every result: a line count is only checkable
// against the rule that produced it.
const convention = "physical: every line of the non-test *.go files, blank lines and comments included\n" +
	"code: only the lines holding at least one Go token"

type fileCount struct {
	Path     string `json:"path"`
	Physical int    `json:"physical"`
	Code     int    `json:"code"`
	Digest   string `json:"sha256"`
}

type counts struct {
	Root       string      `json:"root"`
	Physical   int         `json:"physical"`
	Code       int         `json:"code"`
	Files      int         `json:"files"`
	TestFiles  int         `json:"test_files"`
	ScanErrors int         `json:"scan_errors"`
	Detail     []fileCount `json:"detail"`
}

func main() {
	root := flag.String("root", "", "directory whose non-test *.go files are counted")
	asJSON := flag.Bool("json", false, "print JSON instead of text")
	save := flag.String("save", "", "write this run's counts to a file")
	baseline := flag.String("baseline", "", "compare this run against a saved file")
	flag.Parse()

	if *root == "" {
		fail("-root is required")
	}
	now, err := count(*root)
	if err != nil {
		fail(err.Error())
	}
	if *save != "" {
		data, err := json.MarshalIndent(now, "", "  ")
		if err != nil {
			fail(err.Error())
		}
		if err := os.WriteFile(*save, append(data, '\n'), 0o644); err != nil {
			fail(err.Error())
		}
	}
	if *baseline == "" {
		report(now, *save, *asJSON)
		return
	}
	before, err := load(*baseline)
	if err != nil {
		fail(err.Error())
	}
	os.Exit(compare(before, now, *asJSON))
}

func fail(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(2)
}

// count walks root and counts every non-test *.go file under it. Test files are
// counted but contribute no lines: adding a characterization test is part of a
// refactor, and moving code into a test is not a way to shrink the package.
func count(root string) (counts, error) {
	c := counts{Root: root}
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := d.Name()
		if d.IsDir() || !strings.HasSuffix(name, ".go") {
			return nil
		}
		if strings.HasSuffix(name, "_test.go") {
			c.TestFiles++
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			rel = path
		}
		code, scanErrors := codeCount(path, data)
		c.Files++
		c.Physical += lineCount(data)
		c.Code += code
		c.ScanErrors += scanErrors
		c.Detail = append(c.Detail, fileCount{
			Path:     filepath.ToSlash(rel),
			Physical: lineCount(data),
			Code:     code,
			Digest:   fmt.Sprintf("%x", sha256.Sum256(data)),
		})
		return nil
	})
	sort.Slice(c.Detail, func(i, j int) bool { return c.Detail[i].Path < c.Detail[j].Path })
	return c, err
}

// lineCount counts physical lines. A trailing line without a newline counts
// once; an empty file counts zero.
func lineCount(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := strings.Count(string(data), "\n")
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}

// codeCount counts the physical lines holding at least one Go token. A blank
// line and a comment-only line hold none; a line of code with a trailing
// comment holds several and counts once. A multiline literal counts on every
// line it spans, because those lines are code. The returned error count is the
// scanner's: a file it cannot tokenize has an untrustworthy code number, and
// the caller reports that rather than a quiet approximation.
func codeCount(path string, data []byte) (int, int) {
	file := token.NewFileSet().AddFile(path, -1, len(data))
	scanErrors := 0
	var s scanner.Scanner
	// Mode 0 drops comments, which is the point: what remains is code.
	s.Init(file, data, func(token.Position, string) { scanErrors++ }, 0)
	lines := make(map[int]bool)
	for {
		pos, tok, lit := s.Scan()
		if tok == token.EOF {
			break
		}
		if tok == token.SEMICOLON && lit == "\n" {
			// Inserted by the scanner at a line end. Its position is the
			// newline itself, so counting the line it spans would credit the
			// next line.
			continue
		}
		start := file.Line(pos)
		for line := start; line <= start+strings.Count(lit, "\n"); line++ {
			lines[line] = true
		}
	}
	return len(lines), scanErrors
}

func load(path string) (counts, error) {
	var c counts
	data, err := os.ReadFile(path)
	if err != nil {
		return c, err
	}
	if err := json.Unmarshal(data, &c); err != nil {
		return c, fmt.Errorf("parse %s: %w", path, err)
	}
	return c, nil
}

func report(c counts, saved string, asJSON bool) {
	if asJSON {
		printJSON(map[string]any{"mode": "loc-baseline", "counts": c, "record": saved, "convention": convention})
		return
	}
	fmt.Printf("--- loc baseline (%s) ---\n", c.Root)
	fmt.Printf("physical %d, code %d, production files %d, test files %d\n", c.Physical, c.Code, c.Files, c.TestFiles)
	fmt.Println(convention)
	warn(c)
	if saved != "" {
		fmt.Printf("record: %s\n", saved)
	}
}

// compare returns the exit status: 0 when neither count grew, 1 when either did.
func compare(before, now counts, asJSON bool) int {
	physical, code, files := now.Physical-before.Physical, now.Code-before.Code, now.Files-before.Files
	pass := physical <= 0 && code <= 0
	added, removed := changedFiles(before, now)
	if asJSON {
		printJSON(map[string]any{
			"mode": "loc-diff", "root": now.Root, "before": before, "after": now,
			"delta":   map[string]int{"physical": physical, "code": code, "files": files},
			"added":   added,
			"removed": removed, "gate_pass": pass, "convention": convention,
		})
	} else {
		fmt.Printf("--- loc diff (%s) ---\n", now.Root)
		fmt.Printf("physical %d -> %d (%+d)\n", before.Physical, now.Physical, physical)
		fmt.Printf("code     %d -> %d (%+d)\n", before.Code, now.Code, code)
		fmt.Printf("files    %d -> %d (%+d)\n", before.Files, now.Files, files)
		if len(added) > 0 {
			fmt.Printf("added: %s\n", strings.Join(added, ", "))
		}
		if len(removed) > 0 {
			fmt.Printf("removed: %s\n", strings.Join(removed, ", "))
		}
		fmt.Println(convention)
		warn(now)
		if pass {
			fmt.Println("gate: PASS, neither production LOC count grew")
		} else {
			fmt.Println("gate: FAIL, a production LOC count grew")
			fmt.Println("Both counts gate: deleting a comment does not pay for a line of code,")
			fmt.Println("and documentation that explains a decision stays.")
		}
	}
	if pass {
		return 0
	}
	return 1
}

// changedFiles names the production files the two runs do not share, so a count
// that moved because a file appeared or disappeared says so.
func changedFiles(before, now counts) (added, removed []string) {
	was := map[string]bool{}
	for _, f := range before.Detail {
		was[f.Path] = true
	}
	is := map[string]bool{}
	for _, f := range now.Detail {
		is[f.Path] = true
		if !was[f.Path] {
			added = append(added, f.Path)
		}
	}
	for _, f := range before.Detail {
		if !is[f.Path] {
			removed = append(removed, f.Path)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func warn(c counts) {
	if c.ScanErrors > 0 {
		fmt.Printf("warning: %d scanner error(s); the code count is unreliable until the package parses\n", c.ScanErrors)
	}
}

func printJSON(payload map[string]any) {
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		fail(err.Error())
	}
	fmt.Println(string(data))
}
GOEOF
    if ! (cd "$LOC_WORK" && GOWORK=off GOFLAGS= go build -o loc . >"$LOC_WORK/build.log" 2>&1); then
        echo "error: cannot build the LOC counter" >&2
        cat "$LOC_WORK/build.log" >&2
        exit 2
    fi
    LOC_ARGS=(-root "$LOC_ABS")
    $JSON_OUTPUT && LOC_ARGS+=(-json)
    if [[ "$MODE" == "loc-baseline" ]]; then
        LOC_ARGS+=(-save "$LOC_RECORD")
    else
        LOC_ARGS+=(-baseline "$LOC_RECORD")
    fi
    LOC_RC=0
    "$LOC_WORK/loc" "${LOC_ARGS[@]}" || LOC_RC=$?
    exit "$LOC_RC"
fi

# ------------------------------------------------------------------ diff mode
if [[ "$MODE" == "diff" ]]; then
    if [[ ! -f "$OUT_DIR/baseline.summary" || ! -f "$OUT_DIR/after.summary" ]]; then
        echo "error: need both a baseline and an after run first" >&2
        exit 2
    fi
    DIFF_OUT="$(diff -u "$OUT_DIR/baseline.summary" "$OUT_DIR/after.summary" 2>&1)" && DIFF_RC=0 || DIFF_RC=1
    apply_limit "$DIFF_OUT"
    DISPLAY="$LIMITED_TEXT"
    if $JSON_OUTPUT; then
        printf '{"mode":"diff","identical":%s,"diff":"%s","truncated":%s}\n' \
            "$( [[ $DIFF_RC -eq 0 ]] && echo true || echo false )" \
            "$(json_escape "$DISPLAY")" \
            "$($TRUNCATED && echo true || echo false)"
    elif [[ $DIFF_RC -eq 0 ]]; then
        echo "identical: recorded check results match, including any failures or skips"
    else
        echo "$DISPLAY"
        $TRUNCATED && echo "... (truncated at $LIMIT lines)"
        echo
        echo "Differences above may include new tests, resolved failures, or regressions."
        echo "Inspect the results; record equality alone does not establish behavior."
    fi
    exit "$DIFF_RC"
fi

# ----------------------------------------------------------------- leaks mode
# The goroutineleak pprof profile is GA since Go 1.27; Go 1.26 gates it behind
# GOEXPERIMENT=goroutineleakprofile, and setting that on 1.27+ fails the build.
if [[ "$MODE" == "leaks" ]]; then
    GO_MINOR="$(go version 2>/dev/null | sed -E 's/.*go1\.([0-9]+).*/\1/')"
    if ! [[ "$GO_MINOR" =~ ^[0-9]+$ ]] || [[ "$GO_MINOR" -lt 26 ]]; then
        echo "error: leaks mode needs Go 1.26+ (got go1.${GO_MINOR:-?})" >&2
        exit 2
    fi
    LEAK_LOG="$OUT_DIR/leaks.log"
    if [[ "$GO_MINOR" -ge 27 ]]; then
        go test -count=1 -timeout "$GO_TEST_TIMEOUT" "$TARGET" >"$LEAK_LOG" 2>&1 && LEAK_RC=0 || LEAK_RC=1
    else
        GOEXPERIMENT=goroutineleakprofile go test -count=1 -timeout "$GO_TEST_TIMEOUT" "$TARGET" \
            >"$LEAK_LOG" 2>&1 && LEAK_RC=0 || LEAK_RC=1
    fi
    apply_limit "$(cat "$LEAK_LOG")"
    LEAK_OUT="$LIMITED_TEXT"
    if $JSON_OUTPUT; then
        printf '{"mode":"leaks","go_minor":%s,"passed":false,"tests_passed":%s,"leaks_checked":false,"reason":"No in-process leak profile was collected or inspected","output":"%s","truncated":%s}\n' \
            "$GO_MINOR" \
            "$( [[ $LEAK_RC -eq 0 ]] && echo true || echo false )" \
            "$(json_escape "$LEAK_OUT")" \
            "$($TRUNCATED && echo true || echo false)"
    else
        echo "=== tests for leak investigation (go1.$GO_MINOR) ==="
        echo "$LEAK_OUT"
        $TRUNCATED && echo "... (truncated at $LIMIT lines)"
        if [[ "$GO_MINOR" -ge 27 ]]; then
            echo 'Collect and inspect the profile in-process via runtime/pprof.Lookup("goroutineleak").WriteTo(w, 1)'
            echo "or the /debug/pprof/goroutineleak endpoint."
        fi
        echo "INCOMPLETE: no leak profile was collected or inspected by this harness."
        echo "Use an instrumented test or the project's existing leak assertions."
        echo "--- full log: $LEAK_LOG ---"
    fi
    # 3, not 2: 2 is this script's usage/environment error everywhere else.
    [[ "$LEAK_RC" -ne 0 ]] && exit "$LEAK_RC"
    exit 3
fi

# --------------------------------------------------------- baseline / after
SUMMARY="$OUT_DIR/$MODE.summary"
LOG="$OUT_DIR/$MODE.log"
: >"$SUMMARY"
: >"$LOG"
FAILED=0

record() {
    echo "$1: $2" >>"$SUMMARY"
}

GO_VERSION="$(go version 2>/dev/null | awk '{print $3}')"
GO_DIRECTIVE="$(awk '/^go [0-9]/ {print $2; exit}' go.mod 2>/dev/null)"

# gofmt: report unformatted files by name so the summary diff is meaningful.
GOFMT_DIR="${TARGET%%/...}"
GOFMT_DIR="${GOFMT_DIR:-.}"
GOFMT_STATUS="skip"
if command -v gofmt &>/dev/null; then
    UNFORMATTED="$(gofmt -l "$GOFMT_DIR" 2>/dev/null | grep -v '^vendor/' || true)"
    if [[ -z "$UNFORMATTED" ]]; then
        GOFMT_STATUS="pass"
    else
        GOFMT_STATUS="fail"
        FAILED=1
        echo "=== gofmt ===" >>"$LOG"
        echo "$UNFORMATTED" >>"$LOG"
    fi
fi
record gofmt "$GOFMT_STATUS"

run_step() {
    local name="$1"; shift
    echo "=== $name ===" >>"$LOG"
    if "$@" >>"$LOG" 2>&1; then
        record "$name" pass
        return 0
    fi
    record "$name" fail
    FAILED=1
    return 1
}

run_step build go build "$TARGET"
run_step vet go vet "$TARGET"

# Per-test verdicts go into the summary: if a test disappears, starts being
# skipped, or starts failing, the baseline/after diff catches it. Timings are
# stripped because they vary per run.
echo "=== go test ===" >>"$LOG"
TEST_RAW="$OUT_DIR/$MODE.test.raw"
if go test -count=1 -timeout "$GO_TEST_TIMEOUT" -v "$TARGET" >"$TEST_RAW" 2>&1; then
    record test pass
else
    record test fail
    FAILED=1
fi
cat "$TEST_RAW" >>"$LOG"
grep -E '^[[:space:]]*--- (PASS|FAIL|SKIP):' "$TEST_RAW" 2>/dev/null \
    | sed -E 's/ \([0-9.]+s\)//' | sort >>"$SUMMARY" || true

run_step race go test -count=1 -race -timeout "$GO_TEST_TIMEOUT" "$TARGET"

# Pending modernizations, informational: excluded from the strict summary
# because a refactor is expected to reduce them, not hold them equal.
FIX_PENDING="n/a"
if go fix -diff "$TARGET" >"$OUT_DIR/$MODE.fix.diff" 2>/dev/null; then
    FIX_PENDING="$(wc -l <"$OUT_DIR/$MODE.fix.diff" | tr -d ' ')"
fi

LINT_FINDINGS="n/a"
LINT_STATUS="unavailable"
LINT_EXIT_CODE=null
LINT_LOG=""
if command -v golangci-lint &>/dev/null; then
    LINT_LOG="$OUT_DIR/$MODE.lint.raw"
    LINT_EXIT_CODE=0
    golangci-lint run "$TARGET" >"$LINT_LOG" 2>&1 || LINT_EXIT_CODE=$?
    cat "$LINT_LOG" >>"$LOG"
    LINT_COUNT="$(grep -cE '\.go:[0-9]+:[0-9]+:' "$LINT_LOG" 2>/dev/null || true)"
    if [[ "$LINT_EXIT_CODE" -eq 0 ]]; then
        LINT_STATUS="pass"
        LINT_FINDINGS="${LINT_COUNT:-0}"
    elif [[ "$LINT_EXIT_CODE" -eq 1 && "${LINT_COUNT:-0}" -gt 0 ]]; then
        LINT_STATUS="fail"
        LINT_FINDINGS="$LINT_COUNT"
    fi
fi

if $JSON_OUTPUT; then
    printf '{"mode":"%s","target":"%s","toolchain":"%s","go_directive":"%s","gofmt":"%s","summary_path":"%s","fix_pending_lines":"%s","lint_findings":"%s","lint_status":"%s","lint_exit_code":%s,"lint_log_path":"%s","passed":%s}\n' \
        "$MODE" "$(json_escape "$TARGET")" "$(json_escape "${GO_VERSION:-unknown}")" \
        "$(json_escape "${GO_DIRECTIVE:-none}")" "$GOFMT_STATUS" \
        "$(json_escape "$SUMMARY")" "$FIX_PENDING" "$LINT_FINDINGS" \
        "$LINT_STATUS" "$LINT_EXIT_CODE" "$(json_escape "$LINT_LOG")" \
        "$( [[ $FAILED -eq 0 ]] && echo true || echo false )"
else
    echo "--- $MODE ($TARGET) ---"
    echo "toolchain: ${GO_VERSION:-unknown} / go.mod directive: ${GO_DIRECTIVE:-none}"
    apply_limit "$(cat "$SUMMARY")"
    printf '%s\n' "$LIMITED_TEXT"
    $TRUNCATED && echo "... (truncated at $LIMIT lines)"
    echo "go fix pending: $FIX_PENDING diff lines (informational)"
    echo "lint: $LINT_STATUS; findings: $LINT_FINDINGS (informational)"
    if [[ -n "$LINT_LOG" ]]; then
        echo "lint exit: $LINT_EXIT_CODE; log: $LINT_LOG"
    else
        echo "lint unavailable: golangci-lint not found in PATH"
    fi
    echo "--- full log: $LOG ---"
    if [[ "$MODE" == "baseline" && "$FAILED" -eq 1 ]]; then
        echo
        echo "Baseline is red. Record known failures and continue independently" >&2
        echo "verifiable work; distinguish new failures before attributing them." >&2
    fi
fi

exit "$FAILED"
