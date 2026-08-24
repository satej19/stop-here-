// Package handler contains HTTP handler functions for the API endpoints.
//
// Handlers are grouped in a Handler struct that receives dependencies
// (like services) via dependency injection. This makes handlers
// easy to test and keeps them decoupled from specific implementations.
//
// Rate limiting is NOT handled here — it is enforced by middleware
// before requests reach these handlers (see internal/middleware).
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Handler holds dependencies needed by the HTTP handler functions.
// Currently lightweight, but this struct will grow as we add services
// and other dependencies in later phases.
type Handler struct{}

// NewHandler creates a new Handler.
//
// Example:
//
//	h := handler.NewHandler()
//	router.GET("/health", h.Health)
//	router.GET("/api/data", h.Data)
func NewHandler() *Handler {
	return &Handler{}
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

// Data handles GET /api/data.
//
// This is a sample endpoint that returns data to the client.
// Rate limiting is handled by the middleware layer — by the time
// a request reaches this handler, it has already been approved.
//
// Response (200 OK):
//
//	{"message": "Here is your data!", "client_ip": "..."}
func (h *Handler) Data(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message":   "Here is your data!",
		"client_ip": c.ClientIP(),
	})
}
