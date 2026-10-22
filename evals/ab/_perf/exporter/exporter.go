// Package exporter writes the nightly order export the finance team imports.
package exporter

import (
	"fmt"
	"io"
	"regexp"
	"slices"
	"strings"
)

// Row is one order line.
type Row struct {
	OrderID  int64
	Customer string
	SKU      string
	Qty      int
	Cents    int64
}

// columns are the export's header fields, in order.
var columns = []string{"order", "customer", "sku", "qty", "amount"}

// Export writes a header line, one line per row, and a trailer line naming
// every customer once, in the order each first appears. Fields are separated
// by commas; amount is in currency units with two decimals. It stops at the
// first write error and returns it.
func Export(w io.Writer, rows []Row) error {
	if _, err := io.WriteString(w, header(columns)); err != nil {
		return err
	}
	for _, r := range rows {
		if _, err := io.WriteString(w, formatRow(r)); err != nil {
			return err
		}
	}
	_, err := io.WriteString(w, "customers:"+strings.Join(customers(rows), ";")+"\n")
	return err
}

// header renders the column names with their first letter capitalized.
func header(cols []string) string {
	first := regexp.MustCompile(`^[a-z]`)
	var b strings.Builder
	for i, c := range cols {
		if i > 0 {
			b.WriteByte(',')
		}
		b.WriteString(first.ReplaceAllStringFunc(c, strings.ToUpper))
	}
	b.WriteByte('\n')
	return b.String()
}

func formatRow(r Row) string {
	return fmt.Sprintf("%d,%s,%s,%d,%s\n", r.OrderID, r.Customer, r.SKU, r.Qty, formatCents(r.Cents))
}

func formatCents(c int64) string {
	sign := ""
	if c < 0 {
		sign, c = "-", -c
	}
	return fmt.Sprintf("%s%d.%02d", sign, c/100, c%100)
}

// customers lists each customer once, in order of first appearance.
func customers(rows []Row) []string {
	var out []string
	for _, r := range rows {
		out = addUnique(out, r.Customer)
	}
	return out
}

func addUnique(list []string, s string) []string {
	if slices.Contains(list, s) {
		return list
	}
	return append(list, s)
}
