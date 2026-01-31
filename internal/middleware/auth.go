package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"ioi-amms/internal/auth"
)

type contextKey string

const UserContextKey contextKey = "user"

// Role constants matching PRD Section 4
const (
	RoleTechnician = "technician"
	RoleSupervisor = "supervisor"
	RoleStoreman   = "storeman"
	RoleManager    = "manager"
	RoleAdmin      = "admin"
)

// Permission constants for fine-grained access control
const (
	PermissionAssetRead      = "asset:read"
	PermissionAssetWrite     = "asset:write"
	PermissionAssetDelete    = "asset:delete"
	PermissionWORead         = "wo:read"
	PermissionWOWrite        = "wo:write"
	PermissionWOAssign       = "wo:assign"
	PermissionWOClose        = "wo:close"
	PermissionInventoryRead  = "inventory:read"
	PermissionInventoryWrite = "inventory:write"
	PermissionUserManage     = "user:manage"
	PermissionReportView     = "report:view"
)

// RolePermissions maps roles to their allowed permissions
var RolePermissions = map[string][]string{
	RoleTechnician: {
		PermissionAssetRead,
		PermissionWORead, PermissionWOWrite,
	},
	RoleSupervisor: {
		PermissionAssetRead, PermissionAssetWrite,
		PermissionWORead, PermissionWOWrite, PermissionWOAssign, PermissionWOClose,
		PermissionReportView,
	},
	RoleStoreman: {
		PermissionAssetRead,
		PermissionInventoryRead, PermissionInventoryWrite,
	},
	RoleManager: {
		PermissionAssetRead, PermissionAssetWrite, PermissionAssetDelete,
		PermissionWORead, PermissionWOWrite, PermissionWOAssign, PermissionWOClose,
		PermissionInventoryRead, PermissionInventoryWrite,
		PermissionReportView,
		PermissionUserManage,
	},
	RoleAdmin: {
		PermissionAssetRead, PermissionAssetWrite, PermissionAssetDelete,
		PermissionWORead, PermissionWOWrite, PermissionWOAssign, PermissionWOClose,
		PermissionInventoryRead, PermissionInventoryWrite,
		PermissionReportView,
		PermissionUserManage,
	},
}

// AuthMiddleware validates the JWT token and adds claims to context
func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			errorResponse(w, http.StatusUnauthorized, "Authorization header required")
			return
		}

		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			errorResponse(w, http.StatusUnauthorized, "Invalid authorization format")
			return
		}

		claims, err := auth.ValidateToken(parts[1])
		if err != nil {
			if err == auth.ErrExpiredToken {
				errorResponse(w, http.StatusUnauthorized, "Token expired")
				return
			}
			errorResponse(w, http.StatusUnauthorized, "Invalid token")
			return
		}

		ctx := context.WithValue(r.Context(), UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole returns middleware that checks if user has one of the required roles
func RequireRole(roles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
			if !ok {
				errorResponse(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			for _, role := range roles {
				if claims.Role == role {
					next.ServeHTTP(w, r)
					return
				}
			}

			errorResponse(w, http.StatusForbidden, "Insufficient permissions")
		})
	}
}

// RequirePermission returns middleware that checks if user's role has the required permission
func RequirePermission(permission string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
			if !ok {
				errorResponse(w, http.StatusUnauthorized, "User not authenticated")
				return
			}

			permissions, exists := RolePermissions[claims.Role]
			if !exists {
				errorResponse(w, http.StatusForbidden, "Unknown role")
				return
			}

			for _, p := range permissions {
				if p == permission {
					next.ServeHTTP(w, r)
					return
				}
			}

			errorResponse(w, http.StatusForbidden, "Permission denied: "+permission)
		})
	}
}

// GetUserFromContext extracts user claims from the request context
func GetUserFromContext(r *http.Request) (*auth.Claims, bool) {
	claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
	return claims, ok
}

// HasPermission checks if a role has a specific permission
func HasPermission(role, permission string) bool {
	permissions, exists := RolePermissions[role]
	if !exists {
		return false
	}
	for _, p := range permissions {
		if p == permission {
			return true
		}
	}
	return false
}

func errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{
		"error":   http.StatusText(status),
		"message": message,
	})
}
