package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/kong-ji/test/internal/config"
	"github.com/kong-ji/test/internal/router"
	"github.com/kong-ji/test/internal/server"
)

func main() {
	cfg := config.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Level(),
	})))

	slog.Info("starting service", "env", cfg.Env, "log_level", cfg.LogLevel)

	srv := server.New(server.Options{
		Addr:            cfg.Addr(),
		Handler:         router.New(),
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server exited unexpectedly", "error", err)
		os.Exit(1)
	}

	slog.Info("service stopped cleanly")
}
