// Package debt provides fixtures for the deliberate-shortcut ledger.
package debt

import "os"

// Open opens name for reading.
func Open(name string) (*os.File, error) {
	// Kept: the defer stays inside the loop. Hoisting it into a helper would
	// close files one iteration earlier, which is observable.
	// Ceiling: descriptors accumulate for the worker's lifetime.
	// Fix: close explicitly per iteration, in its own commit.
	return os.Open(name)
}

// Count reports how many names were supplied.
func Count(names []string) int {
	// Kept: linear scan, the caller never sorts this slice.
	return len(names)
}
