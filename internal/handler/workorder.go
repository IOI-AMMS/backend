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

// WorkOrderHandler handles work order HTTP requests
type WorkOrderHandler struct {
	repo *repository.WorkOrderRepository
}

// NewWorkOrderHandler creates a new work order handler
func NewWorkOrderHandler(repo *repository.WorkOrderRepository) *WorkOrderHandler {
	return &WorkOrderHandler{repo: repo}
}

// RegisterRoutes registers work order routes
func (h *WorkOrderHandler) RegisterRoutes(r chi.Router) {
	r.Get("/work-orders", h.List)
	r.Post("/work-orders", h.Create)
	r.Get("/work-orders/{id}", h.Get)
	r.Put("/work-orders/{id}", h.Update)
	r.Patch("/work-orders/{id}/status", h.UpdateStatus)
}

// List handles GET /work-orders
func (h *WorkOrderHandler) List(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := model.WorkOrderListParams{
		TenantID: claims.TenantID,
		Status:   r.URL.Query()["status"],
		Priority: r.URL.Query()["priority"],
		AssetID:  r.URL.Query().Get("assetId"),
		SortBy:   r.URL.Query().Get("sortBy"),
		SortDir:  r.URL.Query().Get("sortDir"),
		Page:     parseIntParam(r, "page", 1),
		Limit:    parseIntParam(r, "limit", 10),
	}

	result, err := h.repo.List(r.Context(), params)
	if err != nil {
		slog.Error("Failed to list work orders", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch work orders")
		return
	}

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

// Get handles GET /work-orders/{id}
func (h *WorkOrderHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	wo, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrWorkOrderNotFound {
			errorResponse(w, http.StatusNotFound, "Work order not found")
			return
		}
		slog.Error("Failed to get work order", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch work order")
		return
	}

	jsonResponse(w, http.StatusOK, wo)
}

// CreateWorkOrderRequest represents the create work order request body (v1.1)
type CreateWorkOrderRequest struct {
	AssetID     *string `json:"assetId"`
	Priority    string  `json:"priority"`
	Origin      string  `json:"origin"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

// Create handles POST /work-orders
func (h *WorkOrderHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreateWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Origin == "" {
		errorResponse(w, http.StatusBadRequest, "Origin is required")
		return
	}

	if req.Title == "" {
		errorResponse(w, http.StatusBadRequest, "Title is required")
		return
	}

	// Set defaults
	if req.Priority == "" {
		req.Priority = model.WOPriorityMedium
	}

	wo := &model.WorkOrder{
		TenantID:    claims.TenantID,
		AssetID:     req.AssetID,
		Status:      model.WOStatusRequested, // v1.1 default status
		Priority:    req.Priority,
		Origin:      req.Origin,
		Title:       req.Title,
		Description: req.Description,
	}

	if err := h.repo.Create(r.Context(), wo); err != nil {
		slog.Error("Failed to create work order", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to create work order")
		return
	}

	jsonResponse(w, http.StatusCreated, wo)
}

// UpdateWorkOrderRequest for updating work orders (v1.1)
type UpdateWorkOrderRequest struct {
	Priority    string  `json:"priority"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
}

// Update handles PUT /work-orders/{id}
func (h *WorkOrderHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.repo.FindByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrWorkOrderNotFound {
			errorResponse(w, http.StatusNotFound, "Work order not found")
			return
		}
		slog.Error("Failed to find work order", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update work order")
		return
	}

	var req UpdateWorkOrderRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Priority != "" {
		existing.Priority = req.Priority
	}
	if req.Title != "" {
		existing.Title = req.Title
	}
	existing.Description = req.Description

	if err := h.repo.Update(r.Context(), existing); err != nil {
		slog.Error("Failed to update work order", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update work order")
		return
	}

	jsonResponse(w, http.StatusOK, existing)
}

// UpdateStatusRequest represents status update request
type UpdateStatusRequest struct {
	Status string `json:"status"`
}

// UpdateStatus handles PATCH /work-orders/{id}/status
func (h *WorkOrderHandler) UpdateStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var req UpdateStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.Status == "" {
		errorResponse(w, http.StatusBadRequest, "Status is required")
		return
	}

	if err := h.repo.UpdateStatus(r.Context(), id, req.Status); err != nil {
		slog.Error("Failed to update work order status", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update status")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"status": req.Status})
}
