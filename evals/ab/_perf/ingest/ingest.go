// Package ingest imports records into the remote record store.
package ingest

import (
	"context"
	"fmt"
)

// Import stores records through s in order and returns how many it stored. A
// record that appears twice keeps the later value. At the first record the
// store rejects, Import stops and returns the count stored before it and an
// error naming the record's index.
func Import(ctx context.Context, s *Store, records []Record) (int, error) {
	for i, r := range records {
		if err := s.Put(ctx, r); err != nil {
			return i, fmt.Errorf("record %d: %w", i, err)
		}
	}
	return len(records), nil
}

// Summarize describes an import for the job's log line.
func Summarize(records []Record) string {
	empty := 0
	for _, r := range records {
		if r.Value == "" {
			empty = empty + 1
		}
	}
	s := fmt.Sprintf("%d records", len(records))
	if empty > 0 {
		s = s + fmt.Sprintf(", %d with an empty value", empty)
	}
	return s
}
