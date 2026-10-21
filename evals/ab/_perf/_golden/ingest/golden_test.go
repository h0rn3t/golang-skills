package ingest

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestGoldenImport(t *testing.T) {
	s := NewStore(time.Microsecond)
	records := make([]Record, 0, 1203)
	for i := range 1200 {
		records = append(records, Record{ID: "r" + strconv.Itoa(i%700), Value: "v" + strconv.Itoa(i)})
	}
	records = append(records, Record{ID: "r5", Value: "last"}, Record{ID: "r6", Value: ""}, Record{ID: "r7", Value: "x"})
	n, err := Import(t.Context(), s, records)
	if n != len(records) || err != nil {
		t.Fatalf("Import(%d records) = %d, %v; want %d, nil", len(records), n, err, len(records))
	}
	for id, want := range map[string]string{"r5": "last", "r6": "", "r7": "x", "r100": "v800", "r699": "v699"} {
		if got, ok := s.Get(id); !ok || got != want {
			t.Errorf("Get(%q) = %q, %v; want %q, true", id, got, ok, want)
		}
	}
	if n, err := Import(t.Context(), s, nil); n != 0 || err != nil {
		t.Errorf("Import(nil) = %d, %v; want 0, nil", n, err)
	}
}

func TestGoldenImportRejected(t *testing.T) {
	s := NewStore(time.Microsecond)
	records := make([]Record, 0, 800)
	for i := range 800 {
		records = append(records, Record{ID: "id" + strconv.Itoa(i), Value: "v"})
	}
	records[603].ID = ""
	n, err := Import(t.Context(), s, records)
	if n != 603 || !errors.Is(err, ErrInvalid) || !strings.Contains(err.Error(), "record 603") {
		t.Fatalf("Import(invalid at 603) = %d, %v; want 603 and an ErrInvalid naming record 603", n, err)
	}
	if _, ok := s.Get("id602"); !ok {
		t.Error("record 602, before the rejected one, was not stored")
	}
	if _, ok := s.Get("id604"); ok {
		t.Error("record 604, after the rejected one, was stored")
	}
}

func TestGoldenImportCanceled(t *testing.T) {
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	n, err := Import(ctx, NewStore(time.Millisecond), []Record{{ID: "a"}, {ID: "b"}})
	if n != 0 || !errors.Is(err, context.Canceled) {
		t.Errorf("Import(canceled) = %d, %v; want 0, context.Canceled", n, err)
	}
}

func TestGoldenSummarize(t *testing.T) {
	if got, want := Summarize([]Record{{ID: "a", Value: "1"}, {ID: "b"}}), "2 records, 1 with an empty value"; got != want {
		t.Errorf("Summarize() = %q, want %q", got, want)
	}
	if got, want := Summarize(nil), "0 records"; got != want {
		t.Errorf("Summarize(nil) = %q, want %q", got, want)
	}
}
