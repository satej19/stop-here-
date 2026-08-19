// Package config handles application configuration (ports, limits, environment settings).
//
// Configuration is loaded from environment variables with sensible defaults.
// This allows the same binary to run in different environments (dev, staging, prod)
// simply by changing environment variables — no code changes needed.
package config

import (
	"log"
	"os"
	"strconv"
	"time"
)

// Config holds all application configuration values.
// Each field maps to an environment variable with a default fallback.
type Config struct {
	// Port is the HTTP server listen port.
	// Env: PORT | Default: 8080
	Port string

	// RateLimit is the maximum number of requests a single client
	// can make within one rate-limit window.
	// Env: RATE_LIMIT | Default: 10
	RateLimit int

	// RateWindow is the duration of the rate-limit window.
	// After this duration, the client's request count resets.
	// Env: RATE_WINDOW | Default: 1m (60 seconds)
	RateWindow time.Duration
}

// Load reads configuration from environment variables and returns a Config.
// Missing or invalid values fall back to safe defaults. Invalid numeric or
// duration values are logged as warnings so operators notice misconfigurations.
func Load() Config {
	cfg := Config{
		Port:       getEnv("PORT", "8080"),
		RateLimit:  getEnvInt("RATE_LIMIT", 10),
		RateWindow: getEnvDuration("RATE_WINDOW", 60*time.Second),
	}

	log.Printf("[config] Port=%s  RateLimit=%d  RateWindow=%s",
		cfg.Port, cfg.RateLimit, cfg.RateWindow)

	return cfg
}

// getEnv returns the value of the environment variable named by key,
// or defaultVal if the variable is not set or is empty.
func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

// getEnvInt reads an integer from the environment variable named by key.
// Returns defaultVal if the variable is missing, empty, or not a valid integer.
func getEnvInt(key string, defaultVal int) int {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	n, err := strconv.Atoi(val)
	if err != nil {
		log.Printf("[config] WARNING: invalid integer for %s=%q, using default %d", key, val, defaultVal)
		return defaultVal
	}
	return n
}

// getEnvDuration reads a time.Duration from the environment variable named by key.
// The value must be in Go duration format (e.g., "30s", "2m", "1h").
// Returns defaultVal if the variable is missing, empty, or not a valid duration.
func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	d, err := time.ParseDuration(val)
	if err != nil {
		log.Printf("[config] WARNING: invalid duration for %s=%q, using default %s", key, val, defaultVal)
		return defaultVal
	}
	return d
}
