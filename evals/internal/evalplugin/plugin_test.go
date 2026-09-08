package evalplugin

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCopyWithoutOptionalResources(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{".claude-plugin", "skills"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := Copy(root, t.TempDir()); err != nil {
		t.Errorf("Copy(plugin without agents or hooks) error = %v, want nil", err)
	}
}

func TestCopyKeepsReferencesInsideWorkAndAnswersOutside(t *testing.T) {
	root, work := t.TempDir(), t.TempDir()
	for _, name := range []string{".claude-plugin", "skills/go-http/references", "agents", "hooks", "evals/_golden"} {
		if err := os.MkdirAll(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	reference := filepath.Join("skills", "go-http", "references", "CLIENTS.md")
	for _, name := range []string{reference, "evals/_golden/answer.go", "README.md"} {
		if err := os.WriteFile(filepath.Join(root, name), []byte(name), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	dir, err := Copy(root, work)
	if err != nil {
		t.Fatalf("Copy() error = %v, want nil", err)
	}
	rel, err := filepath.Rel(work, filepath.Join(dir, reference))
	if err != nil || !filepath.IsLocal(rel) {
		t.Fatalf("copied reference relative to work = %q, %v; want a local path", rel, err)
	}
	got, err := os.ReadFile(filepath.Join(dir, reference))
	if err != nil || string(got) != reference {
		t.Errorf("copied reference = %q, %v; want %q, nil", got, err, reference)
	}
	for _, name := range []string{"evals", "README.md"} {
		if _, err := os.Stat(filepath.Join(dir, name)); !os.IsNotExist(err) {
			t.Errorf("copied %s stat error = %v, want not-exist", name, err)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, reference), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err = os.ReadFile(filepath.Join(root, reference))
	if err != nil || string(got) != reference {
		t.Errorf("original reference after copy edit = %q, %v; want %q, nil", got, err, reference)
	}
}
