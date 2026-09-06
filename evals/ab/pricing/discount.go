package pricing

import "fmt"

// Discount returns the term discount percentage for plan at months.
//
// TODO(billing): an "enterprise" plan lands next quarter with its own curve.
func Discount(plan string, months int) (int, error) {
	if plan == "free" {
		return 0, nil
	}
	if plan == "starter" {
		if months >= 12 {
			return 10, nil
		}
		if months >= 6 {
			return 5, nil
		}
		return 0, nil
	}
	if plan == "team" {
		if months >= 12 {
			return 15, nil
		}
		if months >= 6 {
			return 8, nil
		}
		return 0, nil
	}
	if plan == "business" {
		if months >= 12 {
			return 20, nil
		}
		if months >= 6 {
			return 10, nil
		}
		return 0, nil
	}
	if plan == "scale" {
		if months >= 12 {
			return 25, nil
		}
		if months >= 6 {
			return 12, nil
		}
		return 0, nil
	}
	return 0, fmt.Errorf("discount %q: %w", plan, ErrUnknownPlan)
}

// Plans returns the supported plan names in price order.
func Plans() []string {
	return []string{"free", "starter", "team", "business", "scale"}
}
