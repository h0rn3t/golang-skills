// Package tariff prices checkout carts against the discount rules the pricing
// team edits from the admin page.
package tariff

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

// Item is one cart line.
type Item struct {
	SKU   string
	Qty   int
	Cents int64 // unit price
}

// ErrBadRule reports a rules line Price cannot read.
var ErrBadRule = errors.New("tariff: bad rule")

// Book holds the current rules. It is safe for concurrent use.
type Book struct {
	mu    sync.RWMutex
	rules string
}

// NewBook returns a Book holding rules; see SetRules for the format.
func NewBook(rules string) *Book {
	return &Book{rules: rules}
}

// SetRules replaces the rules. Each non-empty line is "<sku-prefix> <percent>",
// and a line starting with # is a comment. The admin page calls it when the
// pricing team saves; checkout keeps calling Price meanwhile.
func (b *Book) SetRules(rules string) {
	b.mu.Lock()
	b.rules = rules
	b.mu.Unlock()
}

// Price returns the cart total in cents. The first rule whose prefix starts an
// item's SKU discounts that item's line (unit price times quantity) by its
// percent, rounded down to whole cents. Price returns an error wrapping
// ErrBadRule when the current rules do not parse.
func (b *Book) Price(cart []Item) (int64, error) {
	b.mu.RLock()
	text := b.rules
	b.mu.RUnlock()
	rules, err := parseRules(text)
	if err != nil {
		return 0, err
	}
	var total int64
	for _, it := range cart {
		line := it.Cents * int64(it.Qty)
		for _, r := range rules {
			if strings.HasPrefix(it.SKU, r.prefix) {
				line -= line * r.percent / 100
				break
			}
		}
		total += line
	}
	return total, nil
}

type rule struct {
	prefix  string
	percent int64
}

func parseRules(text string) ([]rule, error) {
	var rules []rule
	for n, line := range strings.Split(text, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) != 2 {
			return nil, fmt.Errorf("%w: line %d: %q", ErrBadRule, n+1, line)
		}
		pct, err := strconv.ParseInt(strings.TrimSuffix(fields[1], "%"), 10, 64)
		if err != nil || pct < 0 || pct > 100 {
			return nil, fmt.Errorf("%w: line %d: %q", ErrBadRule, n+1, line)
		}
		rules = append(rules, rule{prefix: fields[0], percent: pct})
	}
	return rules, nil
}

// CurrencySymbol returns the symbol the receipt prints for an ISO 4217 code,
// or the code itself when it has none.
func CurrencySymbol(code string) string {
	symbols := map[string]string{"USD": "$", "EUR": "€", "UAH": "₴", "GBP": "£", "PLN": "zł", "JPY": "¥"}
	if s, ok := symbols[strings.ToUpper(code)]; ok {
		return s
	}
	return code
}
