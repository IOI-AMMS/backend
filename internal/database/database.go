package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"ioi-amms/internal/config"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Service interface {
	Health() map[string]string
	Close()
	Pool() *pgxpool.Pool
}

type service struct {
	db  *pgxpool.Pool
	cfg *config.DatabaseConfig
}

// NewWithConfig creates a database service using the config package
func NewWithConfig(cfg *config.Config) Service {
	connStr := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Database.User,
		cfg.Database.Password,
		cfg.Database.Host,
		cfg.Database.Port,
		cfg.Database.Name,
		cfg.Database.SSLMode,
	)

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		slog.Error("Failed to parse database config", slog.String("error", err.Error()))
		panic(err)
	}

	poolConfig.MaxConns = int32(cfg.Database.MaxConns)

	db, err := pgxpool.NewWithConfig(context.Background(), poolConfig)
	if err != nil {
		slog.Error("Failed to connect to database", slog.String("error", err.Error()))
		panic(err)
	}

	slog.Info("Connected to database",
		slog.String("host", cfg.Database.Host),
		slog.String("database", cfg.Database.Name),
	)

	return &service{
		db:  db,
		cfg: &cfg.Database,
	}
}

// Pool returns the underlying connection pool for repositories
func (s *service) Pool() *pgxpool.Pool {
	return s.db
}

// Health checks the health of the database connection
func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx)
	if err != nil {
		slog.Error("Database health check failed", slog.String("error", err.Error()))
		return map[string]string{
			"status": "down",
			"error":  err.Error(),
		}
	}

	return map[string]string{
		"status":  "up",
		"message": "It's healthy",
	}
}

func (s *service) Close() {
	slog.Info("Disconnected from database", slog.String("database", s.cfg.Name))
	s.db.Close()
}
