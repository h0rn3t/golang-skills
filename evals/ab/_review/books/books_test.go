package books

import (
	"slices"
	"testing"
	"time"
)

func TestFee(t *testing.T) {
	tests := []struct {
		name        string
		amount, bps int64
		want        int64
	}{
		{name: "whole cents", amount: 10_000, bps: 250, want: 250},
		{name: "rounds half up", amount: 150, bps: 250, want: 4},
		{name: "small refund carries no fee", amount: -30, bps: 250, want: 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Fee(tt.amount, tt.bps); got != tt.want {
				t.Errorf("Fee(%d, %d) = %d, want %d", tt.amount, tt.bps, got, tt.want)
			}
		})
	}
}

func TestParseAccountIDs(t *testing.T) {
	got, err := ParseAccountIDs("7, 42")
	if err != nil {
		t.Fatalf("ParseAccountIDs() error = %v", err)
	}
	if want := []int64{7, 42}; !slices.Equal(got, want) {
		t.Errorf("ParseAccountIDs() = %v, want %v", got, want)
	}
}

func TestDuplicates(t *testing.T) {
	at := time.Date(2026, time.March, 1, 12, 0, 0, 0, time.UTC)
	existing := []Entry{{ID: 1, AccountID: 7, Credit: 500, PostedAt: at}}
	batch := []Entry{
		{ID: 9, AccountID: 7, Credit: 500, PostedAt: at},
		{ID: 10, AccountID: 7, Credit: 600, PostedAt: at},
	}
	if got := Duplicates(existing, batch); len(got) != 1 || got[0].ID != 9 {
		t.Errorf("Duplicates() = %v, want the replayed entry 9", got)
	}
}

func TestNextBillingDate(t *testing.T) {
	got := NextBillingDate(time.Date(2026, time.March, 15, 9, 0, 0, 0, time.UTC))
	if want := time.Date(2026, time.April, 15, 9, 0, 0, 0, time.UTC); !got.Equal(want) {
		t.Errorf("NextBillingDate() = %v, want %v", got, want)
	}
}

func TestStartOfDay(t *testing.T) {
	in := time.Date(2026, time.March, 1, 17, 30, 0, 0, time.UTC)
	want := time.Date(2026, time.March, 1, 0, 0, 0, 0, time.UTC)
	if got := StartOfDay(in, time.UTC); !got.Equal(want) {
		t.Errorf("StartOfDay() = %v, want %v", got, want)
	}
}
