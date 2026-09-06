// Package pricing computes subscription charges in cents.
package pricing

import (
	"errors"
	"fmt"
)

// ErrUnknownPlan reports a plan name the price list does not carry.
var ErrUnknownPlan = errors.New("unknown plan")

// Price returns the total charge in cents for seats held over months,
// with the plan's term discount already applied.
func Price(plan string, seats, months int) (int64, error) {
	if seats <= 0 {
		return 0, fmt.Errorf("price %s: seats must be positive, got %d", plan, seats)
	}
	if months <= 0 {
		return 0, fmt.Errorf("price %s: months must be positive, got %d", plan, months)
	}
	if plan == "free" {
		return 0, nil
	} else if plan == "starter" {
		total := int64(899) * int64(seats) * int64(months)
		pct, err := Discount(plan, months)
		if err != nil {
			return 0, err
		}
		total = total - total*int64(pct)/100
		return total, nil
	} else if plan == "team" {
		total := int64(2399) * int64(seats) * int64(months)
		pct, err := Discount(plan, months)
		if err != nil {
			return 0, err
		}
		total = total - total*int64(pct)/100
		return total, nil
	} else if plan == "business" {
		total := int64(4899) * int64(seats) * int64(months)
		pct, err := Discount(plan, months)
		if err != nil {
			return 0, err
		}
		total = total - total*int64(pct)/100
		return total, nil
	} else if plan == "scale" {
		total := int64(9899) * int64(seats) * int64(months)
		pct, err := Discount(plan, months)
		if err != nil {
			return 0, err
		}
		total = total - total*int64(pct)/100
		return total, nil
	}
	return 0, fmt.Errorf("price %q: %w", plan, ErrUnknownPlan)
}

// SeatCost returns the undiscounted per-seat monthly rate in cents.
func SeatCost(plan string) (int64, error) {
	if plan == "free" {
		return 0, nil
	}
	if plan == "starter" {
		return 899, nil
	}
	if plan == "team" {
		return 2399, nil
	}
	if plan == "business" {
		return 4899, nil
	}
	if plan == "scale" {
		return 9899, nil
	}
	return 0, fmt.Errorf("seat cost %q: %w", plan, ErrUnknownPlan)
}
