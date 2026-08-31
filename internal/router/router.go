// Package router wires HTTP routes to their handlers.
package router

import (
	"net/http"
	"strings"

	"github.com/kong-ji/test/internal/handler"
	"github.com/kong-ji/test/internal/response"
)

// route describes a single registered endpoint.
type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

// routes is the single source of truth for the API surface.
var routes = []route{
	// "{$}" anchors the pattern to the root path only; a bare "GET /" would
	// match every GET request and silently swallow unknown routes.
	{http.MethodGet, "/{$}", handler.Index},
	{http.MethodGet, "/healthz", handler.Health},
}

// New builds the application HTTP handler.
func New() http.Handler {
	mux := http.NewServeMux()

	allowed := make(map[string][]string, len(routes))
	for _, r := range routes {
		mux.HandleFunc(r.method+" "+r.pattern, r.handler)
		key := normalizePath(r.pattern)
		allowed[key] = append(allowed[key], r.method)
	}

	// Catch-all: answer 405 when the path exists but the method does not,
	// 404 when nothing matches at all.
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimSuffix(r.URL.Path, "/")
		if key == "" {
			key = "/"
		}
		if methods, ok := allowed[key]; ok {
			w.Header().Set("Allow", strings.Join(methods, ", "))
			response.Error(w, http.StatusMethodNotAllowed,
				"method not allowed: "+r.Method+" "+r.URL.Path)
			return
		}
		handler.NotFound(w, r)
	})

	return mux
}

// normalizePath turns a ServeMux pattern into a comparable request path,
// e.g. "/{$}" -> "/" and "/healthz" -> "/healthz".
func normalizePath(pattern string) string {
	return "/" + strings.Trim(strings.TrimSuffix(pattern, "{$}"), "/")
}
