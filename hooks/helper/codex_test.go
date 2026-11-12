package main

import (
	"slices"
	"strings"
	"testing"
)

func TestPatchFiles(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		patch string
		want  []string
	}{
		{name: "empty"},
		{name: "not a patch", patch: "echo '*** Update File: wrong.go'"},
		{name: "incomplete patch", patch: "*** Begin Patch\n*** Add File: wrong.go\n+package p"},
		{name: "add and update", patch: "*** Begin Patch\n*** Add File: a.go\n+package p\n*** Update File: dir with spaces/b.go\n@@\n-old\n+new\n*** End Patch", want: []string{"a.go", "dir with spaces/b.go"}},
		{name: "move destination", patch: "*** Begin Patch\n*** Update File: old.go\n*** Move to: new.go\n@@\n-old\n+new\n*** End Patch", want: []string{"new.go"}},
		{name: "deletion", patch: "*** Begin Patch\n*** Delete File: removed.go\n*** End Patch"},
		{name: "content cannot inject a header", patch: "*** Begin Patch\n*** Add File: real.go\n+*** Update File: decoy.go\n+*** Move to: decoy2.go\n*** End Patch", want: []string{"real.go"}},
		{name: "duplicates", patch: "*** Begin Patch\n*** Update File: a.go\n@@\n-x\n+y\n*** Update File: a.go\n@@\n-y\n+z\n*** End Patch", want: []string{"a.go"}},
		{name: "absolute and shell characters", patch: "*** Begin Patch\n*** Add File: /tmp/$(touch secret); 'quoted'.go\n+package p\n*** End Patch", want: []string{"/tmp/$(touch secret); 'quoted'.go"}},
		{name: "CRLF", patch: "*** Begin Patch\r\n*** Add File: a.go\r\n+package p\r\n*** End Patch\r\n", want: []string{"a.go"}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := patchFiles(tt.patch); !slices.Equal(got, tt.want) {
				t.Errorf("patchFiles(%q) = %q, want %q", tt.patch, got, tt.want)
			}
		})
	}
}

func FuzzPatchFiles(f *testing.F) {
	f.Add("*** Begin Patch\n*** Add File: a.go\n+package p\n*** End Patch")
	f.Add("*** Begin Patch\n*** Update File: old.go\n*** Move to: new.go\n*** End Patch")
	f.Fuzz(func(t *testing.T, patch string) {
		seen := map[string]bool{}
		for _, path := range patchFiles(patch) {
			if path == "" || strings.ContainsAny(path, "\x00\r\n") || seen[path] {
				t.Fatalf("patchFiles(%q) returned invalid or duplicate path %q", patch, path)
			}
			seen[path] = true
		}
	})
}
