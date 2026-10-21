package quotes

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"testing"
)

func BenchmarkGoldenHandle(b *testing.B) {
	prices := map[string]int64{}
	req := Request{Customer: "  maria   de la  CRUZ "}
	for i := range 40 {
		sku := fmt.Sprintf("SKU-%d%03d", i%4, i*7)
		prices[sku] = int64(100 + i*25)
		req.Lines = append(req.Lines, Line{SKU: sku, Qty: 1 + i%5})
	}
	var rules []Rule
	for i := range 30 {
		rules = append(rules, Rule{Pattern: fmt.Sprintf("^SKU-%d[0-9]*%d$", i%4, i%10), Percent: 1 + i%15})
	}
	body, err := json.Marshal(req)
	if err != nil {
		b.Fatal(err)
	}
	for b.Loop() {
		if err := Handle(io.Discard, bytes.NewReader(body), prices, rules); err != nil {
			b.Fatal(err)
		}
	}
}
