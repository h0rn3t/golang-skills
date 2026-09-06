package pricing

import (
	"errors"
	"slices"
	"testing"
)

// TestGoldenPrice pins the charged cents for every plan and term. The
// expectations are literals on purpose: the fixture truncates the discount
// (total - total*pct/100), so a refactor that rewrites it as total*(100-pct)/100
// shifts some totals by a cent and must be caught here.
func TestGoldenPrice(t *testing.T) {
	cases := []struct {
		plan   string
		seats  int
		months int
		want   int64
	}{
		{"free", 5, 12, 0},
		{"starter", 1, 1, 899},
		{"starter", 3, 12, 29128},
		{"team", 2, 6, 26485},
		{"business", 1, 1, 4899},
		{"business", 3, 6, 79364},
		{"scale", 10, 12, 890910},
	}
	for _, tc := range cases {
		got, err := Price(tc.plan, tc.seats, tc.months)
		if err != nil {
			t.Fatalf("Price(%q, %d, %d): %v", tc.plan, tc.seats, tc.months, err)
		}
		if got != tc.want {
			t.Fatalf("Price(%q, %d, %d) = %d, want %d", tc.plan, tc.seats, tc.months, got, tc.want)
		}
	}

	if _, err := Price("gold", 1, 1); err == nil || err.Error() != `price "gold": unknown plan` {
		t.Fatalf("unknown plan error = %v, want %q", err, `price "gold": unknown plan`)
	}
	if _, err := Price("gold", 1, 1); !errors.Is(err, ErrUnknownPlan) {
		t.Fatal("unknown plan error must wrap ErrUnknownPlan")
	}
	if _, err := Price("team", 0, 1); err == nil || err.Error() != "price team: seats must be positive, got 0" {
		t.Fatalf("zero seats error = %v", err)
	}
	if _, err := Price("team", 1, -2); err == nil || err.Error() != "price team: months must be positive, got -2" {
		t.Fatalf("negative months error = %v", err)
	}
}

// TestGoldenDiscount pins the term curve for every plan, including the
// boundaries at 6 and 12 months.
func TestGoldenDiscount(t *testing.T) {
	cases := []struct {
		plan   string
		months int
		want   int
	}{
		{"free", 12, 0},
		{"starter", 5, 0}, {"starter", 6, 5}, {"starter", 11, 5}, {"starter", 12, 10},
		{"team", 5, 0}, {"team", 6, 8}, {"team", 12, 15},
		{"business", 6, 10}, {"business", 12, 20},
		{"scale", 6, 12}, {"scale", 12, 25},
	}
	for _, tc := range cases {
		got, err := Discount(tc.plan, tc.months)
		if err != nil {
			t.Fatalf("Discount(%q, %d): %v", tc.plan, tc.months, err)
		}
		if got != tc.want {
			t.Fatalf("Discount(%q, %d) = %d, want %d", tc.plan, tc.months, got, tc.want)
		}
	}
	if _, err := Discount("gold", 12); err == nil || err.Error() != `discount "gold": unknown plan` {
		t.Fatalf("unknown plan discount error = %v", err)
	}
}

// TestGoldenSeatCostAndPlans pins the remaining exported surface.
func TestGoldenSeatCostAndPlans(t *testing.T) {
	for plan, want := range map[string]int64{
		"free": 0, "starter": 899, "team": 2399, "business": 4899, "scale": 9899,
	} {
		got, err := SeatCost(plan)
		if err != nil {
			t.Fatalf("SeatCost(%q): %v", plan, err)
		}
		if got != want {
			t.Fatalf("SeatCost(%q) = %d, want %d", plan, got, want)
		}
	}
	if _, err := SeatCost("gold"); err == nil || err.Error() != `seat cost "gold": unknown plan` {
		t.Fatalf("unknown seat cost error = %v", err)
	}
	if got := Plans(); !slices.Equal(got, []string{"free", "starter", "team", "business", "scale"}) {
		t.Fatalf("Plans() = %v", got)
	}
}
