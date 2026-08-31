// Package config loads application configuration from environment variables.
package config

import (
	"log/slog"
	"os"
	"strconv"
	"strings"
	"time"
)

// Config holds the runtime configuration of the service.
type Config struct {
	Env             string
	Port            string
	LogLevel        string
	ReadTimeout     time.Duration
	WriteTimeout    time.Duration
	IdleTimeout     time.Duration
	ShutdownTimeout time.Duration
}

// Load reads configuration from the environment, falling back to defaults.
func Load() Config {
	return Config{
		Env:             envOrDefault("APP_ENV", "development"),
		Port:            envOrDefault("APP_PORT", "8080"),
		LogLevel:        envOrDefault("APP_LOG_LEVEL", "info"),
		ReadTimeout:     durationOrDefault("APP_READ_TIMEOUT", 5*time.Second),
		WriteTimeout:    durationOrDefault("APP_WRITE_TIMEOUT", 10*time.Second),
		IdleTimeout:     durationOrDefault("APP_IDLE_TIMEOUT", 60*time.Second),
		ShutdownTimeout: durationOrDefault("APP_SHUTDOWN_TIMEOUT", 10*time.Second),
	}
}

// Addr returns the address the HTTP server listens on.
func (c Config) Addr() string {
	if strings.Contains(c.Port, ":") {
		return c.Port
	}
	return ":" + c.Port
}

// Level parses the configured log level, defaulting to info on invalid input.
func (c Config) Level() slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(c.LogLevel)); err != nil {
		return slog.LevelInfo
	}
	return level
}

func envOrDefault(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func durationOrDefault(key string, fallback time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback
	}
	if v, err := time.ParseDuration(raw); err == nil {
		return v
	}
	if seconds, err := strconv.Atoi(raw); err == nil {
		return time.Duration(seconds) * time.Second
	}
	return fallback
}
