package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ioi-amms/internal/config"
	"ioi-amms/internal/database"
	"ioi-amms/internal/routes"
)

type Server struct {
	config *config.Config
	db     database.Service
	http   *http.Server
}

func NewServer(cfg *config.Config) *Server {
	db := database.NewWithConfig(cfg)

	server := &Server{
		config: cfg,
		db:     db,
		http: &http.Server{
			Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
			Handler:      routes.NewRouter(db, cfg),
			IdleTimeout:  cfg.Server.IdleTimeout,
			ReadTimeout:  cfg.Server.ReadTimeout,
			WriteTimeout: cfg.Server.WriteTimeout,
		},
	}

	return server
}

// Start runs the HTTP server
func (s *Server) Start() error {
	slog.Info("Starting server",
		slog.Int("port", s.config.Server.Port),
		slog.String("environment", s.config.Environment),
	)
	return s.http.ListenAndServe()
}

// Shutdown gracefully stops the server
func (s *Server) Shutdown(ctx context.Context) error {
	slog.Info("Shutting down server...")

	if err := s.http.Shutdown(ctx); err != nil {
		return err
	}

	s.db.Close()
	slog.Info("Server shutdown complete")
	return nil
}

// Run starts the server and handles graceful shutdown
func Run(cfg *config.Config) error {
	server := NewServer(cfg)

	// Channel to listen for shutdown signals
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	// Run server in goroutine
	go func() {
		if err := server.Start(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server error", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	// Wait for shutdown signal
	<-stop
	slog.Info("Received shutdown signal")

	// Create shutdown context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return server.Shutdown(ctx)
}
