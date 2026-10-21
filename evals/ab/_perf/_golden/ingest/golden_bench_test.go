package ingest

import (
	"strconv"
	"testing"
	"time"
)

func BenchmarkGoldenImport(b *testing.B) {
	records := make([]Record, 200)
	for i := range records {
		records[i] = Record{ID: "r" + strconv.Itoa(i), Value: "v"}
	}
	for b.Loop() {
		if _, err := Import(b.Context(), NewStore(100*time.Microsecond), records); err != nil {
			b.Fatal(err)
		}
	}
}
