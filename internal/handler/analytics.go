package handler

import (
	"net/http"

	"ioi-amms/internal/auth"
	"ioi-amms/internal/middleware"
	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"

	"github.com/go-chi/chi/v5"
)

type AnalyticsHandler struct {
	repo *repository.AnalyticsRepository
}

func NewAnalyticsHandler(repo *repository.AnalyticsRepository) *AnalyticsHandler {
	return &AnalyticsHandler{repo: repo}
}

func (h *AnalyticsHandler) RegisterRoutes(r chi.Router) {
	r.Get("/analytics/dashboard", h.GetDashboardData)
}

// GetDashboardData returns aggregated stats and alerts
func (h *AnalyticsHandler) GetDashboardData(w http.ResponseWriter, r *http.Request) {
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		unauthorizedError(w, "User not authenticated")
		return
	}

	stats, err := h.repo.GetDashboardStats(r.Context(), claims.TenantID)
	if err != nil {
		internalError(w, err.Error())
		return
	}

	alerts, err := h.repo.GetTopAlerts(r.Context(), claims.TenantID)
	if err != nil {
		internalError(w, err.Error())
		return
	}
	// Ensure alerts is empty slice not null
	if alerts == nil {
		alerts = []model.Alert{}
	}

	resp := model.AnalyticsResponse{
		Stats:  *stats,
		Alerts: alerts,
	}

	jsonResponse(w, http.StatusOK, resp)
}
