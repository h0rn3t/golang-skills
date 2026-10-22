package tariff

import (
	"strconv"
	"strings"
	"testing"
)

func BenchmarkPrice(b *testing.B) {
	var rules strings.Builder
	for i := range 120 {
		rules.WriteString("SKU-" + strconv.Itoa(1000+i) + " " + strconv.Itoa(i%40) + "\n")
	}
	book := NewBook(rules.String())
	cart := []Item{
		{SKU: "SKU-1007-red", Qty: 2, Cents: 1999},
		{SKU: "SKU-1050-xl", Qty: 1, Cents: 4500},
		{SKU: "GIFT-CARD", Qty: 1, Cents: 2500},
		{SKU: "SKU-1119", Qty: 3, Cents: 799},
		{SKU: "SKU-1063-blue", Qty: 1, Cents: 12900},
	}
	for b.Loop() {
		if _, err := book.Price(cart); err != nil {
			b.Fatal(err)
		}
	}
}
