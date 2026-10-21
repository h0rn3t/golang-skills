package events

import (
	"io"
	"testing"
	"time"
)

func BenchmarkGoldenWriteAll(b *testing.B) {
	events := make([]Event, 1000)
	for i := range events {
		events[i] = Event{ID: int64(i), Kind: "payment", Actor: "Customer-42", At: time.Date(2026, 10, 21, 9, 30, i, 0, time.UTC), Tags: []string{"b", "a"}, Amount: float64(i) * 1.25}
	}
	for b.Loop() {
		if err := WriteAll(io.Discard, events); err != nil {
			b.Fatal(err)
		}
	}
}
