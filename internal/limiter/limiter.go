// Package limiter contains the core rate limiting algorithms and logic.
//
// The package defines a RateLimiter interface so the underlying algorithm
// can be swapped without changing the rest of the application. Currently,
// only an in-memory fixed-window-counter implementation is provided.
// Later phases will add a token-bucket algorithm (Phase 4) and a
// Redis-backed implementation (Phase 5).
package limiter

import (
	"sync"
	"time"
)

// RateLimiter is the interface that any rate limiting implementation must satisfy.
// It provides a single method, Allow, which decides whether a request from the
// given client key should be permitted.
//
// Return values:
//   - allowed:   true if the request is permitted, false if rate-limited.
//   - remaining: number of requests the client can still make in the current window.
//   - resetAt:   the time when the current rate-limit window expires and the
//     counter resets.
type RateLimiter interface {
	Allow(key string) (allowed bool, remaining int, resetAt time.Time)
}

// clientWindow tracks rate-limit state for a single client (identified by key).
// Each client gets their own window that resets independently.
type clientWindow struct {
	count       int       // number of requests made in the current window
	windowStart time.Time // when the current window began
}

// InMemoryLimiter implements RateLimiter using an in-memory fixed-window counter.
//
// How it works:
//  1. Each client key maps to a clientWindow (count + windowStart).
//  2. When a request arrives, if the window has expired, reset count to 0
//     and start a new window.
//  3. If the count is below maxRequests, allow the request and increment.
//  4. Otherwise, reject the request.
//
// Thread safety: All map access is protected by a sync.Mutex.
// This is sufficient for a single-server deployment. For multiple servers,
// use a Redis-backed limiter (Phase 5).
type InMemoryLimiter struct {
	maxRequests int                      // maximum requests allowed per window
	window      time.Duration            // duration of each rate-limit window
	mu          sync.Mutex               // protects the clients map
	clients     map[string]*clientWindow // per-client rate-limit state
}

// NewInMemoryLimiter creates a new in-memory rate limiter.
//
// Parameters:
//   - maxRequests: maximum number of requests a client can make per window.
//   - window:      how long each rate-limit window lasts before resetting.
//
// Example:
//
//	lim := limiter.NewInMemoryLimiter(10, time.Minute)
//	// Allows 10 requests per minute per client.
func NewInMemoryLimiter(maxRequests int, window time.Duration) *InMemoryLimiter {
	return &InMemoryLimiter{
		maxRequests: maxRequests,
		window:      window,
		clients:     make(map[string]*clientWindow),
	}
}

// Allow checks whether a request from the given key is permitted.
//
// The key is typically a client IP address, but can be any unique identifier
// (API key, user ID, etc.).
//
// If the client has not been seen before, a new window is created.
// If the current window has expired, it is reset automatically.
//
// Returns:
//   - allowed:   true if the request is within the rate limit.
//   - remaining: how many more requests the client can make in this window.
//   - resetAt:   when the current window expires (useful for Retry-After headers).
func (l *InMemoryLimiter) Allow(key string) (bool, int, time.Time) {
	l.mu.Lock()
	defer l.mu.Unlock()

	now := time.Now()

	// Get or create the client's window.
	cw, exists := l.clients[key]
	if !exists {
		// First request from this client — create a new window.
		cw = &clientWindow{
			count:       0,
			windowStart: now,
		}
		l.clients[key] = cw
	}

	// If the window has expired, reset the counter and start a new window.
	windowEnd := cw.windowStart.Add(l.window)
	if now.After(windowEnd) {
		cw.count = 0
		cw.windowStart = now
		windowEnd = now.Add(l.window)
	}

	// Check if the client is within their rate limit.
	if cw.count < l.maxRequests {
		cw.count++
		remaining := l.maxRequests - cw.count
		return true, remaining, windowEnd
	}

	// Rate limit exceeded.
	return false, 0, windowEnd
}
