package handler

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"

	"ioi-amms/internal/middleware"
	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"
	"ioi-amms/internal/service"

	"github.com/go-chi/chi/v5"
)

// AssetHandler handles asset-related HTTP requests
type AssetHandler struct {
	repo  *repository.AssetRepository
	audit *service.AuditService
}

// NewAssetHandler creates a new asset handler
func NewAssetHandler(repo *repository.AssetRepository, audit *service.AuditService) *AssetHandler {
	return &AssetHandler{repo: repo, audit: audit}
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

	// Parse query params (v1.1 schema)
	params := model.AssetListParams{
		TenantID:  claims.TenantID,
		Status:    r.URL.Query()["status"],
		OrgUnitID: r.URL.Query().Get("orgUnitId"),
		Search:    r.URL.Query().Get("search"),
		SortBy:    r.URL.Query().Get("sortBy"),
		SortDir:   r.URL.Query().Get("sortDir"),
		Page:      parseIntParam(r, "page", 1),
		Limit:     parseIntParam(r, "limit", 10),
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

// CreateAssetRequest represents the create asset request body (v1.1)
type CreateAssetRequest struct {
	ParentID       *string         `json:"parentId"`
	LocationID     *string         `json:"locationId"`
	OrgUnitID      *string         `json:"orgUnitId"`
	Name           string          `json:"name"`
	Status         string          `json:"status"`
	IsFieldRelated *bool           `json:"isFieldRelated"`
	Manufacturer   *string         `json:"manufacturer"`
	ModelNumber    *string         `json:"modelNumber"`
	Specs          json.RawMessage `json:"specs"`
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
		req.Status = model.AssetStatusDraft
	}
	isFieldRelated := true
	if req.IsFieldRelated != nil {
		isFieldRelated = *req.IsFieldRelated
	}

	asset := &model.Asset{
		TenantID:       claims.TenantID,
		ParentID:       req.ParentID,
		LocationID:     req.LocationID,
		OrgUnitID:      req.OrgUnitID,
		Name:           req.Name,
		Status:         req.Status,
		IsFieldRelated: isFieldRelated,
		Manufacturer:   req.Manufacturer,
		ModelNumber:    req.ModelNumber,
		Specs:          req.Specs,
	}

	if err := h.repo.Create(r.Context(), asset); err != nil {
		slog.Error("Failed to create asset", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to create asset")
		return
	}

	// Audit Log
	h.audit.Log(r.Context(), claims.UserID, model.AuditActionCreate, model.AuditEntityAsset, asset.ID, map[string]interface{}{
		"name": asset.Name,
	})

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

	// Update fields only if provided (Partial Update)
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.ParentID != nil {
		existing.ParentID = req.ParentID
	}
	if req.LocationID != nil {
		existing.LocationID = req.LocationID
	}
	if req.OrgUnitID != nil {
		existing.OrgUnitID = req.OrgUnitID
	}
	if req.Status != "" {
		existing.Status = req.Status
	}
	if req.IsFieldRelated != nil {
		existing.IsFieldRelated = *req.IsFieldRelated
	}
	if req.Manufacturer != nil {
		existing.Manufacturer = req.Manufacturer
	}
	if req.ModelNumber != nil {
		existing.ModelNumber = req.ModelNumber
	}
	if req.Specs != nil {
		existing.Specs = req.Specs
	}

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

// parseIntParam helper
func parseIntParam(r *http.Request, name string, defaultVal int) int {
	val := r.URL.Query().Get(name)
	if val == "" {
		return defaultVal
	}
	i, err := strconv.Atoi(val)
	if err != nil {
		return defaultVal
	}
	return i
}
