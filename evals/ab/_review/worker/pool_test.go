package worker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestRunCollectsFailures(t *testing.T) {
	p := New(context.Background(), 2)
	boom := errors.New("boom")
	jobs := []Job{
		func(context.Context) error { return nil },
		func(context.Context) error { return boom },
	}
	got := p.Run(jobs)
	time.Sleep(50 * time.Millisecond)
	if len(got) != 1 || !errors.Is(got[0], boom) {
		t.Fatalf("Run() = %v, want [boom]", got)
	}
}

func TestRetryGivesUp(t *testing.T) {
	calls := 0
	job := Retry(3, func(context.Context) error {
		calls++
		return errors.New("down")
	})
	if err := job(t.Context()); err == nil {
		t.Fatal("Retry() = nil, want an error after every attempt failed")
	}
	if calls != 3 {
		t.Fatalf("fn called %d times, want 3", calls)
	}
}
