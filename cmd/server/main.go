package main

import (
	"log/slog"
	"os"

	"ioi-amms/internal/config"
	"ioi-amms/internal/server"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Setup structured logging
	setupLogger(cfg)

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		slog.Error("Invalid configuration", slog.String("error", err.Error()))
		os.Exit(1)
	}

	slog.Info("Starting IOI AMMS Backend",
		slog.String("environment", cfg.Environment),
		slog.Int("port", cfg.Server.Port),
	)

	// Run server with graceful shutdown
	if err := server.Run(cfg); err != nil {
		slog.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func setupLogger(cfg *config.Config) {
	var handler slog.Handler

	opts := &slog.HandlerOptions{
		Level: parseLogLevel(cfg.Log.Level),
	}

	if cfg.Log.Format == "json" || cfg.IsProduction() {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	slog.SetDefault(slog.New(handler))
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
