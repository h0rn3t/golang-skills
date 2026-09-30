package degrade

import (
	"fmt"
	"log"
)

// Report is the "log and degrade gracefully" branch of go-error-handling's
// Error Flow, followed
// within a few lines by a return of a different error: not log-and-return.
func Report() (int, error) {
	if err := emitMetrics(); err != nil {
		log.Printf("emit metrics: %v", err)
	}
	v, err := load()
	if err != nil {
		return 0, fmt.Errorf("load: %w", err)
	}
	return v, nil
}

// Emit ends with a log; Load, the next function, returns err two lines later.
func Emit() {
	if err := emitMetrics(); err != nil {
		log.Printf("emit metrics: %v", err)
	}
}
func Load() error { _, err := load(); return err }

// Retry logs the first failure and returns the retry's error, a new one.
func Retry() error {
	if err := emitMetrics(); err != nil {
		log.Printf("emit metrics, retrying: %v", err)
		err = emitMetrics()
		return err
	}
	return nil
}

func emitMetrics() error { return nil }
func load() (int, error) { return 0, nil }
