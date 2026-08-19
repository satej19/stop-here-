package limiter

import (
	"sync"
	"testing"
	"time"
)

// TestAllowWithinLimit verifies that requests within the rate limit are allowed
// and that the remaining count decreases correctly.
func TestAllowWithinLimit(t *testing.T) {
	lim := NewInMemoryLimiter(3, time.Minute)

	// First request — should be allowed, 2 remaining.
	allowed, remaining, _ := lim.Allow("user1")
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}
	if remaining != 2 {
		t.Fatalf("expected 2 remaining, got %d", remaining)
	}

	// Second request — should be allowed, 1 remaining.
	allowed, remaining, _ = lim.Allow("user1")
	if !allowed {
		t.Fatal("expected second request to be allowed")
	}
	if remaining != 1 {
		t.Fatalf("expected 1 remaining, got %d", remaining)
	}

	// Third request — should be allowed, 0 remaining.
	allowed, remaining, _ = lim.Allow("user1")
	if !allowed {
		t.Fatal("expected third request to be allowed")
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}
}

// TestRejectsOverLimit verifies that requests exceeding the limit are rejected
// and that remaining stays at 0.
func TestRejectsOverLimit(t *testing.T) {
	lim := NewInMemoryLimiter(2, time.Minute)

	// Use up the limit.
	lim.Allow("user1")
	lim.Allow("user1")

	// Third request — should be rejected.
	allowed, remaining, _ := lim.Allow("user1")
	if allowed {
		t.Fatal("expected request to be rejected after exceeding limit")
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining, got %d", remaining)
	}
}

// TestWindowReset verifies that the counter resets after the window expires.
// We use a very short window (50ms) so the test runs quickly.
func TestWindowReset(t *testing.T) {
	lim := NewInMemoryLimiter(1, 50*time.Millisecond)

	// First request — allowed.
	allowed, _, _ := lim.Allow("user1")
	if !allowed {
		t.Fatal("expected first request to be allowed")
	}

	// Second request (same window) — rejected.
	allowed, _, _ = lim.Allow("user1")
	if allowed {
		t.Fatal("expected second request to be rejected within the same window")
	}

	// Wait for the window to expire.
	time.Sleep(60 * time.Millisecond)

	// After the window resets — should be allowed again.
	allowed, remaining, _ := lim.Allow("user1")
	if !allowed {
		t.Fatal("expected request to be allowed after window reset")
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining after using the only allowed request, got %d", remaining)
	}
}

// TestResetAtIsCorrect verifies that the resetAt time is set to the end
// of the current window.
func TestResetAtIsCorrect(t *testing.T) {
	window := 2 * time.Second
	lim := NewInMemoryLimiter(5, window)

	before := time.Now()
	_, _, resetAt := lim.Allow("user1")
	after := time.Now()

	// resetAt should be approximately now + window duration.
	// We allow a small tolerance for execution time.
	expectedEarliest := before.Add(window)
	expectedLatest := after.Add(window)

	if resetAt.Before(expectedEarliest) || resetAt.After(expectedLatest) {
		t.Fatalf("resetAt %v not in expected range [%v, %v]", resetAt, expectedEarliest, expectedLatest)
	}
}

// TestIndependentClients verifies that different clients have independent
// rate-limit counters — one client being rate-limited should not affect another.
func TestIndependentClients(t *testing.T) {
	lim := NewInMemoryLimiter(1, time.Minute)

	// user1 uses their only request.
	allowed, _, _ := lim.Allow("user1")
	if !allowed {
		t.Fatal("expected user1's first request to be allowed")
	}

	// user1 is now rate-limited.
	allowed, _, _ = lim.Allow("user1")
	if allowed {
		t.Fatal("expected user1 to be rate-limited")
	}

	// user2 should still be allowed — they have their own counter.
	allowed, remaining, _ := lim.Allow("user2")
	if !allowed {
		t.Fatal("expected user2's first request to be allowed (independent from user1)")
	}
	if remaining != 0 {
		t.Fatalf("expected 0 remaining for user2 (limit=1), got %d", remaining)
	}
}

// TestConcurrentAccess verifies that the limiter is safe to use from
// multiple goroutines. This is a race-condition detector test — if there's
// a data race, `go test -race` will catch it.
func TestConcurrentAccess(t *testing.T) {
	lim := NewInMemoryLimiter(100, time.Minute)

	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lim.Allow("concurrent-user")
		}()
	}
	wg.Wait()

	// After 50 concurrent requests with a limit of 100, 50 should remain.
	_, remaining, _ := lim.Allow("concurrent-user")
	if remaining != 49 {
		t.Fatalf("expected 49 remaining after 51 requests (50 concurrent + 1 check), got %d", remaining)
	}
}
