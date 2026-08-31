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
	"github.com/kong-ji/test/internal/fingerprint"
	"github.com/kong-ji/test/internal/router"
	"github.com/kong-ji/test/internal/rules"
	"github.com/kong-ji/test/internal/server"
)

func main() {
	cfg := config.Load()

	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Level(),
	})))

	slog.Info("starting service", "env", cfg.Env, "log_level", cfg.LogLevel)

	rl, err := rules.Load(cfg.RulesPath)
	if err != nil {
		slog.Error("failed to load rules", "path", cfg.RulesPath, "error", err)
		os.Exit(1)
	}

	engine := fingerprint.New(rl)
	handler := router.NewWithEngine(engine)

	srv := server.New(server.Options{
		Addr:            cfg.Addr(),
		Handler:         handler,
		ReadTimeout:     cfg.ReadTimeout,
		WriteTimeout:    cfg.WriteTimeout,
		IdleTimeout:     cfg.IdleTimeout,
		ShutdownTimeout: cfg.ShutdownTimeout,
	})

	// SIGHUP placeholder: rule hot-reload is not yet wired to a dynamic engine
	// swap; log receipt so the signal is acknowledged without disrupting the
	// running server.
	sighupCh := make(chan os.Signal, 1)
	signal.Notify(sighupCh, syscall.SIGHUP)
	go func() {
		for range sighupCh {
			slog.Info("SIGHUP received, rules reload not yet wired")
		}
	}()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	if err := srv.Run(ctx); err != nil && !errors.Is(err, http.ErrServerClosed) {
		slog.Error("server exited unexpectedly", "error", err)
		os.Exit(1)
	}

	slog.Info("service stopped cleanly")
}
