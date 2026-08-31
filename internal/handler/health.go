// Package handler contains the HTTP handlers of the service.
package handler

import (
	"net/http"
	"runtime"
	"time"

	"github.com/kong-ji/test/internal/response"
)

var startTime = time.Now()

// Index serves the service root.
func Index(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"name": "test",
		"docs": "GET /health",
		"time": time.Now().UTC().Format(time.RFC3339),
	})
}

// NotFound is the catch-all handler for unmatched routes.
func NotFound(w http.ResponseWriter, r *http.Request) {
	response.Error(w, http.StatusNotFound, "not found: "+r.URL.Path)
}

// Health reports the liveness of the service.
func Health(w http.ResponseWriter, r *http.Request) {
	response.JSON(w, http.StatusOK, map[string]any{
		"status":     "ok",
		"uptime":     time.Since(startTime).String(),
		"go":         runtime.Version(),
		"goroutines": runtime.NumGoroutine(),
	})
}
