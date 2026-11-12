package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// TestClaudeRunsWithoutOperatorContext pins that a session and the judge get
// neither the operator's CLAUDE.md files nor auto memory: both would reach
// every arm, the control included, on top of the plugin under test.
func TestClaudeRunsWithoutOperatorContext(t *testing.T) {
	bin := t.TempDir()
	cli := filepath.Join(bin, "claude")
	writeTestFile(t, cli, "#!/bin/sh\nprintf '%s %s' \"$CLAUDE_CODE_DISABLE_CLAUDE_MDS\" \"$CLAUDE_CODE_DISABLE_AUTO_MEMORY\"\n")
	if err := os.Chmod(cli, 0o700); err != nil { //nolint:gosec // G302: the fake CLI has to be executable
		t.Fatal(err)
	}
	t.Setenv("PATH", bin+string(os.PathListSeparator)+os.Getenv("PATH"))

	got, err := claude(time.Minute, t.TempDir())

	if err != nil || string(got) != "1 1" {
		t.Errorf("claude() saw CLAUDE_CODE_DISABLE_CLAUDE_MDS and CLAUDE_CODE_DISABLE_AUTO_MEMORY as %q, err = %v; want \"1 1\"", got, err)
	}
}
