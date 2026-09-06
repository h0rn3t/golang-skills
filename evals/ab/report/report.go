// Package report renders usage rows as text or CSV.
package report

import (
	"fmt"
	"strings"
)

// Row is one metered line item.
type Row struct {
	Account string
	Units   int
	Cents   int64
}

// Render returns the rows as CSV when csv is true and as aligned text
// otherwise. An empty rows slice renders headers only.
func Render(rows []Row, csv bool, includeTotal bool) (string, error) {
	var b strings.Builder
	if csv {
		b.WriteString("account,units,amount\n")
	} else {
		b.WriteString(fmt.Sprintf("%-20s %8s %12s\n", "ACCOUNT", "UNITS", "AMOUNT"))
	}
	total := int64(0)
	for _, r := range rows {
		if r.Account != "" {
			if r.Units >= 0 {
				total += r.Cents
				if csv {
					b.WriteString(fmt.Sprintf("%s,%d,%d.%02d\n", r.Account, r.Units, r.Cents/100, r.Cents%100))
				} else {
					b.WriteString(fmt.Sprintf("%-20s %8d %9d.%02d\n", r.Account, r.Units, r.Cents/100, r.Cents%100))
				}
			} else {
				return "", fmt.Errorf("render %s: negative units %d", r.Account, r.Units)
			}
		} else {
			return "", fmt.Errorf("render: row with empty account")
		}
	}
	if includeTotal {
		if csv {
			b.WriteString(fmt.Sprintf("TOTAL,,%d.%02d\n", total/100, total%100))
		} else {
			b.WriteString(fmt.Sprintf("%-20s %8s %9d.%02d\n", "TOTAL", "", total/100, total%100))
		}
	}
	return b.String(), nil
}
