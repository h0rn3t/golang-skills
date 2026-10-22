package exporter

import (
	"bytes"
	"errors"
	"testing"
)

func TestGoldenExport(t *testing.T) {
	rows := []Row{
		{OrderID: 7, Customer: "acme", SKU: "A-1", Qty: 2, Cents: 1999},
		{OrderID: 8, Customer: "Zeta GmbH", SKU: "B-2", Qty: 1, Cents: -5},
		{OrderID: 9, Customer: "acme", SKU: "C-3", Qty: 10, Cents: 100000},
		{OrderID: 10, Customer: "", SKU: "D-4", Qty: 0, Cents: 0},
		{OrderID: 11, Customer: "Zeta GmbH", SKU: "A-1", Qty: 3, Cents: -12345},
		{OrderID: 12, Customer: "beta", SKU: "E-5", Qty: 1, Cents: 7},
	}
	var buf bytes.Buffer
	if err := Export(&buf, rows); err != nil {
		t.Fatalf("Export() error = %v", err)
	}
	want := "Order,Customer,Sku,Qty,Amount\n" +
		"7,acme,A-1,2,19.99\n" +
		"8,Zeta GmbH,B-2,1,-0.05\n" +
		"9,acme,C-3,10,1000.00\n" +
		"10,,D-4,0,0.00\n" +
		"11,Zeta GmbH,A-1,3,-123.45\n" +
		"12,beta,E-5,1,0.07\n" +
		"customers:acme;Zeta GmbH;;beta\n"
	if got := buf.String(); got != want {
		t.Errorf("Export() wrote\n%q\nwant\n%q", got, want)
	}

	buf.Reset()
	if err := Export(&buf, nil); err != nil {
		t.Fatalf("Export(nil) error = %v", err)
	}
	if got, want := buf.String(), "Order,Customer,Sku,Qty,Amount\ncustomers:\n"; got != want {
		t.Errorf("Export(nil) wrote %q, want %q", got, want)
	}
}

var errGoldenWrite = errors.New("disk full")

// goldenLimitWriter accepts n writes, then fails every later one.
type goldenLimitWriter struct{ n int }

func (w *goldenLimitWriter) Write(p []byte) (int, error) {
	if w.n == 0 {
		return 0, errGoldenWrite
	}
	w.n--
	return len(p), nil
}

func TestGoldenExportError(t *testing.T) {
	rows := []Row{{OrderID: 1, Customer: "a", SKU: "x", Qty: 1, Cents: 1}}
	for n := range 3 {
		if err := Export(&goldenLimitWriter{n: n}, rows); !errors.Is(err, errGoldenWrite) {
			t.Errorf("Export(writer failing after %d writes) = %v, want %v", n, err, errGoldenWrite)
		}
	}
}
