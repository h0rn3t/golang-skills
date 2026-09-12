package pool

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"testing/synctest"
	"time"
)

var goldenErrBoom = errors.New("partner: connection reset")

// goldenCounter observes what the jobs see: how many are in flight right now,
// the most that ever were, and how many finished.
type goldenCounter struct {
	inFlight, peak, done atomic.Int64
}

func (c *goldenCounter) enter() {
	now := c.inFlight.Add(1)
	for {
		peak := c.peak.Load()
		if now <= peak || c.peak.CompareAndSwap(peak, now) {
			return
		}
	}
}

func (c *goldenCounter) leave() {
	c.inFlight.Add(-1)
	c.done.Add(1)
}

// goldenJob is a job that holds its connection for d, or until ctx ends, and
// then reports result.
func goldenJob(c *goldenCounter, d time.Duration, result error) Job {
	return func(ctx context.Context) error {
		c.enter()
		defer c.leave()
		select {
		case <-time.After(d):
			return result
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

// TestRunNeverExceedsWorkers is the first trap. Starting a goroutine per job
// and waiting is the shortest implementation, and it opens every connection
// at once.
func TestRunNeverExceedsWorkers(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var c goldenCounter
		jobs := make([]Job, 10)
		for i := range jobs {
			jobs[i] = goldenJob(&c, time.Second, nil)
		}

		if err := Run(t.Context(), 3, jobs); err != nil {
			t.Fatalf("Run(3 workers, 10 jobs) error = %v, want nil", err)
		}
		if peak := c.peak.Load(); peak > 3 {
			t.Errorf("Run(3 workers) had %d jobs in flight at once, want at most 3", peak)
		}
		if done := c.done.Load(); done != 10 {
			t.Errorf("Run(3 workers, 10 jobs) finished %d jobs, want 10", done)
		}
	})
}

// TestRunReturnsAfterEveryStartedJob is the second trap. Returning the first
// error as soon as it arrives leaves the other workers running after Run has
// said the run is over — under synctest, a job still holding its connection
// when Run returns has not yet counted itself done.
func TestRunReturnsAfterEveryStartedJob(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var c goldenCounter
		jobs := []Job{
			goldenJob(&c, time.Second, goldenErrBoom),
			goldenJob(&c, time.Hour, nil),
			goldenJob(&c, time.Hour, nil),
		}

		start := time.Now()
		err := Run(t.Context(), 3, jobs)
		if !errors.Is(err, goldenErrBoom) {
			t.Fatalf("Run() error = %v, want %v", err, goldenErrBoom)
		}
		if took := time.Since(start); took >= time.Hour {
			t.Errorf("Run() returned %v after the failure, want the running jobs cancelled instead of waited out", took)
		}
		if inFlight := c.inFlight.Load(); inFlight != 0 {
			t.Errorf("Run() returned with %d job(s) still running, want 0", inFlight)
		}
		if done := c.done.Load(); done != 3 {
			t.Errorf("Run() returned after %d of 3 started jobs finished, want 3", done)
		}
	})
}

// TestRunStopsStartingJobsAfterFailure pins the other half of "the first
// failure ends the run": jobs behind the failure in the queue never start.
func TestRunStopsStartingJobsAfterFailure(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var c goldenCounter
		var started atomic.Int64
		counted := func(j Job) Job {
			return func(ctx context.Context) error {
				started.Add(1)
				return j(ctx)
			}
		}
		jobs := []Job{
			counted(goldenJob(&c, time.Second, goldenErrBoom)),
		}
		for range 20 {
			jobs = append(jobs, counted(goldenJob(&c, time.Hour, nil)))
		}

		err := Run(t.Context(), 1, jobs)
		if !errors.Is(err, goldenErrBoom) {
			t.Fatalf("Run(1 worker) error = %v, want %v", err, goldenErrBoom)
		}
		// One more start is tolerated: a worker slot and the cancellation can
		// become ready in the same instant, and which one a select sees first
		// is not the implementation's to choose.
		if n := started.Load(); n > 2 {
			t.Errorf("Run(1 worker) started %d of 21 jobs after the first failed, want the rest never started", n)
		}
	})
}

func TestRunCancelledContext(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var c goldenCounter
		ctx, cancel := context.WithCancel(t.Context())
		jobs := []Job{goldenJob(&c, time.Hour, nil), goldenJob(&c, time.Hour, nil)}

		var wg sync.WaitGroup
		var err error
		wg.Go(func() { err = Run(ctx, 2, jobs) })
		time.Sleep(time.Minute)
		cancel()
		wg.Wait()

		if !errors.Is(err, context.Canceled) {
			t.Errorf("Run(cancelled ctx) error = %v, want %v", err, context.Canceled)
		}
		if done := c.done.Load(); done != 2 {
			t.Errorf("Run(cancelled ctx) returned after %d of 2 jobs finished, want 2", done)
		}
	})
}

func TestRunAllSucceed(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		var c goldenCounter
		jobs := []Job{goldenJob(&c, time.Second, nil), goldenJob(&c, 2*time.Second, nil)}

		if err := Run(t.Context(), 1, jobs); err != nil {
			t.Errorf("Run() error = %v, want nil", err)
		}
		if done := c.done.Load(); done != 2 {
			t.Errorf("Run() finished %d jobs, want 2", done)
		}
	})
}

func TestRunEdges(t *testing.T) {
	if err := Run(t.Context(), 4, nil); err != nil {
		t.Errorf("Run(no jobs) error = %v, want nil", err)
	}
	for _, workers := range []int{0, -1} {
		err := Run(t.Context(), workers, []Job{func(context.Context) error { return nil }})
		if err == nil {
			t.Errorf("Run(workers = %d) error = nil, want non-nil", workers)
			continue
		}
		if !strings.Contains(err.Error(), "workers") {
			t.Errorf("Run(workers = %d) error = %q, want workers named", workers, err)
		}
	}
}
