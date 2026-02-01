package handler

import (
	"encoding/json"
	"net/http"

	"ioi-amms/internal/auth"
	"ioi-amms/internal/middleware"
	"ioi-amms/internal/model"
	"ioi-amms/internal/repository"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	repo *repository.UserRepository
}

func NewUserHandler(repo *repository.UserRepository) *UserHandler {
	return &UserHandler{repo: repo}
}

func (h *UserHandler) RegisterRoutes(r chi.Router) {
	r.Route("/users", func(r chi.Router) {
		r.Get("/", h.ListUsers)
		r.Post("/", h.CreateUser)
		r.Get("/", h.ListUsers)
		r.Post("/", h.CreateUser)
		r.Put("/{id}", h.UpdateUser) // New endpoint for Settings screen
		r.Put("/{id}/password", h.ResetPassword)
	})
}

// UpdateUser handles updating user details (Role, Status, FullName)
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		badRequest(w, "User ID is required", nil)
		return
	}

	var req struct {
		FullName  string  `json:"fullName"`
		Role      string  `json:"role"`
		IsActive  bool    `json:"isActive"`
		OrgUnitID *string `json:"orgUnitId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "Invalid request body", nil)
		return
	}

	// Fetch existing user to ensure tenant isolation
	claims, _ := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	existingUser, err := h.repo.FindByID(r.Context(), userID)
	if err != nil {
		if err == repository.ErrUserNotFound {
			notFoundError(w, "User not found")
			return
		}
		internalError(w, err.Error())
		return
	}

	// Ensure tenant match
	if existingUser.TenantID != claims.TenantID {
		notFoundError(w, "User not found") // Hide cross-tenant existence
		return
	}

	// Update fields
	if req.FullName != "" {
		existingUser.FullName = req.FullName
	}
	if req.Role != "" {
		existingUser.Role = req.Role
	}
	existingUser.IsActive = req.IsActive
	existingUser.OrgUnitID = req.OrgUnitID // Can be nil

	if err := h.repo.Update(r.Context(), existingUser); err != nil {
		internalError(w, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, existingUser.ToResponse())
}

// CreateUser handles creating a new user (v1.1 schema)
func (h *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Email     string  `json:"email"`
		Password  string  `json:"password"`
		FullName  string  `json:"fullName"`
		Role      string  `json:"role"`
		OrgUnitID *string `json:"orgUnitId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "Invalid request body", nil)
		return
	}

	if req.Email == "" || req.Password == "" || req.Role == "" {
		badRequest(w, "Email, password, and role are required", nil)
		return
	}

	// Hash password
	hashedPassword, err := auth.HashPassword(req.Password)
	if err != nil {
		internalError(w, "Failed to hash password")
		return
	}

	// Get tenant from context via user claims
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		unauthorizedError(w, "User not authenticated")
		return
	}
	tenantID := claims.TenantID

	user := &model.User{
		TenantID:     tenantID,
		OrgUnitID:    req.OrgUnitID,
		Email:        req.Email,
		PasswordHash: hashedPassword,
		FullName:     req.FullName,
		Role:         req.Role,
		IsActive:     true,
	}

	if err := h.repo.Create(r.Context(), user); err != nil {
		if err == repository.ErrUserExists {
			conflictError(w, "User with this email already exists", nil)
			return
		}
		internalError(w, err.Error())
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(user.ToResponse())
}

// ListUsers retrieves all users for the current tenant (v1.1 schema)
func (h *UserHandler) ListUsers(w http.ResponseWriter, r *http.Request) {
	// Get tenant from context via user claims
	claims, ok := r.Context().Value(middleware.UserContextKey).(*auth.Claims)
	if !ok {
		unauthorizedError(w, "User not authenticated")
		return
	}
	tenantID := claims.TenantID

	users, err := h.repo.List(r.Context(), tenantID)
	if err != nil {
		internalError(w, err.Error())
		return
	}

	// Convert to response format
	var response []model.UserResponse
	for _, u := range users {
		response = append(response, u.ToResponse())
	}

	// Return empty list if no users found instead of null
	if response == nil {
		response = []model.UserResponse{}
	}

	jsonResponse(w, http.StatusOK, response)
}

// ResetPassword handles password reset for a user
func (h *UserHandler) ResetPassword(w http.ResponseWriter, r *http.Request) {
	userID := chi.URLParam(r, "id")
	if userID == "" {
		badRequest(w, "User ID is required", nil)
		return
	}

	var req struct {
		NewPassword string `json:"newPassword"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		badRequest(w, "Invalid request body", nil)
		return
	}

	if req.NewPassword == "" {
		badRequest(w, "New password is required", nil)
		return
	}

	// Hash new password
	hashedPassword, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		internalError(w, "Failed to hash password")
		return
	}

	if err := h.repo.UpdatePassword(r.Context(), userID, hashedPassword); err != nil {
		internalError(w, err.Error())
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"message": "Password updated successfully"})
}
