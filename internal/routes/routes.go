package routes

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"ioi-amms/internal/config"
	"ioi-amms/internal/database"
	"ioi-amms/internal/handler"
	"ioi-amms/internal/middleware"
	"ioi-amms/internal/repository"
	"ioi-amms/internal/service"
	"ioi-amms/internal/storage"

	"github.com/go-chi/chi/v5"
)

func NewRouter(db database.Service, cfg *config.Config) http.Handler {
	r := chi.NewRouter()

	// Initialize rate limiter
	rateLimiter := middleware.DefaultRateLimiter()

	// Global middleware
	r.Use(middleware.CORS)
	r.Use(middleware.Logger)
	r.Use(middleware.RateLimit(rateLimiter))

	// Health and system routes (public, no auth)
	RegisterHealthRoutes(r, db)
	r.Get("/", HelloWorldHandler)

	// Initialize repositories
	assetRepo := repository.NewAssetRepository(db.Pool())
	woRepo := repository.NewWorkOrderRepository(db.Pool())
	locationRepo := repository.NewLocationRepository(db.Pool())
	userRepo := repository.NewUserRepository(db.Pool()) // Already used below, lifting up
	tenantRepo := repository.NewTenantRepository(db.Pool())
	auditRepo := repository.NewAuditRepository(db.Pool())
	inventoryRepo := repository.NewInventoryRepository(db.Pool())
	analyticsRepo := repository.NewAnalyticsRepository(db.Pool()) // [NEW]

	// Initialize services
	// Initialize services
	// Initialize services
	auditService := service.NewAuditService(auditRepo)

	// Initialize storage service (moved up for dependency)
	var fileHandler *handler.FileHandler
	storageService, err := storage.NewService(cfg)
	if err != nil {
		slog.Warn("MinIO storage unavailable, file uploads disabled", slog.String("error", err.Error()))
	} else {
		fileHandler = handler.NewFileHandler(storageService, cfg)
	}

	// Initialize handlers
	assetHandler := handler.NewAssetHandler(assetRepo, auditService)
	woHandler := handler.NewWorkOrderHandler(woRepo)
	locationHandler := handler.NewLocationHandler(locationRepo)
	userHandler := handler.NewUserHandler(userRepo)
	tenantHandler := handler.NewTenantHandler(tenantRepo)
	auditHandler := handler.NewAuditHandler(auditRepo)
	inventoryHandler := handler.NewInventoryHandler(inventoryRepo)
	analyticsHandler := handler.NewAnalyticsHandler(analyticsRepo)
	systemHandler := SystemHealthHandler(db, storageService)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		RegisterAuthRoutes(r, db)

		// Protected routes (require authentication + tenant isolation)
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware)
			r.Use(middleware.TenantIsolation)

			// Asset routes
			assetHandler.RegisterRoutes(r)

			// Work Order routes
			woHandler.RegisterRoutes(r)

			// Location routes
			locationHandler.RegisterRoutes(r)

			// User routes
			userHandler.RegisterRoutes(r)

			// Inventory routes (Parts, Stock, Wallets)
			inventoryHandler.RegisterRoutes(r)

			// Analytics Dashboard (New)
			analyticsHandler.RegisterRoutes(r)

			// File routes (if storage available)
			if fileHandler != nil {
				fileHandler.RegisterRoutes(r)
			}

			// Admin Routes
			r.Group(func(r chi.Router) {
				// Tenant Settings (Admin)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequirePermission(middleware.PermissionTenantSettings))
					tenantHandler.RegisterRoutes(r)
				})

				// Audit Logs (Admin)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequirePermission(middleware.PermissionAuditRead))
					r.Get("/audit-logs", auditHandler.List)
				})

				// System Health (Admin)
				r.Group(func(r chi.Router) {
					r.Use(middleware.RequirePermission(middleware.PermissionSystemHealth))
					r.Get("/system/health", systemHandler)
				})
			})
		})
	})

	return r
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "IOI AMMS API"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
