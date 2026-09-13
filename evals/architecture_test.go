package evals_test

import (
	"encoding/json"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The architecture checker ships as Go source next to its wrapper; its unit
// tests run as a command-line-arguments package, no module needed.
func TestArchitectureCheckerUnitTests(t *testing.T) {
	t.Parallel()
	dir := filepath.Join(repoRoot(t), "skills", "go-code-refactor", "scripts")
	for _, args := range [][]string{
		{"vet", "./check-architecture.go", "./check-architecture_test.go"},
		{"test", "-count=1", "./check-architecture.go", "./check-architecture_test.go"},
	} {
		cmd := exec.CommandContext(t.Context(), "go", args...)
		cmd.Dir = dir
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("go %s: %v\n%s", args[0], err, out)
		}
	}
}

type archReport struct {
	Module     string `json:"module"`
	Layout     string `json:"layout"`
	Total      int    `json:"total"`
	Suppressed int    `json:"suppressed"`
	Status     string `json:"status"`
	Violations []struct {
		Rule string `json:"rule"`
		From string `json:"from"`
		To   string `json:"to"`
	} `json:"violations"`
}

func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatal(err)
	}
}

func runArchCheck(t *testing.T, wantExit int, args ...string) archReport {
	t.Helper()
	script := filepath.Join(repoRoot(t), "skills", "go-code-refactor", "scripts", "check-architecture.sh")
	cmd := exec.CommandContext(t.Context(), "bash", append([]string{script, "--json"}, args...)...)
	out, err := cmd.Output()
	code := 0
	if exit, ok := err.(*exec.ExitError); ok {
		code = exit.ExitCode()
	} else if err != nil {
		t.Fatalf("run checker: %v", err)
	}
	if code != wantExit {
		t.Fatalf("checker exit %d, want %d\n%s", code, wantExit, out)
	}
	var rep archReport
	if wantExit != 2 {
		if err := json.Unmarshal(out, &rep); err != nil {
			t.Fatalf("parse checker JSON: %v\n%s", err, out)
		}
	}
	return rep
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// The fixture compiles, its rollback test passes, and the checker reads it
// as clean; one forbidden edge of each kind added to a copy is reported by
// rule, an exact known entry suppresses it, and a stale entry fails.
func TestArchitectureFixtureAndChecker(t *testing.T) {
	t.Parallel()
	fixture := filepath.Join(repoRoot(t), "skills", "go-code-refactor", "testdata", "architecture")
	for _, args := range [][]string{{"vet", "./..."}, {"test", "-count=1", "./..."}} {
		cmd := exec.CommandContext(t.Context(), "go", args...)
		cmd.Dir = fixture
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("fixture go %s: %v\n%s", args[0], err, out)
		}
	}
	clean := runArchCheck(t, 0, fixture)
	if clean.Total != 0 || clean.Module != "example.com/shop" || clean.Layout != "modules" {
		t.Fatalf("clean fixture reported %+v", clean)
	}

	t.Run("foreign implementation import", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		copyTree(t, fixture, dir)
		writeFile(t, filepath.Join(dir, "internal", "billing", "services", "leak.go"),
			"package services\n\nimport \"example.com/shop/internal/order/repositories\"\n\nvar _ = repositories.NewOrderStore\n")
		rep := runArchCheck(t, 1, dir)
		if len(rep.Violations) != 1 || rep.Violations[0].Rule != "ownership" || rep.Violations[0].From != "internal/billing/services" || rep.Violations[0].To != "internal/order/repositories" {
			t.Fatalf("violations = %+v, want one ownership edge billing/services -> order/repositories", rep.Violations)
		}
		writeFile(t, filepath.Join(dir, "architecture.json"), `{"layout":"modules","platform":["internal/platform/clock"],"modules":{"order":"layered","billing":"layered"},
"known":[{"rule":"ownership","from":"internal/billing/services","to":"internal/order/repositories","reason":"legacy export path","owner":"billing","until":"2026-12-31"}]}`)
		if rep := runArchCheck(t, 0, dir); rep.Suppressed != 1 || rep.Total != 0 {
			t.Fatalf("known entry did not suppress the edge: %+v", rep)
		}
		if err := os.Remove(filepath.Join(dir, "internal", "billing", "services", "leak.go")); err != nil {
			t.Fatal(err)
		}
		if rep := runArchCheck(t, 1, dir); len(rep.Violations) != 1 || rep.Violations[0].Rule != "stale" {
			t.Fatalf("a fixed violation left in known must fail as stale: %+v", rep.Violations)
		}
	})

	t.Run("handler to repository shortcut", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		copyTree(t, fixture, dir)
		writeFile(t, filepath.Join(dir, "internal", "order", "handlers", "v2", "shortcut.go"),
			"package v2\n\nimport \"example.com/shop/internal/order/repositories\"\n\nvar _ = repositories.NewOrderStore\n")
		rep := runArchCheck(t, 1, dir)
		if len(rep.Violations) != 1 || rep.Violations[0].Rule != "layer" || rep.Violations[0].From != "internal/order/handlers/v2" {
			t.Fatalf("violations = %+v, want one layer edge from the nested handlers/v2 package", rep.Violations)
		}
	})

	t.Run("driver and composition", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		copyTree(t, fixture, dir)
		// app already imports order/services, so the composition edge has to
		// come from a package app does not import, or go list reports a cycle.
		writeFile(t, filepath.Join(dir, "internal", "order", "services", "leak.go"),
			"package services\n\nimport \"database/sql\"\n\nvar _ = sql.Open\n")
		writeFile(t, filepath.Join(dir, "internal", "order", "handlers", "admin", "admin.go"),
			"package admin\n\nimport \"example.com/shop/internal/app\"\n\nvar _ = app.New\n")
		rep := runArchCheck(t, 1, dir)
		var rules []string
		for _, v := range rep.Violations {
			rules = append(rules, v.Rule)
		}
		if got := strings.Join(rules, ","); got != "composition,driver" && got != "driver,composition" {
			t.Fatalf("rules = %q, want composition and driver", got)
		}
	})

	t.Run("unclassified package and bad flags", func(t *testing.T) {
		t.Parallel()
		dir := t.TempDir()
		copyTree(t, fixture, dir)
		writeFile(t, filepath.Join(dir, "internal", "util", "util.go"), "package util\n")
		if rep := runArchCheck(t, 1, dir); len(rep.Violations) != 1 || rep.Violations[0].Rule != "unclassified" {
			t.Fatalf("violations = %+v, want one unclassified package", rep.Violations)
		}
		runArchCheck(t, 2, "--limit", "nope", dir)
		if err := os.Remove(filepath.Join(dir, "architecture.json")); err != nil {
			t.Fatal(err)
		}
		runArchCheck(t, 2, dir)
	})
}
