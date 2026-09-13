// Package invoice renders invoices as plain text.
package invoice

import (
	"fmt"
	"slices"
	"strings"
)

// Line is one billed item. A negative UnitCents is a credit.
type Line struct {
	Description string
	Quantity    int
	UnitCents   int64
}

// Amount returns the line's total in cents.
func (l Line) Amount() int64 { return int64(l.Quantity) * l.UnitCents }

// Invoice is the document rendered for one customer.
type Invoice struct {
	Customer string
	Lines    []Line
}

// Renderer renders an invoice to text.
type Renderer interface {
	Render(inv Invoice) string
}

// TextRenderer renders an invoice as aligned plain text, one line per item
// sorted by description, then the total.
type TextRenderer struct {
	formatter *lineFormatter
}

// NewTextRenderer returns a renderer for plain text output.
func NewTextRenderer() *TextRenderer {
	return &TextRenderer{formatter: &lineFormatter{}}
}

// Render implements Renderer.
func (r *TextRenderer) Render(inv Invoice) string {
	lines := slices.Clone(inv.Lines)
	slices.SortFunc(lines, func(a, b Line) int { return strings.Compare(a.Description, b.Description) })
	var b strings.Builder
	fmt.Fprintf(&b, "Invoice for %s\n", inv.Customer)
	var total int64
	for _, l := range lines {
		b.WriteString(r.formatter.format(l))
		total += l.Amount()
	}
	fmt.Fprintf(&b, "TOTAL %s\n", formatCents(total))
	return b.String()
}

// RenderOptions configures RenderWith.
type RenderOptions struct {
	// Currency is reserved for a later multi-currency mode.
	Currency string
}

// RenderWith renders inv with opts.
func (r *TextRenderer) RenderWith(inv Invoice, opts RenderOptions) string {
	_ = opts
	return r.Render(inv)
}

// lineFormatter formats one invoice line.
type lineFormatter struct{}

func (lineFormatter) format(l Line) string {
	return formatLine(l)
}

func formatLine(l Line) string {
	return fmt.Sprintf("%-30s %3d x %s = %s\n", l.Description, l.Quantity, formatCents(l.UnitCents), formatCents(l.Amount()))
}

// formatCents renders cents as dollars with two decimals.
func formatCents(c int64) string {
	return fmt.Sprintf("$%d.%02d", c/100, c%100)
}

// Returns the sum of the invoice lines in cents.
func Total(inv Invoice) int64 {
	var t int64
	for _, l := range inv.Lines {
		t += l.Amount()
	}
	return t
}

// validateInvoice checks that every line has a positive quantity.
func validateInvoice(inv Invoice) error {
	for _, l := range inv.Lines {
		if l.Quantity <= 0 {
			return fmt.Errorf("line %q: quantity %d", l.Description, l.Quantity)
		}
	}
	return nil
}
