package statement

import (
	"io"
	"testing"
	"time"
)

func BenchmarkGoldenWrite(b *testing.B) {
	entries := make([]Entry, 5000)
	for i := range entries {
		entries[i] = Entry{ID: int64(100000 + i), Booked: time.Date(2026, 10, 1+i%28, 12, 0, 0, 0, time.UTC), Memo: "Card payment, store 42", Amount: int64(i*37) - 90000}
	}
	for b.Loop() {
		if err := Write(io.Discard, entries); err != nil {
			b.Fatal(err)
		}
	}
}
