package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

// patchFiles returns add/update destinations; deleted files have no content to check.
func patchFiles(patch string) []string {
	patch = strings.TrimSpace(strings.ReplaceAll(patch, "\r\n", "\n"))
	if !strings.HasPrefix(patch, "*** Begin Patch\n") || !strings.HasSuffix(patch, "\n*** End Patch") {
		return nil
	}
	lines := strings.Split(patch, "\n")
	var paths []string
	for i := 1; i < len(lines)-1; i++ {
		var path string
		switch {
		case strings.HasPrefix(lines[i], "*** Add File: "):
			path = strings.TrimPrefix(lines[i], "*** Add File: ")
		case strings.HasPrefix(lines[i], "*** Update File: "):
			path = strings.TrimPrefix(lines[i], "*** Update File: ")
			if i+1 < len(lines)-1 && strings.HasPrefix(lines[i+1], "*** Move to: ") {
				i++
				path = strings.TrimPrefix(lines[i], "*** Move to: ")
			}
		}
		if path != "" && !strings.ContainsAny(path, "\x00\r\n") {
			paths = append(paths, path)
		}
	}
	return dedupe(paths)
}

// codexEdit adapts one multi-file apply_patch event to the existing file checker.
func codexEdit() int {
	d, ok := readEvent()
	if !ok || d.str("hook_event_name") != "PostToolUse" || d.str("tool_name") != "apply_patch" {
		return 0
	}
	var contexts []string
	var diagnostics strings.Builder
	for _, path := range patchFiles(d.toolInput().str("command")) {
		if filepath.Ext(path) != ".go" {
			continue
		}
		if !filepath.IsAbs(path) {
			path = filepath.Join(d.str("cwd"), path)
		}
		path, err := filepath.Abs(path)
		if err != nil {
			fmt.Fprintf(&diagnostics, "%s: resolve edit path: %v\n", path, err)
			continue
		}
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		// Preserve session/agent identity so receipts still belong to this turn.
		d["tool_name"] = "Write"
		d["tool_input"] = map[string]any{"file_path": path}
		body, err := json.Marshal(d)
		if err != nil {
			fmt.Fprintf(&diagnostics, "%s: encode edit event: %v\n", path, err)
			continue
		}
		ctx, cancel := context.WithTimeout(context.Background(), 150*time.Second)
		cmd := exec.CommandContext(ctx, "bash", "hooks/go-vet-on-edit.sh")
		cmd.Dir = os.Getenv("GOLANG_SKILLS_HOOK_ROOT")
		cmd.Stdin = bytes.NewReader(body)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		err = cmd.Run()
		cancel()
		diagnostics.Write(stderr.Bytes())
		if err != nil {
			if stderr.Len() == 0 {
				fmt.Fprintf(&diagnostics, "%s: edit checks failed: %v\n", path, err)
			}
			continue
		}
		var result struct {
			HookSpecificOutput struct {
				AdditionalContext string `json:"additionalContext"`
			} `json:"hookSpecificOutput"`
		}
		if err := json.Unmarshal(stdout.Bytes(), &result); err != nil {
			fmt.Fprintf(&diagnostics, "%s: edit checks returned invalid JSON: %v\n", path, err)
			continue
		}
		contexts = append(contexts, result.HookSpecificOutput.AdditionalContext)
	}
	context := strings.Join(contexts, "\n")
	if diagnostics.Len() != 0 {
		// Exit 2 makes stdout invisible to the host; keep passing receipts visible too.
		fmt.Fprintln(os.Stderr, context)
		fmt.Fprint(os.Stderr, diagnostics.String())
		return 2
	}
	if context != "" {
		fmt.Println(contextJSON("PostToolUse", context))
	}
	return 0
}
