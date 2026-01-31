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

	// Initialize handlers
	assetHandler := handler.NewAssetHandler(assetRepo)
	woHandler := handler.NewWorkOrderHandler(woRepo)
	locationHandler := handler.NewLocationHandler(locationRepo)

	// Initialize storage service (optional, log warning if unavailable)
	var fileHandler *handler.FileHandler
	storageService, err := storage.NewService(cfg)
	if err != nil {
		slog.Warn("MinIO storage unavailable, file uploads disabled", slog.String("error", err.Error()))
	} else {
		fileHandler = handler.NewFileHandler(storageService, cfg)
	}

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

			// File routes (if storage available)
			if fileHandler != nil {
				fileHandler.RegisterRoutes(r)
			}
		})
	})

	return r
}

func HelloWorldHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"message": "IOI AMMS API"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
