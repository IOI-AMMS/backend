package main

import (
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"strings"

	"ioi-amms/internal/config"
	"ioi-amms/internal/server"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	// Run database migrations
	if err := runMigrations(cfg); err != nil {
		slog.Error("Migration failed", slog.String("error", err.Error()))
		// Don't exit on migration error in dev, but maybe in prod?
		// For now, log fatal
		os.Exit(1)
	}

	// Run server with graceful shutdown
	if err := server.Run(cfg); err != nil {
		slog.Error("Server error", slog.String("error", err.Error()))
		os.Exit(1)
	}
}

func runMigrations(cfg *config.Config) error {
	// Construct database URL safely using net/url
	dbURI := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(cfg.Database.User, cfg.Database.Password),
		Host:     fmt.Sprintf("%s:%s", cfg.Database.Host, cfg.Database.Port),
		Path:     cfg.Database.Name,
		RawQuery: fmt.Sprintf("sslmode=%s", cfg.Database.SSLMode),
	}

	// Point to the migrations folder inside the container
	// Dockerfile copies to /app/migrations
	sourceURL := "file:///app/migrations"

	// If running locally (dev), try relative paths
	if _, err := os.Stat("/app/migrations"); os.IsNotExist(err) {
		if _, err := os.Stat("migrations"); err == nil {
			sourceURL = "file://migrations"
		} else if _, err := os.Stat("../migrations"); err == nil {
			sourceURL = "file://../migrations"
		}
	}

	m, err := migrate.New(sourceURL, dbURI.String())
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		// Check if it's a dirty database error
		if strings.Contains(err.Error(), "Dirty database") {
			slog.Warn("Dirty database detected, dropping and rebuilding schema")
			// Drop all tables and schema_migrations
			if dropErr := m.Drop(); dropErr != nil {
				return fmt.Errorf("failed to drop database: %w", dropErr)
			}
			// Need new migrate instance after Drop
			m2, err := migrate.New(sourceURL, dbURI.String())
			if err != nil {
				return fmt.Errorf("failed to create new migrate instance: %w", err)
			}
			defer m2.Close()
			// Retry migration from scratch
			if retryErr := m2.Up(); retryErr != nil && retryErr != migrate.ErrNoChange {
				return fmt.Errorf("failed to run migrate up after drop: %w", retryErr)
			}
		} else {
			return fmt.Errorf("failed to run migrate up: %w", err)
		}
	}

	slog.Info("Database migrations applied successfully")
	return nil
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
