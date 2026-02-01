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

// InventoryHandler handles inventory HTTP requests
type InventoryHandler struct {
	repo *repository.InventoryRepository
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(repo *repository.InventoryRepository) *InventoryHandler {
	return &InventoryHandler{repo: repo}
}

// RegisterRoutes registers inventory routes
func (h *InventoryHandler) RegisterRoutes(r chi.Router) {
	// Parts catalog
	r.Get("/parts", h.ListParts)
	r.Post("/parts", h.CreatePart)
	r.Get("/parts/{id}", h.GetPart)
	r.Put("/parts/{id}", h.UpdatePart)
	r.Delete("/parts/{id}", h.DeletePart)

	// Inventory stock
	r.Get("/inventory/stock", h.ListStock)
	r.Post("/inventory/stock", h.UpsertStock)
	r.Post("/inventory/stock/adjust", h.AdjustStock)

	// User wallets
	r.Get("/inventory/wallets", h.ListMyWallet)
	r.Get("/inventory/wallets/{userId}", h.ListUserWallet)
}

// ==================== PARTS ====================

// ListParts handles GET /parts
func (h *InventoryHandler) ListParts(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := model.PartListParams{
		TenantID: claims.TenantID,
		Category: r.URL.Query().Get("category"),
		Search:   r.URL.Query().Get("search"),
		Page:     parseIntParam(r, "page", 1),
		Limit:    parseIntParam(r, "limit", 10),
	}

	result, err := h.repo.ListParts(r.Context(), params)
	if err != nil {
		slog.Error("Failed to list parts", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch parts")
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

// GetPart handles GET /parts/{id}
func (h *InventoryHandler) GetPart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	part, err := h.repo.FindPartByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrPartNotFound {
			errorResponse(w, http.StatusNotFound, "Part not found")
			return
		}
		slog.Error("Failed to get part", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch part")
		return
	}

	jsonResponse(w, http.StatusOK, part)
}

// CreatePartRequest for creating a part
type CreatePartRequest struct {
	SKU           string  `json:"sku"`
	Name          string  `json:"name"`
	Category      *string `json:"category"`
	UOM           string  `json:"uom"`
	MinStockLevel float64 `json:"minStockLevel"`
	IsStockItem   *bool   `json:"isStockItem"`
}

// CreatePart handles POST /parts
func (h *InventoryHandler) CreatePart(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req CreatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SKU == "" || req.Name == "" {
		errorResponse(w, http.StatusBadRequest, "SKU and Name are required")
		return
	}

	// Defaults
	if req.UOM == "" {
		req.UOM = "Each"
	}
	isStockItem := true
	if req.IsStockItem != nil {
		isStockItem = *req.IsStockItem
	}

	part := &model.Part{
		TenantID:      claims.TenantID,
		SKU:           req.SKU,
		Name:          req.Name,
		Category:      req.Category,
		UOM:           req.UOM,
		MinStockLevel: req.MinStockLevel,
		IsStockItem:   isStockItem,
	}

	if err := h.repo.CreatePart(r.Context(), part); err != nil {
		slog.Error("Failed to create part", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to create part")
		return
	}

	jsonResponse(w, http.StatusCreated, part)
}

// UpdatePart handles PUT /parts/{id}
func (h *InventoryHandler) UpdatePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	existing, err := h.repo.FindPartByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrPartNotFound {
			errorResponse(w, http.StatusNotFound, "Part not found")
			return
		}
		slog.Error("Failed to find part", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update part")
		return
	}

	var req CreatePartRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.SKU != "" {
		existing.SKU = req.SKU
	}
	if req.Name != "" {
		existing.Name = req.Name
	}
	if req.Category != nil {
		existing.Category = req.Category
	}
	if req.UOM != "" {
		existing.UOM = req.UOM
	}
	existing.MinStockLevel = req.MinStockLevel
	if req.IsStockItem != nil {
		existing.IsStockItem = *req.IsStockItem
	}

	if err := h.repo.UpdatePart(r.Context(), existing); err != nil {
		slog.Error("Failed to update part", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update part")
		return
	}

	jsonResponse(w, http.StatusOK, existing)
}

// DeletePart handles DELETE /parts/{id}
func (h *InventoryHandler) DeletePart(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	if err := h.repo.DeletePart(r.Context(), id); err != nil {
		if err == repository.ErrPartNotFound {
			errorResponse(w, http.StatusNotFound, "Part not found")
			return
		}
		slog.Error("Failed to delete part", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to delete part")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ==================== INVENTORY STOCK ====================

// ListStock handles GET /inventory/stock
func (h *InventoryHandler) ListStock(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := model.InventoryStockListParams{
		TenantID:   claims.TenantID,
		PartID:     r.URL.Query().Get("partId"),
		LocationID: r.URL.Query().Get("locationId"),
		LowStock:   r.URL.Query().Get("lowStock") == "true",
		Page:       parseIntParam(r, "page", 1),
		Limit:      parseIntParam(r, "limit", 10),
	}

	result, err := h.repo.ListStock(r.Context(), params)
	if err != nil {
		slog.Error("Failed to list stock", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch stock")
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

// UpsertStockRequest for creating/updating stock
type UpsertStockRequest struct {
	PartID         string  `json:"partId"`
	LocationID     string  `json:"locationId"`
	QuantityOnHand float64 `json:"quantityOnHand"`
	BinLabel       *string `json:"binLabel"`
}

// UpsertStock handles POST /inventory/stock
func (h *InventoryHandler) UpsertStock(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req UpsertStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PartID == "" || req.LocationID == "" {
		errorResponse(w, http.StatusBadRequest, "PartID and LocationID are required")
		return
	}

	stock := &model.InventoryStock{
		TenantID:       claims.TenantID,
		PartID:         req.PartID,
		LocationID:     req.LocationID,
		QuantityOnHand: req.QuantityOnHand,
		BinLabel:       req.BinLabel,
	}

	if err := h.repo.UpsertStock(r.Context(), stock); err != nil {
		slog.Error("Failed to upsert stock", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to update stock")
		return
	}

	jsonResponse(w, http.StatusOK, stock)
}

// AdjustStockRequest for adjusting stock quantity
type AdjustStockRequest struct {
	PartID     string  `json:"partId"`
	LocationID string  `json:"locationId"`
	Delta      float64 `json:"delta"` // Positive to add, negative to subtract
}

// AdjustStock handles POST /inventory/stock/adjust
func (h *InventoryHandler) AdjustStock(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req AdjustStockRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		errorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if req.PartID == "" || req.LocationID == "" {
		errorResponse(w, http.StatusBadRequest, "PartID and LocationID are required")
		return
	}

	if err := h.repo.AdjustStock(r.Context(), claims.TenantID, req.PartID, req.LocationID, req.Delta); err != nil {
		if err == repository.ErrStockNotFound {
			errorResponse(w, http.StatusNotFound, "Stock record not found")
			return
		}
		slog.Error("Failed to adjust stock", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to adjust stock")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Stock adjusted successfully"})
}

// ==================== WALLETS ====================

// ListMyWallet handles GET /inventory/wallets (current user)
func (h *InventoryHandler) ListMyWallet(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		errorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	wallets, err := h.repo.ListWallets(r.Context(), claims.UserID)
	if err != nil {
		slog.Error("Failed to list wallets", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch wallet")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"data": wallets})
}

// ListUserWallet handles GET /inventory/wallets/{userId}
func (h *InventoryHandler) ListUserWallet(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "userId")

	wallets, err := h.repo.ListWallets(r.Context(), userID)
	if err != nil {
		slog.Error("Failed to list wallets", slog.String("error", err.Error()))
		errorResponse(w, http.StatusInternalServerError, "Failed to fetch wallet")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]interface{}{"data": wallets})
}
