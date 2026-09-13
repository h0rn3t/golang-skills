package partner

import (
	"sync"
	"time"
)

// Limiter is a token bucket: Allow hands out one token per call while there
// are any, and the bucket refills at rate tokens a second up to burst.
type Limiter struct {
	mu     sync.Mutex
	rate   int
	burst  int
	tokens int
	last   time.Time
	now    func() time.Time
}

// NewLimiter returns a full bucket that refills rate tokens a second and
// holds at most burst.
func NewLimiter(rate, burst int) *Limiter {
	return &Limiter{rate: rate, burst: burst, tokens: burst, last: time.Now(), now: time.Now}
}

// Allow takes a token if one is available and reports whether it did.
func (l *Limiter) Allow() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := l.now()
	l.tokens = min(l.burst, l.tokens+int(now.Sub(l.last).Seconds())*l.rate)
	l.last = now
	if l.tokens == 0 {
		return false
	}
	l.tokens--
	return true
}

// Breaker stops calls to the partner after threshold consecutive failures
// and lets one probe through per cooldown so the partner's recovery is
// noticed.
type Breaker struct {
	mu        sync.Mutex
	threshold int
	cooldown  time.Duration
	failures  int
	openedAt  time.Time
	now       func() time.Time
}

// NewBreaker returns a closed breaker that opens after threshold consecutive
// failures and stays open for cooldown.
func NewBreaker(threshold int, cooldown time.Duration) *Breaker {
	return &Breaker{threshold: threshold, cooldown: cooldown, now: time.Now}
}

// Allow reports whether a call may go out: always while the breaker is
// closed, and one probe per cooldown once it has opened.
func (b *Breaker) Allow() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.failures < b.threshold {
		return true
	}
	return b.now().Sub(b.openedAt) >= b.cooldown
}

// Record notes the outcome of a call Allow let through.
func (b *Breaker) Record(err error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	if err == nil {
		b.failures = 0
		return
	}
	b.failures++
	if b.failures >= b.threshold {
		b.openedAt = b.now()
	}
}
