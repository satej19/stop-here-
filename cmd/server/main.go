package main

import (
	"log"

	"github.com/satej19/ai-gateway/internal/config"
	"github.com/satej19/ai-gateway/internal/handler"
	"github.com/satej19/ai-gateway/internal/limiter"
	"github.com/satej19/ai-gateway/internal/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration from environment variables (with defaults).
	cfg := config.Load()

	// Create the in-memory rate limiter using config values.
	lim := limiter.NewInMemoryLimiter(cfg.RateLimit, cfg.RateWindow)

	// Create the handler (no longer depends on the limiter).
	h := handler.NewHandler()

	// Set up the Gin router.
	r := gin.Default()

	// Health check endpoint — not rate-limited.
	r.GET("/health", h.Health)

	// API group — all routes here are rate-limited via middleware.
	api := r.Group("/api")
	api.Use(middleware.RateLimiter(lim))
	{
		api.GET("/data", h.Data)
	}

	// Start the server.
	log.Printf("Server starting on :%s", cfg.Port)
	r.Run(":" + cfg.Port)
}
