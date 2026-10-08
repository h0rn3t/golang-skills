package purge

import (
	"errors"
	"os"
	"path/filepath"
	"slices"
	"testing"
)

const (
	goldenAcme   = "3f0c8a8e-7c1d-4a8b-9d6e-1f2a3b4c5d6e"
	goldenGlobex = "9b2d4f6a-1e3c-4d5b-8a7f-0c1e2d3f4a5b"
)

// goldenTree creates two workspaces with a file each and returns the
// directory and the sorted list of every path under it.
func goldenTree(t *testing.T) (string, func() []string) {
	t.Helper()
	dir := t.TempDir()
	for _, ws := range []string{goldenAcme, goldenGlobex} {
		if err := os.MkdirAll(filepath.Join(dir, ws, "data"), 0o750); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, ws, "data", "report.csv"), []byte("x"), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	return dir, func() []string {
		var paths []string
		_ = filepath.WalkDir(dir, func(p string, _ os.DirEntry, err error) error {
			if err == nil {
				rel, _ := filepath.Rel(dir, p)
				paths = append(paths, rel)
			}
			return nil
		})
		slices.Sort(paths)
		return paths
	}
}

func goldenOpen(t *testing.T, dir string) *Workspaces {
	t.Helper()
	ws, err := Open(dir)
	if err != nil {
		t.Fatalf("Open(%q) error = %v", dir, err)
	}
	t.Cleanup(func() {
		if err := ws.Close(); err != nil {
			t.Errorf("Close() error = %v", err)
		}
	})
	return ws
}

func TestDeleteRemovesOneWorkspace(t *testing.T) {
	dir, tree := goldenTree(t)
	if err := goldenOpen(t, dir).Delete(goldenAcme); err != nil {
		t.Fatalf("Delete(%s) error = %v", goldenAcme, err)
	}
	want := []string{".", goldenGlobex, filepath.Join(goldenGlobex, "data"), filepath.Join(goldenGlobex, "data", "report.csv")}
	if got := tree(); !slices.Equal(got, want) {
		t.Errorf("tree after Delete(%s) = %v, want %v", goldenAcme, got, want)
	}
}

func TestDeleteMissingWorkspace(t *testing.T) {
	dir, tree := goldenTree(t)
	before := tree()
	if err := goldenOpen(t, dir).Delete("00000000-0000-4000-8000-000000000000"); err != nil {
		t.Errorf("Delete(missing workspace) error = %v, want nil", err)
	}
	if got := tree(); !slices.Equal(got, before) {
		t.Errorf("tree after Delete(missing workspace) = %v, want %v", got, before)
	}
}

// TestDeleteRejectsNamesThatAreNotOneWorkspace is the trap. A {id} wildcard
// hands over a/.. for a%2F.., and that name, "", and a nested path all pass a
// lexical locality check while naming the whole directory or part of another
// workspace; RemoveAll then deletes every tenant's files.
func TestDeleteRejectsNamesThatAreNotOneWorkspace(t *testing.T) {
	for _, id := range []string{
		"",
		".",
		"..",
		"a/..",
		goldenGlobex + "/..",
		goldenGlobex + "/data",
		"../" + goldenGlobex,
		"/" + goldenGlobex,
		"report",
	} {
		dir, tree := goldenTree(t)
		before := tree()
		err := goldenOpen(t, dir).Delete(id)
		if !errors.Is(err, ErrInvalidID) {
			t.Errorf("Delete(%q) error = %v, want one wrapping ErrInvalidID", id, err)
		}
		if got := tree(); !slices.Equal(got, before) {
			t.Errorf("Delete(%q) removed files: tree = %v, want %v", id, got, before)
		}
	}
}
