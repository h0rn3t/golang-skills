package tariff

import (
	"errors"
	"sync"
	"testing"
)

const goldenRules = "# spring sale\nSKU-10 25\n\nSKU-1 10%\nGIFT 0\n"

var goldenCart = []Item{
	{SKU: "SKU-105", Qty: 3, Cents: 999},   // SKU-10: 2997 - 749 = 2248
	{SKU: "SKU-17", Qty: 1, Cents: 1001},   // SKU-1: 1001 - 100 = 901
	{SKU: "GIFT-50", Qty: 2, Cents: 5000},  // GIFT: 10000
	{SKU: "BOOK-1", Qty: 1, Cents: 1234},   // no rule: 1234
	{SKU: "SKU-100", Qty: 0, Cents: 99999}, // empty line: 0
}

func TestGoldenPrice(t *testing.T) {
	book := NewBook(goldenRules)
	if got, err := book.Price(goldenCart); got != 14383 || err != nil {
		t.Errorf("Price(goldenCart) = %d, %v; want 14383, nil", got, err)
	}
	if got, err := book.Price(nil); got != 0 || err != nil {
		t.Errorf("Price(nil) = %d, %v; want 0, nil", got, err)
	}

	book.SetRules("SKU 50\n")
	if got, err := book.Price(goldenCart); got != 1499+501+10000+1234 || err != nil {
		t.Errorf("Price(goldenCart) after SetRules = %d, %v; want %d, nil", got, err, 1499+501+10000+1234)
	}
	book.SetRules(goldenRules)
	if got, err := book.Price(goldenCart); got != 14383 || err != nil {
		t.Errorf("Price(goldenCart) after restoring the rules = %d, %v; want 14383, nil", got, err)
	}
}

func TestGoldenBadRules(t *testing.T) {
	for _, rules := range []string{"SKU-1\n", "SKU-1 ten\n", "SKU-1 101\n", "SKU-1 -5\n", "SKU-1 5 extra\n"} {
		if _, err := NewBook(rules).Price(goldenCart); !errors.Is(err, ErrBadRule) {
			t.Errorf("NewBook(%q).Price() error = %v, want %v", rules, err, ErrBadRule)
		}
	}
	book := NewBook(goldenRules)
	book.SetRules("SKU-1 ten\n")
	if _, err := book.Price(goldenCart); !errors.Is(err, ErrBadRule) {
		t.Errorf("Price() after SetRules(bad) error = %v, want %v", err, ErrBadRule)
	}
	book.SetRules(goldenRules)
	if got, err := book.Price(goldenCart); got != 14383 || err != nil {
		t.Errorf("Price() after SetRules(good) = %d, %v; want 14383, nil", got, err)
	}
}

func TestGoldenConcurrentSetRules(t *testing.T) {
	book := NewBook(goldenRules)
	var wg sync.WaitGroup
	wg.Go(func() {
		for i := range 200 {
			if i%2 == 0 {
				book.SetRules("SKU 50\n")
			} else {
				book.SetRules(goldenRules)
			}
		}
	})
	for range 4 {
		wg.Go(func() {
			for range 200 {
				got, err := book.Price(goldenCart)
				if err != nil || (got != 14383 && got != 13234) {
					t.Errorf("Price(goldenCart) during SetRules = %d, %v; want 14383 or 13234", got, err)
					return
				}
			}
		})
	}
	wg.Wait()
}

func TestGoldenCurrencySymbol(t *testing.T) {
	for code, want := range map[string]string{"usd": "$", "UAH": "₴", "CHF": "CHF"} {
		if got := CurrencySymbol(code); got != want {
			t.Errorf("CurrencySymbol(%q) = %q, want %q", code, got, want)
		}
	}
}
