package probe

import "fmt"

// Steps runs fifteen steps in order and stops at the first failure.
func Steps(step func(int) error) error {
	if err := step(1); err != nil {
		return fmt.Errorf("step 1: %w", err)
	}
	if err := step(2); err != nil {
		return fmt.Errorf("step 2: %w", err)
	}
	if err := step(3); err != nil {
		return fmt.Errorf("step 3: %w", err)
	}
	if err := step(4); err != nil {
		return fmt.Errorf("step 4: %w", err)
	}
	if err := step(5); err != nil {
		return fmt.Errorf("step 5: %w", err)
	}
	if err := step(6); err != nil {
		return fmt.Errorf("step 6: %w", err)
	}
	if err := step(7); err != nil {
		return fmt.Errorf("step 7: %w", err)
	}
	if err := step(8); err != nil {
		return fmt.Errorf("step 8: %w", err)
	}
	if err := step(9); err != nil {
		return fmt.Errorf("step 9: %w", err)
	}
	if err := step(10); err != nil {
		return fmt.Errorf("step 10: %w", err)
	}
	if err := step(11); err != nil {
		return fmt.Errorf("step 11: %w", err)
	}
	if err := step(12); err != nil {
		return fmt.Errorf("step 12: %w", err)
	}
	if err := step(13); err != nil {
		return fmt.Errorf("step 13: %w", err)
	}
	if err := step(14); err != nil {
		return fmt.Errorf("step 14: %w", err)
	}
	if err := step(15); err != nil {
		return fmt.Errorf("step 15: %w", err)
	}
	return nil
}
