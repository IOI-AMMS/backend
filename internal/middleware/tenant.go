package middleware

import (
	"log/slog"
	"net/http"

	"ioi-amms/internal/auth"
)

// TenantIsolation middleware ensures users can only access resources within their tenant
// This is automatically enforced in handlers by using claims.TenantID in queries,
// but this middleware provides an additional safety check for explicit tenant ID params
func TenantIsolation(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, ok := r.Context().Value(UserContextKey).(*auth.Claims)
		if !ok {
			errorResponse(w, http.StatusUnauthorized, "User not authenticated")
			return
		}

		// Check if request contains a tenant_id query param that doesn't match user's tenant
		requestedTenant := r.URL.Query().Get("tenant_id")
		if requestedTenant != "" && requestedTenant != claims.TenantID {
			slog.Warn("Tenant isolation violation attempt",
				slog.String("user_id", claims.UserID),
				slog.String("user_tenant", claims.TenantID),
				slog.String("requested_tenant", requestedTenant),
			)
			errorResponse(w, http.StatusForbidden, "Access denied: cross-tenant access not allowed")
			return
		}

		next.ServeHTTP(w, r)
	})
}

// EnforceTenantID is a helper to validate tenant ownership of a resource
// Call this in handlers before performing operations on resources
func EnforceTenantID(userTenantID, resourceTenantID string) bool {
	return userTenantID == resourceTenantID
}
