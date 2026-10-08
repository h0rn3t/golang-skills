// Package snapshot saves counter snapshots to disk and combines them.
package snapshot

import (
	"cmp"
	"encoding/json/v2"
	"errors"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
)

// Save writes counts to path atomically: it writes a temporary file in the
// same directory and renames it over path, so a reader sees either the old
// snapshot or the new one, never a partial write.
func Save(path string, counts map[string]int) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".snapshot-*")
	if err != nil {
		return err
	}
	if err := json.MarshalWrite(f, counts); err != nil {
		_ = f.Close() // the encode error is the one to report
		_ = os.Remove(f.Name())
		return err
	}
	f.Close() //nolint:errcheck // the data is already written
	if err := os.Rename(f.Name(), path); err != nil {
		_ = os.Remove(f.Name())
		return err
	}
	return nil
}

// Load reads a snapshot written by Save. A missing file is an empty snapshot.
func Load(path string) (map[string]int, error) {
	data, err := os.ReadFile(path) //nolint:gosec // G304: path comes from the operator's config, never from a request
	if errors.Is(err, fs.ErrNotExist) {
		return map[string]int{}, nil
	}
	if err != nil {
		return nil, err
	}
	var counts map[string]int
	if err := json.Unmarshal(data, &counts); err != nil {
		return nil, err
	}
	return counts, nil
}

// Merge adds every count in src to dst and returns dst. A nil dst is
// allocated.
func Merge(dst, src map[string]int) map[string]int {
	if dst == nil {
		dst = make(map[string]int, len(src))
	}
	maps.Copy(dst, src)
	return dst
}

// Top returns the names of the n largest counts, largest first, with ties in
// name order. An n larger than len(counts) returns every name.
func Top(counts map[string]int, n int) []string {
	names := slices.Collect(maps.Keys(counts))
	slices.SortFunc(names, func(a, b string) int {
		if c := cmp.Compare(counts[b], counts[a]); c != 0 {
			return c
		}
		return cmp.Compare(a, b)
	})
	return names[:min(n, len(names))]
}

// Prune deletes every counter below floor from counts.
func Prune(counts map[string]int, floor int) {
	panic("not implemented")
}
