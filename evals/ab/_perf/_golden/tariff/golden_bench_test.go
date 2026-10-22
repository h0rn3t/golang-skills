package tariff

import (
	"strconv"
	"strings"
	"testing"
)

// BenchmarkGoldenPrice prices a different cart every iteration, as checkout
// does: no two calls ask the same question.
func BenchmarkGoldenPrice(b *testing.B) {
	var rules strings.Builder
	for i := range 150 {
		rules.WriteString("SKU-" + strconv.Itoa(2000+i*3) + " " + strconv.Itoa(5+i%30) + "\n")
	}
	book := NewBook(rules.String())
	skus := make([]string, 500)
	for i := range skus {
		skus[i] = "SKU-" + strconv.Itoa(2000+i) + "-v"
	}
	cart := make([]Item, 6)
	n := 0
	for b.Loop() {
		for j := range cart {
			n++
			cart[j] = Item{SKU: skus[n%len(skus)], Qty: 1 + n%7, Cents: int64(100 + n%100000)}
		}
		if _, err := book.Price(cart); err != nil {
			b.Fatal(err)
		}
	}
}
