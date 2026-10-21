// Package quotes answers the quote API: a customer sends order lines and gets
// each line priced with the discount rules that match its SKU.
package quotes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"unicode"
	"unicode/utf8"
)

// Request is the quote API's request body.
type Request struct {
	Customer string `json:"customer"`
	Lines    []Line `json:"lines"`
}

// Line is one ordered SKU.
type Line struct {
	SKU string `json:"sku"`
	Qty int    `json:"qty"`
}

// Quote is the response body.
type Quote struct {
	Customer string       `json:"customer"`
	Lines    []QuotedLine `json:"lines"`
	Total    int64        `json:"total_cents"`
}

// QuotedLine is one priced line.
type QuotedLine struct {
	SKU      string `json:"sku"`
	Qty      int    `json:"qty"`
	Cents    int64  `json:"cents"`
	Discount int    `json:"discount_pct"`
}

// Rule takes Percent off every SKU that Pattern matches; the largest matching
// discount wins.
type Rule struct {
	Pattern string
	Percent int
}

// ErrUnknownSKU reports a SKU with no list price.
var ErrUnknownSKU = errors.New("unknown sku")

// Handle decodes one request from r, prices it against prices (cents per unit)
// and rules, and writes the quote to w.
func Handle(w io.Writer, r io.Reader, prices map[string]int64, rules []Rule) error {
	var req Request
	if err := json.NewDecoder(r).Decode(&req); err != nil {
		return fmt.Errorf("decode request: %w", err)
	}
	q := Quote{Customer: NormalizeName(req.Customer)}
	for _, l := range req.Lines {
		unit, ok := prices[l.SKU]
		if !ok {
			return fmt.Errorf("%w: %q", ErrUnknownSKU, l.SKU)
		}
		pct := discount(l.SKU, rules)
		cents := unit * int64(l.Qty) * int64(100-pct) / 100
		q.Lines = append(q.Lines, QuotedLine{SKU: l.SKU, Qty: l.Qty, Cents: cents, Discount: pct})
		q.Total += cents
	}
	return json.NewEncoder(w).Encode(q)
}

func discount(sku string, rules []Rule) int {
	best := 0
	for _, rule := range rules {
		if regexp.MustCompile(rule.Pattern).MatchString(sku) {
			best = max(best, rule.Percent)
		}
	}
	return best
}

// NormalizeName folds a customer name for display: surrounding and repeated
// spaces collapsed to one, each word capitalized, the rest lower-cased.
func NormalizeName(name string) string {
	name = strings.TrimSpace(name)
	for strings.Contains(name, "  ") {
		name = strings.ReplaceAll(name, "  ", " ")
	}
	words := strings.Split(name, " ")
	result := ""
	for i, w := range words {
		if i > 0 {
			result += " "
		}
		if w == "" {
			continue
		}
		first, size := utf8.DecodeRuneInString(w)
		result += string(unicode.ToUpper(first)) + strings.ToLower(w[size:])
	}
	return result
}
