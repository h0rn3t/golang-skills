package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The edit hook's check runner: every check leaves an immutable receipt, and
// the verifier credits a receipt only while its inputs are unchanged. It
// records evidence; the gate policy lives in the scripts.

var checkNames = []string{"gofmt", "vet", "fix", "test", "lint"}

const inputBudget = 128 << 20

// receipt is one check's record, <run-dir>/<check>.json.
type receipt struct {
	Version        int               `json:"version"`
	Check          string            `json:"check"`
	Status         string            `json:"status"`
	ExitCode       *int              `json:"exit_code"`
	Command        []string          `json:"command"`
	Cwd            string            `json:"cwd"`
	File           string            `json:"file"`
	Scope          scope             `json:"scope"`
	Config         string            `json:"config"`
	Filter         string            `json:"filter"`
	ToolVersion    string            `json:"tool_version"`
	Toolchain      map[string]string `json:"toolchain"`
	Started        string            `json:"started"`
	Ended          string            `json:"ended"`
	InputsBefore   string            `json:"inputs_before"`
	InputsAfter    string            `json:"inputs_after"`
	Reusable       bool              `json:"reusable"`
	Reason         string            `json:"reason"`
	SnapshotErrors []string          `json:"snapshot_errors"`
	Diagnostic     string            `json:"diagnostic"`
}

type scope struct {
	Kind   string `json:"kind"`
	Target string `json:"target"`
}

// generation names the receipt directory of the last Go edit; attempted adds
// the checks the verifier credited for it.
type generation struct {
	RunDir  string   `json:"run_dir"`
	Package string   `json:"package,omitempty"`
	State   string   `json:"state,omitempty"`
	Credits []string `json:"credits,omitempty"`
}

func receiptMain(args []string) int {
	code, err := receiptCommand(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "edit-hook receipt unavailable: "+err.Error())
		return 2
	}
	return code
}

func receiptCommand(args []string) (int, error) {
	if len(args) == 0 {
		return 0, errors.New("unknown receipt operation")
	}
	op, args := args[0], args[1:]
	switch {
	case op == "run" && len(args) >= 7:
		return run(args)
	case op == "begin" && len(args) == 4:
		return 0, begin(args[0], args[1], args[2], args[3])
	case op == "verify-gate" && len(args) == 2:
		return verifyGate(args[0], args[1])
	case op == "current-credits" && len(args) >= 1:
		fmt.Println(encode(nonNil(currentCredits(args[0], args[1:]))))
		return 0, nil
	case op == "verify" && len(args) >= 1:
		return verify(args[0], args[1:])
	case (op == "report" || op == "context") && len(args) == 1:
		return 0, report(args[0], op == "context")
	case op == "module" && len(args) == 1:
		code, module := capture([]string{"go", "env", "GOMOD"}, args[0], false)
		if code == 0 && module != os.DevNull && isFile(module) {
			return 0, nil
		}
		return 1, nil
	case op == "config" && len(args) == 1:
		code, output := capture([]string{"golangci-lint", "config", "path"}, args[0], true)
		fmt.Println(output)
		return code, nil
	}
	return 0, errors.New("unknown receipt operation")
}

func digest(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// remaining is the time left before the edit hook's global deadline.
func remaining() time.Duration {
	deadline, err := strconv.ParseFloat(os.Getenv("GOLANG_SKILLS_HOOK_DEADLINE"), 64)
	if err != nil {
		return 135 * time.Second
	}
	return max(time.Millisecond, time.Until(time.Unix(0, int64(deadline*1e9))))
}

// runCapped runs argv in dir with its output combined, killing its process
// group after limit; timedOut reports that kill.
func runCapped(argv []string, dir string, logical bool, limit time.Duration) (code int, output []byte, timedOut bool, err error) {
	ctx, cancel := context.WithTimeout(context.Background(), limit)
	defer cancel()
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...) //nolint:gosec // G204: the checks go-vet-on-edit.sh names, run as argv
	cmd.Dir = dir
	pwd := dir
	if !logical {
		pwd = resolve(dir)
	}
	cmd.Env = append(os.Environ(), "PWD="+pwd)
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &out
	ownGroup(cmd)
	cmd.Cancel = func() error {
		killGroup(cmd)
		return nil
	}
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		return 127, nil, false, err
	}
	_ = cmd.Wait() // the exit status is read from ProcessState
	if ctx.Err() != nil {
		return 124, out.Bytes(), true, nil
	}
	return exitStatus(cmd.ProcessState), out.Bytes(), false, nil
}

// capture runs a metadata probe: at most 8 seconds, trimmed output.
func capture(argv []string, dir string, logical bool) (int, string) {
	code, out, timedOut, err := runCapped(argv, dir, logical, min(8*time.Second, remaining()))
	switch {
	case err != nil:
		return 127, err.Error()
	case timedOut:
		return 124, "metadata probe deadline exceeded"
	}
	return code, strings.TrimSpace(string(out))
}

var goBuildPrefix = regexp.MustCompile(`-(ffile|fdebug)-prefix-map=.*?go-build\d+=/tmp/go-build`)

// snapshot digests every input a check reads: the module trees (local
// replacements and workspace members included), the tools, the build
// environment, the config, and for new-since-HEAD lint the revision. Any
// input it cannot read lands in problems, which prevents reuse.
func snapshot(file, check, config, filterKind, receiptDir string) (sum string, problems []string, version string, toolchain map[string]string) {
	cwd := filepath.Dir(resolve(file))
	tool := "go"
	switch check {
	case "gofmt":
		tool = "gofmt"
	case "lint":
		tool = "golangci-lint"
	}
	executable, lookErr := exec.LookPath(tool)
	versionArgs := []string{tool, "version"}
	if tool == "gofmt" {
		versionArgs = []string{"go", "version"}
	}
	in := &inputs{receiptDir: resolve(receiptDir)}
	rc, version := capture(versionArgs, cwd, false)
	if rc != 0 || lookErr != nil {
		in.problem("tool version unavailable")
	}
	env := map[string]string{}
	if rc, raw := capture([]string{"go", "env", "-json"}, cwd, false); rc != 0 || json.Unmarshal([]byte(raw), &env) != nil {
		in.problem("Go build environment unavailable")
	}
	in.environment(env)

	for _, root := range in.moduleRoots(env, check, cwd) {
		if fi, err := os.Stat(root); err != nil || !fi.IsDir() {
			in.problem("input root missing: " + root)
			continue
		}
		if err := in.walk(root); err != nil {
			in.problem(err.Error())
			break
		}
	}
	work := env["GOWORK"]
	extras := []string{executable, env["GOENV"], config}
	if work != "" && work != "off" {
		extras = append(extras, work, work+".sum")
	}
	// A compiler change with the same version is not the same toolchain.
	goExec, _ := exec.LookPath("go")
	extras = append(extras, goExec, filepath.Join(env["GOROOT"], "pkg", "tool", env["GOHOSTOS"]+"_"+env["GOHOSTARCH"], "compile"))
	for _, p := range extras {
		if p == "" || p == "off" {
			continue
		}
		if err := in.add(resolve(p), p == env["GOENV"] || strings.HasSuffix(p, ".sum")); err != nil {
			in.problem(err.Error())
		}
	}
	if filterKind == "new-since-HEAD" {
		rc, head := capture([]string{"git", "rev-parse", "HEAD"}, cwd, false)
		if rc != 0 {
			in.problem("lint revision unavailable")
		}
		in.entry("HEAD", head)
	}
	toolchain = map[string]string{}
	for _, k := range []string{"GOVERSION", "GOOS", "GOARCH", "CGO_ENABLED", "GOEXPERIMENT", "GOTOOLCHAIN", "GOROOT"} {
		toolchain[k] = env[k]
	}
	sort.Strings(in.entries)
	all, _ := json.Marshal(in.entries)
	return digest(all), in.problems, version, toolchain
}

var errBudget = errors.New("input budget exceeded")

// inputs accumulates the digest entries of one snapshot and the problems
// that prevent its reuse.
type inputs struct {
	receiptDir string
	entries    []string
	problems   []string
	size       int64
	count      int
}

func (in *inputs) entry(fields ...any) {
	data, _ := json.Marshal(fields)
	in.entries = append(in.entries, string(data))
}

func (in *inputs) problem(p string) {
	in.problems = append(in.problems, p)
}

// environment digests the go env and the build variables. Settings that may
// carry credentials are hashed, never printed or stored.
func (in *inputs) environment(env map[string]string) {
	stable := maps.Clone(env)
	if stable == nil {
		stable = map[string]string{}
	}
	stable["GOGCCFLAGS"] = goBuildPrefix.ReplaceAllString(env["GOGCCFLAGS"], "-${1}-prefix-map=<go-build>=/tmp/go-build")
	stableJSON, _ := json.Marshal(stable)
	in.entry("go-env", digest(stableJSON))
	// Only what the go command and golangci-lint read: the hook and the Bash
	// tool differ in terminal, proxy, and session variables, and hashing
	// those made every receipt stale for the shell that verifies it.
	buildEnv := map[string]string{}
	for _, kv := range os.Environ() {
		k, v, _ := strings.Cut(kv, "=")
		if (strings.HasPrefix(k, "GO") || strings.HasPrefix(k, "CGO_") || strings.HasPrefix(k, "PKG_CONFIG") || k == "CC" || k == "CXX" || k == "AR") &&
			!strings.HasPrefix(k, "GOLANG_SKILLS_") {
			buildEnv[k] = v
		}
	}
	buildJSON, _ := json.Marshal(buildEnv)
	in.entry("build-env", digest(buildJSON))
}

type replacement struct {
	New struct{ Path, Version string }
}

// moduleRoots are the module directories the check reads, sorted: the
// package's module, the workspace members, and every local replacement. Only
// local replacements are followed; a receipt never downloads a dependency.
func (in *inputs) moduleRoots(env map[string]string, check, cwd string) []string {
	roots := map[string]bool{}
	var pending []string
	if gomod := env["GOMOD"]; gomod != "" && gomod != os.DevNull && isFile(gomod) {
		pending = append(pending, filepath.Dir(resolve(gomod)))
	} else if check == "gofmt" {
		roots[cwd] = true
	} else {
		in.problem("no module")
	}
	if work := env["GOWORK"]; work != "" && work != "off" {
		var ws struct {
			Use     []struct{ DiskPath string }
			Replace []replacement
		}
		if rc, data := capture([]string{"go", "work", "edit", "-json", work}, cwd, false); rc != 0 || json.Unmarshal([]byte(data), &ws) != nil {
			in.problem("workspace inputs unavailable")
		}
		for _, u := range ws.Use {
			pending = append(pending, resolve(under(filepath.Dir(work), u.DiskPath)))
		}
		pending = append(pending, localReplacements(filepath.Dir(work), ws.Replace)...)
	}
	for len(pending) > 0 {
		root := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if roots[root] {
			continue
		}
		roots[root] = true
		var mod struct{ Replace []replacement }
		if rc, data := capture([]string{"go", "mod", "edit", "-json", filepath.Join(root, "go.mod")}, cwd, false); rc != 0 || json.Unmarshal([]byte(data), &mod) != nil {
			in.problem("local module inputs unavailable")
			continue
		}
		pending = append(pending, localReplacements(root, mod.Replace)...)
	}
	return sortedKeys(roots)
}

// localReplacements are the directories of the replacements without a
// version, relative to base.
func localReplacements(base string, replace []replacement) []string {
	var dirs []string
	for _, r := range replace {
		if r.New.Version == "" {
			dirs = append(dirs, resolve(under(base, r.New.Path)))
		}
	}
	return dirs
}

// add digests one file: content, modification and status-change times, and
// inode, which detect a rewrite that restores the original bytes. An optional
// file may be absent.
func (in *inputs) add(p string, optional bool) error {
	if _, err := os.Lstat(p); err != nil && optional {
		in.entry(p, "absent")
		return nil
	}
	f, err := openInput(p)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }() // read only: a close error loses nothing
	info, err := f.Stat()
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() {
		return errors.New("nonregular input: " + p)
	}
	in.size += info.Size()
	in.count++
	if in.size > inputBudget || in.count > 20000 {
		return errBudget
	}
	content, err := io.ReadAll(f)
	if err != nil {
		return err
	}
	ino, ctime := changeMeta(info)
	link, err := os.Lstat(p)
	if err != nil {
		return err
	}
	_, linkCtime := changeMeta(link)
	target, _ := os.Readlink(p)
	in.entry(p, digest(content), info.ModTime().UnixNano(), ctime, ino, linkCtime, target)
	return nil
}

// walk adds every file under root except .git and the receipt directory. A
// directory symlink is an input the snapshot cannot follow. It stops with an
// error at the input budget or the hook deadline.
func (in *inputs) walk(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		in.problem(err.Error())
		return nil
	}
	if remaining() < time.Second {
		return errors.New("hook input snapshot deadline exceeded")
	}
	var dirs []string
	for _, e := range entries {
		p := filepath.Join(root, e.Name())
		if e.Name() == ".git" {
			continue
		}
		isDir := e.IsDir()
		if e.Type()&fs.ModeSymlink != 0 {
			if fi, err := os.Stat(p); err == nil && fi.IsDir() {
				if resolve(p) != in.receiptDir {
					in.problem("directory symlink inputs unsupported: " + p)
				}
				continue
			}
		}
		if isDir {
			if resolve(p) != in.receiptDir {
				dirs = append(dirs, p)
			}
			continue
		}
		if err := in.add(p, false); err != nil {
			if errors.Is(err, errBudget) {
				return err
			}
			in.problem(err.Error())
		}
	}
	for _, d := range dirs {
		if err := in.walk(d); err != nil {
			return err
		}
	}
	return nil
}

// under is p relative to base, or p itself when it is absolute: a go.mod or
// go.work path may be either.
func under(base, p string) string {
	if filepath.IsAbs(p) {
		return p
	}
	return filepath.Join(base, p)
}

func sortedKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

var (
	positionLine = regexp.MustCompile(`(?m)^/.*:\d+:\d+: `)
	fixFinding   = regexp.MustCompile(`(?m)^--- |^fix: \d+ of \d+ fixes skipped`)
)

// statusFor turns a finished check into pass, fail, or unavailable.
func statusFor(check string, code int, output string) string {
	if code == 124 || code == 126 || code == 127 || code == 137 || code < 0 {
		return "unavailable"
	}
	switch check {
	case "lint":
		if code != 0 {
			if code == 1 && positionLine.MatchString(output) {
				return "fail"
			}
			return "unavailable"
		}
	case "fix":
		// Overlapping fixes print no hunk, only "fix: N of M fixes skipped".
		if fixFinding.MatchString(output) {
			return "fail"
		}
		if code == 0 && strings.TrimSpace(output) == "" {
			return "pass"
		}
		return "unavailable"
	case "gofmt":
		if code != 0 || strings.TrimSpace(output) != "" {
			return "fail"
		}
		return "pass"
	}
	if code == 0 {
		return "pass"
	}
	return "fail"
}

// renderTestJSON turns go test -json events back into text for the
// transcript: build output, package lines, and the output of every test that
// did not pass or skip, so passing tests' logs stay out. ran reports whether
// any test, example, or fuzz seed passed: exit 0 alone does not show that one
// ran.
func renderTestJSON(stream string) (text string, ran bool) {
	type line struct {
		key  string
		text string
	}
	var lines []line
	settled := map[string]bool{}
	for _, raw := range strings.SplitAfter(stream, "\n") {
		if raw == "" {
			continue
		}
		var ev struct {
			Action, Package, Test, Output string
		}
		if json.Unmarshal([]byte(raw), &ev) != nil || !strings.HasPrefix(strings.TrimSpace(raw), "{") {
			lines = append(lines, line{"", raw})
			continue
		}
		key := ev.Package + "\x00" + ev.Test
		if ev.Test != "" && (ev.Action == "pass" || ev.Action == "skip") {
			settled[key] = true
			ran = ran || ev.Action == "pass"
		}
		if ev.Action == "output" || ev.Action == "build-output" {
			if ev.Test == "" {
				key = ""
			}
			lines = append(lines, line{key, ev.Output})
		}
	}
	var b strings.Builder
	for _, l := range lines {
		if l.key == "" || !settled[l.key] {
			b.WriteString(l.text)
		}
	}
	return b.String(), ran
}

func now() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05.000000-07:00")
}

// atomicJSON writes v to path through a rename, so a reader never sees half
// a record.
func atomicJSON(path string, v any) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	f, err := os.CreateTemp(filepath.Dir(path), ".pending-")
	if err != nil {
		return err
	}
	_, err = f.WriteString(encode(v))
	if err = errors.Join(err, f.Close()); err != nil {
		_ = os.Remove(f.Name()) // the write already failed
		return err
	}
	return os.Rename(f.Name(), path)
}

// begin opens an edit generation: the run directory becomes the latest of
// the session's verification state.
func begin(directory, file, session, stateBase string) error {
	state := filepath.Join(resolve(stateBase), "verification", digest([]byte(session)))
	info := generation{RunDir: resolve(directory), Package: filepath.Dir(resolve(file)), State: state}
	if err := atomicJSON(filepath.Join(directory, ".session"), info); err != nil {
		return err
	}
	if err := atomicJSON(filepath.Join(state, "latest"), info); err != nil {
		return err
	}
	// The speed gate only reads reports from sessions that changed code.
	if !strings.HasSuffix(file, "_test.go") {
		return touch(filepath.Join(state, "production"))
	}
	return nil
}

// markAttempt records a verifier attempt for the current generation; a
// failed attempt releases direct checks without granting credit.
func markAttempt(directory string, credits []string, replace bool) error {
	var info generation
	if err := readJSON(filepath.Join(directory, ".session"), &info); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	var latest generation
	if err := readJSON(filepath.Join(info.State, "latest"), &latest); err != nil {
		return err
	}
	if latest.RunDir != resolve(directory) {
		return nil
	}
	attemptedPath := filepath.Join(info.State, "attempted")
	if !replace {
		var previous generation
		if readJSON(attemptedPath, &previous) == nil && previous.RunDir == latest.RunDir {
			credits = dedupe(append(credits, previous.Credits...))
			sort.Strings(credits)
		}
	}
	return atomicJSON(attemptedPath, object{{"run_dir", latest.RunDir}, {"credits", nonNil(credits)}})
}

func nonNil(list []string) []string {
	if list == nil {
		return []string{}
	}
	return list
}

// run executes one check, or records one that was planned as skipped or
// unavailable, and writes its receipt. Arguments: run-dir, check, file,
// config, filter, planned status ("run" to execute), reason, command.
func run(args []string) (int, error) {
	directory, check, file, config, filterKind, planned, reason := args[0], args[1], args[2], args[3], args[4], args[5], args[6]
	command := args[7:]
	before, errorsBefore, version, toolchain := snapshot(file, check, config, filterKind, directory)
	started := now()
	var code *int
	output, status := "", planned
	if planned == "run" {
		switch {
		case len(command) == 0:
			return 0, errors.New("no command to run")
		case remaining() < 5*time.Second:
			status, reason = "unavailable", "hook deadline exceeded before check"
		default:
			// The per-check cap of go-vet-on-edit.sh's run_check, applied
			// here rather than through timeout(1), which macOS lacks.
			limit, _ := strconv.ParseFloat(os.Getenv("GOLANG_SKILLS_CHECK_LIMIT"), 64)
			cap := time.Duration(limit * float64(time.Second))
			capped := limit > 0 && cap < remaining()-4*time.Second
			if !capped {
				cap = max(time.Millisecond, remaining()-4*time.Second)
			}
			c, raw, timedOut, err := runCapped(command, filepath.Dir(resolve(file)), false, cap)
			switch {
			case err != nil:
				c, raw = 127, []byte(err.Error()+"\n")
			case timedOut && capped:
				raw = fmt.Appendf(raw, "\ncheck exceeded its %ss limit\n", strconv.FormatFloat(limit, 'g', -1, 64))
			case timedOut:
				raw = append(raw, "\nedit-hook global deadline exceeded\n"...)
			}
			code = &c
			output = strings.ToValidUTF8(string(raw), "�")
			status = statusFor(check, c, output)
			if check == "test" && slices.Contains(command, "-json") {
				var ran bool
				output, ran = renderTestJSON(output)
				if status == "pass" && !ran {
					status, reason = "skipped", "no test ran (TestMain without m.Run, or every test skipped)"
				}
			}
		}
	}
	ended := now()
	after, errorsAfter, _, _ := snapshot(file, check, config, filterKind, directory)
	absFile, _ := filepath.Abs(file)
	target := filepath.Dir(resolve(file))
	kind := "package"
	if check == "gofmt" {
		kind, target = "file", resolve(file)
	}
	r := receipt{
		Version: 1, Check: check, Status: status, ExitCode: code, Command: nonNil(command),
		Cwd: filepath.Dir(resolve(file)), File: absFile, Scope: scope{kind, target},
		Config: config, Filter: filterKind, ToolVersion: version, Toolchain: toolchain,
		Started: started, Ended: ended, InputsBefore: before, InputsAfter: after,
		Reusable:       status == "pass" && before == after && len(errorsBefore) == 0 && len(errorsAfter) == 0,
		Reason:         reason,
		SnapshotErrors: nonNil(append(errorsBefore, errorsAfter...)),
		Diagnostic:     output,
	}
	if err := atomicJSON(filepath.Join(directory, check+".json"), r); err != nil {
		return 0, err
	}
	fmt.Print(output)
	switch {
	case code == nil:
		return 0, nil
	case *code < 0:
		return 128 - *code, nil
	}
	return *code, nil
}

func readReceipt(p string) (receipt, error) {
	var r receipt
	err := readJSON(p, &r)
	if err == nil && r.Check == "" {
		err = errors.New("not a receipt: " + p)
	}
	return r, err
}

// current reports whether r's inputs still hash as they did around the check.
func current(r receipt, directory string) (bool, []string) {
	sum, problems, _, _ := snapshot(r.File, r.Check, r.Config, r.Filter, directory)
	return r.Version == 1 && r.Status == "pass" && r.Reusable && len(problems) == 0 &&
		r.InputsBefore == r.InputsAfter && r.InputsAfter == sum, problems
}

// lintConfigMatches reports a lint command that reads the config the receipt
// names: a repository whose own config the hook replaced earns no credit.
func lintConfigMatches(r receipt) bool {
	i := slices.Index(r.Command, "--config")
	return i >= 0 && i+1 < len(r.Command) && r.Command[i+1] == r.Config
}

// verify checks one receipt. With an expected cwd and command it is the
// Bash entry point, whose exit 0 grants hook credit; without, it is the
// state-only check, which never does.
func verify(path string, expected []string) (int, error) {
	r, err := readReceipt(path)
	if err != nil {
		return 0, err
	}
	directory := filepath.Dir(path)
	valid, problems := current(r, directory)
	// A hook test run is -short without -race, and its exit 0 does not show
	// that a test ran: it never stands in for a test check.
	matching := len(expected) >= 2 && r.Check != "test" &&
		resolve(expected[0]) == r.Cwd && slices.Equal(expected[1:], r.Command) &&
		(r.Check != "lint" || lintConfigMatches(r))
	credit := valid && matching
	if len(expected) > 0 {
		var credits []string
		if credit {
			credits = []string{r.Check}
		}
		if err := markAttempt(directory, credits, !valid); err != nil {
			return 0, err
		}
	}
	var reason any = ""
	switch {
	case len(problems) > 0:
		reason = problems
	case !valid:
		reason = "stale or non-passing receipt"
	case len(expected) > 0 && !matching:
		reason = "cwd or command differs from the recorded check"
	}
	fmt.Println(encode(object{
		{"valid", valid}, {"check", r.Check}, {"command", r.Command}, {"cwd", r.Cwd},
		{"scope", object{{"kind", r.Scope.Kind}, {"target", r.Scope.Target}}},
		{"config", r.Config}, {"filter", r.Filter}, {"hook_credit", credit}, {"reason", reason},
	}))
	if (len(expected) > 0 && credit) || (len(expected) == 0 && valid) {
		return 0, nil
	}
	return 1, nil
}

// verifyGate checks the package receipts of one generation together and
// credits each supported check that is current and matches the gate's form.
func verifyGate(directory, target string) (int, error) {
	target = resolve(target)
	var rows []object
	var credited, direct []string
	for _, check := range checkNames {
		r, err := readReceipt(filepath.Join(directory, check+".json"))
		if err != nil {
			rows = append(rows, object{{"check", check}, {"status", "unavailable"}, {"valid", false}, {"hook_credit", false}, {"reason", err.Error()}})
			direct = append(direct, check)
			continue
		}
		valid, problems := current(r, directory)
		// The batch selects package checks. File gofmt and short tests
		// still need the selected broader formatting and race gate.
		eligible := r.Cwd == target && r.Scope == scope{"package", target}
		switch check {
		case "vet":
			eligible = eligible && slices.Equal(r.Command, []string{"go", "vet", "."})
		case "fix":
			eligible = eligible && slices.Equal(r.Command, []string{"go", "fix", "-diff", "."})
		case "lint":
			want := []string{"golangci-lint", "run", "--allow-parallel-runners", "--path-mode=abs",
				"--output.text.print-issued-lines=false", "--show-stats=false", ".", "--config", r.Config}
			eligible = eligible && r.Filter == "all" && r.Config != "" && slices.Equal(r.Command, want)
		default:
			eligible = false
		}
		credit := valid && eligible
		var reason any = ""
		switch {
		case len(problems) > 0:
			reason = problems
		case !credit:
			reason = "run the selected check directly: stale, unavailable or broader scope/flags"
		}
		rows = append(rows, object{
			{"check", check}, {"status", r.Status}, {"valid", valid}, {"hook_credit", credit},
			{"scope", object{{"kind", r.Scope.Kind}, {"target", r.Scope.Target}}},
			{"command", r.Command}, {"config", r.Config}, {"filter", r.Filter}, {"reason", reason},
		})
		if credit {
			credited = append(credited, check)
		} else {
			direct = append(direct, check)
		}
	}
	if err := markAttempt(directory, credited, true); err != nil {
		return 0, err
	}
	fmt.Println(encode(object{
		{"scope", object{{"kind", "package"}, {"target", target}}},
		{"checks", rows},
		{"required_direct", nonNil(direct)},
		{"note", "No build, race or applicable govulncheck evidence in these receipts; complete the selected gate."},
	}))
	// A failed attempt releases direct checks but never grants credit.
	if len(credited) > 0 {
		return 0, nil
	}
	return 1, nil
}

// currentCredits is the subset of names whose receipts are still current.
func currentCredits(directory string, names []string) []string {
	var verified []string
	for _, name := range names {
		if !slices.Contains(checkNames, name) {
			continue
		}
		r, err := readReceipt(filepath.Join(directory, name+".json"))
		if err != nil {
			continue
		}
		if ok, _ := current(r, directory); ok {
			verified = append(verified, name)
		}
	}
	return verified
}

// report prints the receipts of a generation with the verifier command to
// run next: as PostToolUse JSON, or as plain text for stderr.
func report(directory string, contextOnly bool) error {
	root := os.Getenv("GOLANG_SKILLS_HOOK_ROOT")
	skill := "go-linting"
	if ns := pluginName(filepath.Join(root, ".claude-plugin", "plugin.json")); ns != "" {
		skill = ns + ":" + skill
	}
	var rows []string
	for _, check := range checkNames {
		r, err := readReceipt(filepath.Join(directory, check+".json"))
		if err != nil {
			return err
		}
		exit := "none"
		if r.ExitCode != nil {
			exit = strconv.Itoa(*r.ExitCode)
		}
		rows = append(rows, fmt.Sprintf("%s: %s; exit=%s; scope=%s; filter=%s; state-stable=%t; reuse=unverified; reason=%s",
			check, r.Status, exit, r.Scope.Kind, r.Filter, r.Reusable, r.Reason))
	}
	vet, err := readReceipt(filepath.Join(directory, "vet.json"))
	if err != nil {
		return err
	}
	batch := shellJoin([]string{"bash", filepath.Join(root, "hooks", "go-check-receipt.sh"), "--gate", resolve(directory), vet.Cwd})
	text := fmt.Sprintf("With a Bash tool, before the final report load the `%s` skill (Skill tool, name `%s`), then verify matching receipts below or run the selected checks directly.\n", skill, skill) +
		"Edit-hook receipts: " + directory + "\nReuse is UNVERIFIED; recorded pass is not pass (hook).\n" + strings.Join(rows, "\n") +
		"\nFirst verification action after the final edit (Bash tool):\n" + batch +
		"\nUse only hook_credit=true checks from that JSON. Never write lint pass for reuse=unverified, even with a disclaimer. Run required_direct checks and the selected build/race gate." +
		"\nFile-only gofmt and package checks cannot satisfy a wider gate. Package short tests have no race check. If verification is unavailable, run the check directly or report it unavailable."
	if contextOnly {
		fmt.Println(text)
	} else {
		fmt.Println(contextJSON("PostToolUse", text))
	}
	return nil
}
