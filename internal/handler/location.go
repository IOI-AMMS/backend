package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ioi-amms/internal/middleware"
	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"

	"github.com/go-chi/chi/v5"
)

// LocationHandler handles location HTTP requests
type LocationHandler struct {
	repo *repository.LocationRepository
}

// NewLocationHandler creates a new location handler
func NewLocationHandler(repo *repository.LocationRepository) *LocationHandler {
	return &LocationHandler{repo: repo}
}

// RegisterRoutes registers location routes
func (h *LocationHandler) RegisterRoutes(r chi.Router) {
	r.Get("/locations", h.List)
	r.Post("/locations", h.Create)
	r.Get("/locations/{id}", h.Get)
	r.Get("/locations/{id}/children", h.ListChildren)
	r.Put("/locations/{id}", h.Update)
	r.Delete("/locations/{id}", h.Delete)
}

// List handles GET /locations
func (h *LocationHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	locations, err := h.repo.ListByTenant(r.Context(), claims.TenantID)
	if err != nil {
		slog.Error("Failed to list locations", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch locations")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"data": locations})
}

// Get handles GET /locations/{id}
func (h *LocationHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	loc, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrLocationNotFound {
			errorResponse(w, http.StatusNotFound, "Location not found")
			return
		}
		slog.Error("Failed to get location", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch location")
		return
	}

	jsonResponse(w, http.StatusOK, loc)
}

// ListChildren handles GET /locations/{id}/children
func (h *LocationHandler) ListChildren(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	children, err := h.repo.ListChildren(r.Context(), id)
	if err != nil {
		slog.Error("Failed to list children", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch children")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"data": children})
}

// CreateLocationRequest represents create location request body
type CreateLocationRequest struct {
	ParentID *string `json:"parentId"`
	Name     string  `json:"name"`
	Type     string  `json:"type"`
}

// Create handles POST /locations
func (h *LocationHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" || req.Type == "" {
		errorResponse(w, http.StatusBadRequest, "Name and type are required")
		return
	}

	loc := &model.Location{
		TenantID: claims.TenantID,
		ParentID: req.ParentID,
		Name:     req.Name,
		Type:     req.Type,
	}

	if err := h.repo.Create(r.Context(), loc); err != nil {
		slog.Error("Failed to create location", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to create location")
		return
	}

	jsonResponse(w, http.StatusCreated, loc)
}

// Update handles PUT /locations/{id}
func (h *LocationHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrLocationNotFound {
			errorResponse(w, http.StatusNotFound, "Location not found")
			return
		}
		slog.Error("Failed to find location", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update location")
		return
	}

	var req CreateLocationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	existing.ParentID = req.ParentID
	existing.Name = req.Name
	existing.Type = req.Type

	if err := h.repo.Update(r.Context(), existing); err != nil {
		slog.Error("Failed to update location", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update location")
		return
	}

	jsonResponse(w, http.StatusOK, existing)
}

// Delete handles DELETE /locations/{id}
func (h *LocationHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err == repository.ErrLocationNotFound {
			errorResponse(w, http.StatusNotFound, "Location not found")
			return
		}
		slog.Error("Failed to delete location", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to delete location")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
