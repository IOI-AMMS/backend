package routes

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"

	"ioi-amms/internal/auth"
	"ioi-amms/internal/database"
	"ioi-amms/internal/repository"

	"github.com/go-chi/chi/v5"
)

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token        string   `json:"token"`
	RefreshToken string   `json:"refreshToken"`
	User         UserInfo `json:"user"`
}

type UserInfo struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	TenantID string `json:"tenantId"`
}

func RegisterAuthRoutes(r chi.Router, db database.Service) {
	userRepo := repository.NewUserRepository(db.Pool())

	r.Post("/auth/login", LoginHandler(userRepo))
	r.Post("/auth/refresh", RefreshHandler(userRepo))
}

func LoginHandler(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		if req.Email == "" || req.Password == "" {
			errorJSON(w, http.StatusBadRequest, "Email and password required")
			return
		}

		ctx := context.Background()

		// Look up user from database
		user, err := userRepo.FindByEmail(ctx, req.Email)
		if err != nil {
			if err == repository.ErrUserNotFound {
				// Fallback to dev user for initial setup
				if req.Email == "admin@ioi.com" && req.Password == "password123" {
					handleDevLogin(w)
					return
				}
				errorJSON(w, http.StatusUnauthorized, "Invalid credentials")
				return
			}
			slog.Error("Database error during login", slog.String("error", err.Error()))
			errorJSON(w, http.StatusInternalServerError, "Internal server error")
			return
		}

		// Verify password
		if !auth.CheckPassword(req.Password, user.PasswordHash) {
			errorJSON(w, http.StatusUnauthorized, "Invalid credentials")
			return
		}

		token, err := auth.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		refreshToken, err := auth.GenerateRefreshToken(user.ID)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, "Failed to generate refresh token")
			return
		}

		resp := LoginResponse{
			Token:        token,
			RefreshToken: refreshToken,
			User: UserInfo{
				ID:       user.ID,
				Email:    user.Email,
				Name:     user.FirstName + " " + user.LastName,
				Role:     user.Role,
				TenantID: user.TenantID,
			},
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// handleDevLogin provides a fallback for development when no users exist in DB
func handleDevLogin(w http.ResponseWriter) {
	// Use valid UUIDs that match seed data (or generate new ones for dev)
	devUserID := "00000000-0000-0000-0000-000000000002"
	devTenantID := "00000000-0000-0000-0000-000000000001"

	token, _ := auth.GenerateToken(devUserID, devTenantID, "admin@ioi.com", "manager")
	refreshToken, _ := auth.GenerateRefreshToken(devUserID)

	resp := LoginResponse{
		Token:        token,
		RefreshToken: refreshToken,
		User: UserInfo{
			ID:       devUserID,
			Email:    "admin@ioi.com",
			Name:     "Dev Admin",
			Role:     "manager",
			TenantID: devTenantID,
		},
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func RefreshHandler(userRepo *repository.UserRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			RefreshToken string `json:"refreshToken"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			errorJSON(w, http.StatusBadRequest, "Invalid request body")
			return
		}

		claims, err := auth.ValidateToken(req.RefreshToken)
		if err != nil {
			errorJSON(w, http.StatusUnauthorized, "Invalid refresh token")
			return
		}

		ctx := context.Background()

		// Look up user to get current role/tenant
		user, err := userRepo.FindByID(ctx, claims.UserID)
		if err != nil {
			// Fallback for dev user
			if claims.UserID == "00000000-0000-0000-0000-000000000002" {
				newToken, _ := auth.GenerateToken("00000000-0000-0000-0000-000000000002", "00000000-0000-0000-0000-000000000001", "admin@ioi.com", "manager")
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(map[string]string{"token": newToken})
				return
			}
			errorJSON(w, http.StatusUnauthorized, "User not found")
			return
		}

		newToken, err := auth.GenerateToken(user.ID, user.TenantID, user.Email, user.Role)
		if err != nil {
			errorJSON(w, http.StatusInternalServerError, "Failed to generate token")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"token": newToken,
		})
	}
}

func errorJSON(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	})
}
