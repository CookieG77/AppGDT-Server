package ratelimit

import (
	"testing"
	"time"
)

// newTestLimiter returns a limiter whose clock is controlled by the test.
func newTestLimiter(max int, window time.Duration) (*Limiter, *time.Time) {
	l := New(max, window)
	now := time.Date(2026, 9, 28, 12, 0, 0, 0, time.UTC)
	l.now = func() time.Time { return now }
	return l, &now
}

func TestBlocksAfterMaxFailures(t *testing.T) {
	l, _ := newTestLimiter(3, 15*time.Minute)

	for i := 1; i <= 3; i++ {
		if ok, _ := l.Allow("a@b.c"); !ok {
			t.Fatalf("attempt %d must be allowed", i)
		}
		reached := l.Fail("a@b.c")
		if reached != (i == 3) {
			t.Errorf("Fail #%d reported limit reached = %v", i, reached)
		}
	}

	ok, wait := l.Allow("a@b.c")
	if ok || wait != 15*time.Minute {
		t.Errorf("Allow after the limit = %v, %v; want false, 15m", ok, wait)
	}
	if ok, _ := l.Allow("other@b.c"); !ok {
		t.Error("another key must not be blocked")
	}
}

func TestUnblocksAfterWindow(t *testing.T) {
	l, now := newTestLimiter(2, time.Minute)
	l.Fail("ip")
	l.Fail("ip")

	*now = now.Add(59 * time.Second)
	if ok, wait := l.Allow("ip"); ok || wait != time.Second {
		t.Errorf("still blocked expected, got %v, %v", ok, wait)
	}

	*now = now.Add(time.Second)
	if ok, _ := l.Allow("ip"); !ok {
		t.Error("the key must be allowed again after the window")
	}
}

func TestFailuresOutsideWindowDoNotAddUp(t *testing.T) {
	l, now := newTestLimiter(2, time.Minute)
	l.Fail("ip")
	*now = now.Add(2 * time.Minute)
	if l.Fail("ip") {
		t.Error("a failure in a new window must start a new count")
	}
	if ok, _ := l.Allow("ip"); !ok {
		t.Error("one failure in the window must not block")
	}
}

func TestReset(t *testing.T) {
	l, _ := newTestLimiter(1, time.Minute)
	l.Fail("a@b.c")
	l.Reset("a@b.c")
	if ok, _ := l.Allow("a@b.c"); !ok {
		t.Error("Reset must unblock the key")
	}
}
