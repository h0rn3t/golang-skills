// Package evalplugin stages plugin resources for isolated evaluation sessions.
package evalplugin

import (
	"fmt"
	"os"
	"path/filepath"
)

// Env returns the environment a measured claude session runs in: the
// operator's own, with CLAUDE.md loading and auto memory switched off.
// --restricted ignores the settings files only, so without these a CLAUDE.md
// in the operator's home or above the scratch directory, and the memory of an
// earlier run, would reach every arm on top of the plugin under test.
func Env() []string {
	return append(os.Environ(), "CLAUDE_CODE_DISABLE_CLAUDE_MDS=1", "CLAUDE_CODE_DISABLE_AUTO_MEMORY=1")
}

// Copy places plugin resources inside work so restricted file tools can read
// references without exposing the repository's evaluation fixtures or answers.
// The caller owns the lifetime of work and the copied directory.
func Copy(root, work string) (string, error) {
	dir, err := os.MkdirTemp(work, ".eval-plugin-")
	if err != nil {
		return "", err
	}
	for _, name := range []string{".claude-plugin", "skills", "agents", "hooks"} {
		src := filepath.Join(root, name)
		if _, err := os.Stat(src); os.IsNotExist(err) && (name == "agents" || name == "hooks") {
			continue
		}
		if err := os.CopyFS(filepath.Join(dir, name), os.DirFS(src)); err != nil {
			return "", fmt.Errorf("copy plugin %s: %w", name, err)
		}
	}
	return dir, nil
}
