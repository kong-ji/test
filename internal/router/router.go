// Package router wires HTTP routes to their handlers.
package router

import (
	"net/http"
	"strings"

	"github.com/kong-ji/test/internal/fingerprint"
	"github.com/kong-ji/test/internal/handler"
	"github.com/kong-ji/test/internal/response"
)

// route describes a single registered endpoint.
type route struct {
	method  string
	pattern string
	handler http.HandlerFunc
}

// baseRoutes is the single source of truth for the base API surface.
var baseRoutes = []route{
	// "{$}" anchors the pattern to the root path only; a bare "GET /" would
	// match every GET request and silently swallow unknown routes.
	{http.MethodGet, "/{$}", handler.Index},
	{http.MethodGet, "/health", handler.Health},
}

// New builds the application HTTP handler without the identify endpoint.
func New() http.Handler {
	mux := http.NewServeMux()
	registerBase(mux)
	installCatchAll(mux)
	return mux
}

// NewWithEngine builds the application HTTP handler including the identify
// endpoint, backed by the given fingerprint engine.
func NewWithEngine(e *fingerprint.Engine) http.Handler {
	mux := http.NewServeMux()
	registerBase(mux)
	mux.Handle("POST /fingerprint", handler.NewIdentifyHandler(e))
	installCatchAll(mux)
	return mux
}

// registerBase registers the base routes on the mux.
func registerBase(mux *http.ServeMux) {
	for _, r := range baseRoutes {
		mux.HandleFunc(r.method+" "+r.pattern, r.handler)
	}
}

// installCatchAll installs the catch-all handler that answers 405 when the
// path exists but the method does not, and 404 when nothing matches at all.
func installCatchAll(mux *http.ServeMux) {
	allowed := make(map[string][]string)
	for _, r := range baseRoutes {
		key := normalizePath(r.pattern)
		allowed[key] = append(allowed[key], r.method)
	}
	allowed["/fingerprint"] = []string{http.MethodPost}

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
}

// normalizePath turns a ServeMux pattern into a comparable request path,
// e.g. "/{$}" -> "/" and "/healthz" -> "/healthz".
func normalizePath(pattern string) string {
	return "/" + strings.Trim(strings.TrimSuffix(pattern, "{$}"), "/")
}
