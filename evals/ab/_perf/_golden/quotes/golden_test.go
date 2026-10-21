package quotes

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestGoldenHandle(t *testing.T) {
	prices := map[string]int64{"SKU-1001": 250, "SKU-2002": 1999, "BOX-7": 40}
	rules := []Rule{{Pattern: "^SKU-1", Percent: 10}, {Pattern: "^SKU-[0-9]+$", Percent: 5}, {Pattern: "7$", Percent: 50}}
	body := `{"customer":"  ada   LOVELACE ","lines":[{"sku":"SKU-1001","qty":3},{"sku":"SKU-2002","qty":1},{"sku":"BOX-7","qty":2}]}`
	var out bytes.Buffer
	if err := Handle(&out, strings.NewReader(body), prices, rules); err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	want := `{"customer":"Ada Lovelace","lines":[{"sku":"SKU-1001","qty":3,"cents":675,"discount_pct":10},{"sku":"SKU-2002","qty":1,"cents":1899,"discount_pct":5},{"sku":"BOX-7","qty":2,"cents":40,"discount_pct":50}],"total_cents":2614}` + "\n"
	if got := out.String(); got != want {
		t.Errorf("Handle() wrote\n%s\nwant\n%s", got, want)
	}

	err := Handle(&out, strings.NewReader(`{"customer":"x","lines":[{"sku":"NOPE","qty":1}]}`), prices, rules)
	if !errors.Is(err, ErrUnknownSKU) {
		t.Errorf("Handle(unknown sku) error = %v, want %v", err, ErrUnknownSKU)
	}
	if err := Handle(&out, strings.NewReader(`{`), prices, rules); err == nil {
		t.Error("Handle(truncated body) error = nil, want a decode error")
	}
}

func TestGoldenNormalizeName(t *testing.T) {
	for in, want := range map[string]string{
		"  maria   de la  CRUZ ": "Maria De La Cruz",
		"élodie":                 "Élodie",
		"":                       "",
		"x":                      "X",
	} {
		if got := NormalizeName(in); got != want {
			t.Errorf("NormalizeName(%q) = %q, want %q", in, got, want)
		}
	}
}
