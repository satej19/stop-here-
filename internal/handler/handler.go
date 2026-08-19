// Package handler contains HTTP handler functions for the API endpoints.
//
// Handlers are grouped in a Handler struct that receives dependencies
// (like the rate limiter) via dependency injection. This makes handlers
// easy to test and keeps them decoupled from specific implementations.
package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/satej19/ai-gateway/internal/limiter"
)

// Handler holds dependencies needed by the HTTP handler functions.
// Currently, it only needs a RateLimiter, but this struct will grow
// as we add services and other dependencies in later phases.
type Handler struct {
	limiter limiter.RateLimiter
}

// NewHandler creates a Handler with the given rate limiter injected.
//
// Example:
//
//	lim := limiter.NewInMemoryLimiter(10, time.Minute)
//	h := handler.NewHandler(lim)
//	router.GET("/api/data", h.LimitedEndpoint)
func NewHandler(lim limiter.RateLimiter) *Handler {
	return &Handler{
		limiter: lim,
	}
}

// Health handles GET /health.
//
// This is a simple liveness check that always returns 200 OK.
// It is NOT rate-limited — monitoring systems and load balancers
// need to reach this endpoint without restrictions.
//
// Response:
//
//	200 OK  →  {"status": "ok"}
func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "ok",
		"message": "Hello from go server :)",
	})
}

// LimitedEndpoint handles GET /api/data.
//
// This is a sample rate-limited endpoint. On each request, it:
//  1. Extracts the client's IP address as the rate-limit key.
//  2. Calls the rate limiter's Allow() method.
//  3. Sets rate-limit response headers so clients can see their quota.
//  4. Returns 200 with data if allowed, or 429 if rate-limited.
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
// Success response (200 OK):
//
//	{"message": "Here is your data!", "client_ip": "..."}
//
// Rate-limited response (429 Too Many Requests):
//
//	{"error": "rate limit exceeded", "retry_after_seconds": <float>}
func (h *Handler) LimitedEndpoint(c *gin.Context) {
	// Use the client's IP address as the rate-limit key.
	// Gin's ClientIP() respects X-Forwarded-For and X-Real-IP headers,
	// which is important when running behind a load balancer.
	clientIP := c.ClientIP()

	// Ask the limiter: is this client allowed to make a request?
	allowed, remaining, resetAt := h.limiter.Allow(clientIP)

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
		c.JSON(http.StatusTooManyRequests, gin.H{
			"error":               "rate limit exceeded",
			"retry_after_seconds": retryAfter,
		})
		return
	}

	// Request is allowed — return the data.
	c.JSON(http.StatusOK, gin.H{
		"message":   "Here is your data!",
		"client_ip": clientIP,
	})
}
