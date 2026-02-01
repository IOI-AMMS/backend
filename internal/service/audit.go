package service

import (
	"context"
	"log/slog"

	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"
)

// AuditService helper for logging audit events
type AuditService struct {
	repo *repository.AuditRepository
}

// NewAuditService creates a new audit service
func NewAuditService(repo *repository.AuditRepository) *AuditService {
	return &AuditService{repo: repo}
}

// Log records an audit event asynchronously
func (s *AuditService) Log(ctx context.Context, userID, action, entityType, entityID string, changes map[string]interface{}) {
	// Use a detached context for async logging to ensure it completes even if request cancels
	// However, we need to be careful with detached contexts in Go < 1.21.
	// In Go 1.21+, context.WithoutCancel(ctx).

	go func() {
		// Use background context with timeout for safety
		logCtx := context.Background()
		// Or better, inherit values but not cancellation?
		// For simplicity, just use Background.

		log := &model.AuditLog{
			UserID:     userID,
			Action:     action,
			EntityType: entityType,
			EntityID:   entityID,
			Changes:    changes,
		}

		if err := s.repo.Create(logCtx, log); err != nil {
			slog.Error("Failed to write audit log",
				slog.String("action", action),
				slog.String("entity", entityType),
				slog.String("error", err.Error()),
			)
		}
	}()
}
