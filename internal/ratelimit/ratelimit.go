// Package ratelimit counts failed attempts per key (a client IP, an email...)
// and blocks a key once it reaches a limit.
//
// Counters are kept in memory: they are lost when the server restarts and
// are not shared between several instances of the server. This is enough for
// a single server; several instances would need a shared store instead (the
// database, Redis...).
package ratelimit

import (
	"sync"
	"time"
)

// Limiter allows at most max failures per key within a window. The window
// starts at the first failure. Once the limit is reached, the key is blocked
// for a full window starting from its last failure, so that failures spread
// over the end of one window and the start of the next cannot exceed it.
//
// A Limiter is safe for concurrent use.
type Limiter struct {
	mu        sync.Mutex
	max       int
	window    time.Duration
	counters  map[string]*counter
	lastSweep time.Time

	// now returns the current time. It can be replaced in tests to control
	// the passing of time.
	now func() time.Time
}

type counter struct {
	failures int
	resetAt  time.Time
}

// New creates a Limiter allowing max failures per key within window.
func New(max int, window time.Duration) *Limiter {
	return &Limiter{
		max:      max,
		window:   window,
		counters: make(map[string]*counter),
		now:      time.Now,
	}
}

// Allow reports whether the key may attempt again. If it may not, it also
// returns how long remains before the key is unblocked.
func (l *Limiter) Allow(key string) (bool, time.Duration) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	c, ok := l.counters[key]
	if !ok || !now.Before(c.resetAt) || c.failures < l.max {
		return true, 0
	}
	return false, c.resetAt.Sub(now)
}

// Fail records a failed attempt for the key. It returns true when this
// failure makes the key reach the limit, i.e. when the key becomes blocked.
func (l *Limiter) Fail(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := l.now()
	l.sweep(now)

	c, ok := l.counters[key]
	if !ok || !now.Before(c.resetAt) {
		c = &counter{resetAt: now.Add(l.window)}
		l.counters[key] = c
	}

	c.failures++
	if c.failures >= l.max {
		// Block for a full window from the last failure.
		c.resetAt = now.Add(l.window)
	}
	return c.failures == l.max
}

// Reset forgets the failures of the key, for instance after a successful attempt.
func (l *Limiter) Reset(key string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	delete(l.counters, key)
}

// sweep removes the expired counters, at most once per window, so that the
// memory used stays proportional to the keys seen during the last window.
// It must be called with the mutex held.
func (l *Limiter) sweep(now time.Time) {
	if now.Sub(l.lastSweep) < l.window {
		return
	}
	for key, c := range l.counters {
		if !now.Before(c.resetAt) {
			delete(l.counters, key)
		}
	}
	l.lastSweep = now
}
