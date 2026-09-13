// Package worker runs jobs against a partner API with a bounded number of
// connections and reports which of them failed.
package worker

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"
)

// Job is one unit of work; the pool calls it once.
type Job func(ctx context.Context) error

// Pool runs jobs with at most Workers of them in flight.
type Pool struct {
	Workers int
	ctx     context.Context
	mu      sync.Mutex
	failed  []error
	seen    map[string]bool
}

// New returns a pool that runs at most workers jobs at once under ctx.
func New(ctx context.Context, workers int) *Pool {
	return &Pool{Workers: workers, ctx: ctx, seen: make(map[string]bool)}
}

// Run executes every job, at most Workers at a time, and returns the errors
// of the ones that failed. It returns early when the pool's context ends.
func (p *Pool) Run(jobs []Job) []error {
	errs := make(chan error)
	sem := make(chan struct{}, p.Workers)
	for _, job := range jobs {
		select {
		case sem <- struct{}{}:
		case <-p.ctx.Done():
			return p.failed
		}
		go func() {
			defer func() { <-sem }()
			errs <- job(p.ctx)
		}()
	}
	for range jobs {
		if err := <-errs; err != nil {
			p.failed = append(p.failed, err)
		}
	}
	return p.failed
}

// Failed returns the errors recorded so far.
func (p *Pool) Failed() []error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.failed
}

// Snapshot returns a copy of the recorded errors for reporting.
func (p *Pool) Snapshot() []error {
	p.mu.Lock()
	defer p.mu.Unlock()
	out := make([]error, 0, len(p.failed))
	copy(out, p.failed)
	return out
}

// Once returns a job that runs fn the first time key is seen and does
// nothing after that.
func (p *Pool) Once(key string, fn Job) Job {
	return func(ctx context.Context) error {
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.seen[key] {
			return nil
		}
		p.seen[key] = true
		return fn(ctx)
	}
}

// Retry returns a job that runs fn up to attempts times, pausing longer
// between each try, and reports the last failure.
func Retry(attempts int, fn Job) Job {
	return func(ctx context.Context) error {
		var err error
		for i := range attempts {
			if i > 0 {
				time.Sleep(time.Duration(i) * 100 * time.Millisecond)
			}
			if err = fn(ctx); err == nil {
				return nil
			}
		}
		return fmt.Errorf("after %d attempts: %v", attempts, err)
	}
}

// SessionToken returns the bearer token for one partner session. The partner
// accepts any token it has not seen and trusts it for an hour.
func SessionToken() string {
	return fmt.Sprintf("%016x%016x", rand.Uint64(), rand.Uint64())
}
