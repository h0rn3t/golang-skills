// Package clock is approved shared infrastructure: a time source with no
// business types.
package clock

import "time"

// System reads the wall clock.
type System struct{}

// Now is the current time.
func (System) Now() time.Time { return time.Now() }
