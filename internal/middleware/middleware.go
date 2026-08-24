// Package middleware contains HTTP middleware (e.g., rate limiting, logging).
//
// Middleware functions run BEFORE the actual handler. They can inspect
// or modify the request, decide whether to allow it through, and add
// response headers. This keeps cross-cutting concerns (rate limiting,
// auth, logging) separate from business logic.
package middleware

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satej19/ai-gateway/internal/limiter"
)

// RateLimiter creates a Gin middleware that enforces rate limits.
//
// How it works:
//  1. Extracts the client's IP address as the rate-limit key.
//  2. Calls limiter.Allow(key) to check if the request is permitted.
//  3. Sets rate-limit response headers on every response.
//  4. If allowed → calls c.Next() to pass the request to the handler.
//  5. If rejected → aborts with 429 Too Many Requests.
//
// Response headers (always set):
//
//	X-RateLimit-Remaining: <int>   — requests left in this window
//	X-RateLimit-Reset:     <unix>  — when the window resets (Unix timestamp)
//
// Additional header on 429:
//
//	Retry-After: <seconds>         — seconds until the client can retry
//
// Usage:
//
//	lim := limiter.NewInMemoryLimiter(10, time.Minute)
//	router.Use(middleware.RateLimiter(lim))
func RateLimiter(lim limiter.RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use the client's IP address as the rate-limit key.
		// Gin's ClientIP() respects X-Forwarded-For and X-Real-IP headers,
		// which is important when running behind a load balancer.
		key := c.ClientIP()

		// Ask the limiter: is this client allowed to make a request?
		allowed, remaining, resetAt := lim.Allow(key)

		// Always set rate-limit headers so clients can monitor their usage.
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remaining))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

		if !allowed {
			// Calculate how many seconds until the window resets.
			retryAfter := time.Until(resetAt).Seconds()
			if retryAfter < 0 {
				retryAfter = 0
			}

			c.Header("Retry-After", fmt.Sprintf("%.0f", retryAfter))
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error":              "rate limit exceeded",
				"retry_after_seconds": retryAfter,
			})
			return
		}

		// Request is within the limit — pass it to the next handler.
		c.Next()
	}
}
