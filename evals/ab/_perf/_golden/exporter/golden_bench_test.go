package exporter

import (
	"io"
	"strconv"
	"testing"
)

func BenchmarkGoldenExport(b *testing.B) {
	rows := make([]Row, 6000)
	for i := range rows {
		rows[i] = Row{OrderID: int64(i), Customer: "customer-" + strconv.Itoa(i*7919%4500), SKU: "SKU-" + strconv.Itoa(i%300), Qty: 1 + i%5, Cents: int64(i*37 - 900)}
	}
	for b.Loop() {
		if err := Export(io.Discard, rows); err != nil {
			b.Fatal(err)
		}
	}
}
