// Package spool renders report rows as fixed-precision text lines.
package spool

import (
	"fmt"
	"io"
	"strconv"
)

// Row is one line of a report: a name followed by its values.
type Row struct {
	Name   string
	Values []float64
	// Digits overrides the report's precision for this row when it is
	// above zero.
	Digits int
}

type flusher interface {
	fl() error
}

// Render writes one line per row to w: the row's name, then each value with
// d digits after the decimal point (or the row's own Digits when it sets
// them), separated by single spaces, then " = " and the sum of the values,
// always with d digits. When w came from NewBuffer, Render flushes it before
// returning.
func Render(w io.Writer, rows []Row, d int) error {
	if d < 0 {
		return fmt.Errorf("spool: negative precision %d", d)
	} else {
		for i := 0; i < len(rows); i++ {
			r := rows[i]
			if r.Name == "" {
				return fmt.Errorf("spool: row %d has no name", i)
			}
			var line string
			line = r.Name
			digits := d
			if r.Digits > 0 {
				digits = r.Digits
			}
			var total float64
			total = 0
			for j := 0; j < len(r.Values); j++ {
				var s string
				s = strconv.FormatFloat(r.Values[j], 'f', digits, 64)
				line = line + " " + s
				total = total + r.Values[j]
			}
			line = line + " = " + strconv.FormatFloat(total, 'f', d, 64)
			line = line + "\n"
			_, err := io.WriteString(w, line)
			if err != nil {
				return fmt.Errorf("spool: write row %q: %w", r.Name, err)
			}
		}
	}
	f, ok := w.(flusher)
	if ok == true {
		err := f.fl()
		if err != nil {
			return fmt.Errorf("spool: flush: %w", err)
		}
		return nil
	}
	return nil
}
