package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ioi-amms/internal/middleware"
	"ioi-amms/internal/repository"

	"github.com/go-chi/chi/v5"
)

// TenantHandler handles tenant-related HTTP requests
type TenantHandler struct {
	repo *repository.TenantRepository
}

// NewTenantHandler creates a new tenant handler
func NewTenantHandler(repo *repository.TenantRepository) *TenantHandler {
	return &TenantHandler{repo: repo}
}

// RegisterRoutes registers tenant routes
func (h *TenantHandler) RegisterRoutes(r chi.Router) {
	r.Get("/tenant/settings", h.GetSettings)
	r.Patch("/tenant/settings", h.UpdateSettings)
}

// GetSettings handles GET /tenant/settings
func (h *TenantHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	// 1. Authorization: Only Admin (enforced by middleware, but good to double check or if reused)
	// Actually, route registration should handle middleware wrapping.

	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	settings, err := h.repo.GetSettings(r.Context(), claims.TenantID)
	if err != nil {
		if err == repository.ErrTenantNotFound {
			errorResponse(w, http.StatusNotFound, "Tenant not found")
			return
		}
		slog.Error("Failed to get tenant settings", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch settings")
		return
	}

	// Format response
	resp := map[string]interface{}{
		"tenantId": settings.ID,
		"name":     settings.Name,
		"settings": settings.Settings,
	}

	jsonResponse(w, http.StatusOK, resp)
}

// UpdateSettings handles PATCH /tenant/settings
func (h *TenantHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req struct {
		Settings map[string]interface{} `json:"settings"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Settings == nil {
		errorResponse(w, http.StatusBadRequest, "Settings object is required")
		return
	}

	updated, err := h.repo.UpdateSettings(r.Context(), claims.TenantID, req.Settings)
	if err != nil {
		slog.Error("Failed to update tenant settings", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update settings")
		return
	}

	resp := map[string]interface{}{
		"tenantId": updated.ID,
		"name":     updated.Name,
		"settings": updated.Settings,
	}

	jsonResponse(w, http.StatusOK, resp)
}
