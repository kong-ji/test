// Package response provides helpers for writing HTTP responses.
package response

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// JSON writes body as a JSON document with the given status code.
func JSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if body == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(body); err != nil {
		slog.Error("write json response failed", "error", err)
	}
}

// Error writes a JSON error payload in the shape {"error": message}.
func Error(w http.ResponseWriter, status int, message string) {
	JSON(w, status, map[string]string{"error": message})
}
