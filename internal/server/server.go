// Package server wraps an HTTP server with graceful shutdown support.
package server

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"
)

// Options configures the HTTP server.
type Options struct {
	Addr            string
	Handler         http.Handler
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Server is an HTTP server with graceful shutdown.
type Server struct {
	httpServer      *http.Server
	shutdownTimeout time.Duration
}

// New creates a server from the given options.
func New(o Options) *Server {
	return &Server{
		shutdownTimeout: o.ShutdownTimeout,
		httpServer: &http.Server{
			Addr:         o.Addr,
			Handler:      o.Handler,
			ReadTimeout:  o.ReadTimeout,
			WriteTimeout: o.WriteTimeout,
			IdleTimeout:  o.IdleTimeout,
		},
	}
}

// Run starts the server and blocks until ctx is cancelled or the server fails.
// On cancellation it drains in-flight requests before returning.
func (s *Server) Run(ctx context.Context) error {
	errCh := make(chan error, 1)

	go func() {
		slog.Info("server listening", "addr", s.httpServer.Addr)
		if err := s.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
	}

	slog.Info("shutdown signal received, draining connections")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), s.shutdownTimeout)
	defer cancel()

	return s.httpServer.Shutdown(shutdownCtx)
}
