package handler

import (
	"net/http"
	"time"

	"ioi-amms/internal/middleware"
	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"
)

// AuditHandler handles audit log requests
type AuditHandler struct {
	repo *repository.AuditRepository
}

// NewAuditHandler creates a new audit handler
func NewAuditHandler(repo *repository.AuditRepository) *AuditHandler {
	return &AuditHandler{repo: repo}
}

// RegisterRoutes registers audit routes
func (h *AuditHandler) RegisterRoutes(r *http.ServeMux) {
	// Not used with chi, handled in routes.go
}

// List handles GET /audit-logs
func (h *AuditHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse filtering params
	query := r.URL.Query()

	params := model.AuditListParams{
		TenantID:   claims.TenantID, // Enforce tenant isolation
		UserID:     query.Get("userId"),
		EntityType: query.Get("entityType"),
		EntityID:   query.Get("entityId"),
		Action:     query.Get("action"),
		Page:       parseIntParam(r, "page", 1),
		Limit:      parseIntParam(r, "limit", 20),
	}

	if fromStr := query.Get("from"); fromStr != "" {
		if t, err := time.Parse(time.RFC3339, fromStr); err == nil {
			params.From = t
		}
	}

	if toStr := query.Get("to"); toStr != "" {
		if t, err := time.Parse(time.RFC3339, toStr); err == nil {
			params.To = t
		}
	}

	result, err := h.repo.List(r.Context(), params)
	if err != nil {
		errorResponse(w, http.StatusInternalServerError, "Failed to list audit logs")
		return
	}

	jsonResponse(w, http.StatusOK, result)
}
