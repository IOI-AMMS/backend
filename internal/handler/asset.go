package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"ioi-amms/internal/middleware"
	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"

	"github.com/go-chi/chi/v5"
)

// AssetHandler handles asset-related HTTP requests
type AssetHandler struct {
	repo *repository.AssetRepository
}

// NewAssetHandler creates a new asset handler
func NewAssetHandler(repo *repository.AssetRepository) *AssetHandler {
	return &AssetHandler{repo: repo}
}

// RegisterRoutes registers asset routes
func (h *AssetHandler) RegisterRoutes(r chi.Router) {
	r.Get("/assets", h.List)
	r.Post("/assets", h.Create)
	r.Get("/assets/{id}", h.Get)
	r.Put("/assets/{id}", h.Update)
	r.Delete("/assets/{id}", h.Delete)
}

// List handles GET /assets
func (h *AssetHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	// Parse query params
	params := model.AssetListParams{
		TenantID:    claims.TenantID,
		Status:      r.URL.Query()["status"],
		Criticality: r.URL.Query()["criticality"],
		Search:      r.URL.Query().Get("search"),
		SortBy:      r.URL.Query().Get("sortBy"),
		SortDir:     r.URL.Query().Get("sortDir"),
		Page:        parseIntParam(r, "page", 1),
		Limit:       parseIntParam(r, "limit", 10),
	}

	result, err := h.repo.List(r.Context(), params)
	if err != nil {
		slog.Error("Failed to list assets", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch assets")
		return
	}

	// Format response per frontend spec
	resp := map[string]interface{}{
		"data": result.Data,
		"meta": map[string]int{
			"total":      result.Total,
			"page":       result.Page,
			"limit":      result.Limit,
			"totalPages": result.TotalPages,
		},
	}

	jsonResponse(w, http.StatusOK, resp)
}

// Get handles GET /assets/{id}
func (h *AssetHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	asset, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrAssetNotFound {
			errorResponse(w, http.StatusNotFound, "Asset not found")
			return
		}
		slog.Error("Failed to get asset", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch asset")
		return
	}

	jsonResponse(w, http.StatusOK, asset)
}

// CreateAssetRequest represents the create asset request body
type CreateAssetRequest struct {
	ParentID    *string `json:"parentId"`
	LocationID  *string `json:"locationId"`
	Name        string  `json:"name"`
	Status      string  `json:"status"`
	Criticality string  `json:"criticality"`
}

// Create handles POST /assets
func (h *AssetHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Name == "" {
		errorResponse(w, http.StatusBadRequest, "Name is required")
		return
	}

	// Set defaults
	if req.Status == "" {
		req.Status = model.AssetStatusPending
	}
	if req.Criticality == "" {
		req.Criticality = model.AssetCriticalityMedium
	}

	asset := &model.Asset{
		TenantID:    claims.TenantID,
		ParentID:    req.ParentID,
		LocationID:  req.LocationID,
		Name:        req.Name,
		Status:      req.Status,
		Criticality: req.Criticality,
	}

	if err := h.repo.Create(r.Context(), asset); err != nil {
		slog.Error("Failed to create asset", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to create asset")
		return
	}

	jsonResponse(w, http.StatusCreated, asset)
}

// Update handles PUT /assets/{id}
func (h *AssetHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	// Verify asset exists
	existing, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrAssetNotFound {
			errorResponse(w, http.StatusNotFound, "Asset not found")
			return
		}
		slog.Error("Failed to find asset", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update asset")
		return
	}

	var req CreateAssetRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Update fields
	existing.Name = req.Name
	existing.ParentID = req.ParentID
	existing.LocationID = req.LocationID
	existing.Status = req.Status
	existing.Criticality = req.Criticality

	if err := h.repo.Update(r.Context(), existing); err != nil {
		slog.Error("Failed to update asset", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update asset")
		return
	}

	jsonResponse(w, http.StatusOK, existing)
}

// Delete handles DELETE /assets/{id}
func (h *AssetHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.repo.Delete(r.Context(), id); err != nil {
		if err == repository.ErrAssetNotFound {
			errorResponse(w, http.StatusNotFound, "Asset not found")
			return
		}
		slog.Error("Failed to delete asset", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to delete asset")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Helper functions
func parseIntParam(r *http.Request, key string, defaultVal int) int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return defaultVal
	}
	parsed, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return parsed
}
