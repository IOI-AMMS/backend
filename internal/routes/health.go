package routes

import (
	"encoding/json"
	"net/http"
	"runtime"
	"time"

	"ioi-amms/internal/database"

	"github.com/go-chi/chi/v5"
)

var startTime = time.Now()

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string `json:"status"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// ReadinessResponse includes database connectivity
type ReadinessResponse struct {
	Status   string            `json:"status"`
	Checks   map[string]string `json:"checks"`
	Duration string            `json:"duration"`
}

// LivenessResponse for Kubernetes liveness probe
type LivenessResponse struct {
	Status string `json:"status"`
	Uptime string `json:"uptime"`
}

// MetricsResponse provides basic application metrics
type MetricsResponse struct {
	Uptime       string `json:"uptime"`
	GoVersion    string `json:"goVersion"`
	NumGoroutine int    `json:"numGoroutine"`
	NumCPU       int    `json:"numCPU"`
	MemoryAlloc  uint64 `json:"memoryAllocBytes"`
	MemorySys    uint64 `json:"memorySysBytes"`
	HeapAlloc    uint64 `json:"heapAllocBytes"`
	HeapSys      uint64 `json:"heapSysBytes"`
	GCPauseNs    uint64 `json:"gcPauseNs"`
	NumGC        uint32 `json:"numGC"`
}

// RegisterHealthRoutes registers health and readiness endpoints
func RegisterHealthRoutes(r chi.Router, db database.Service) {
	r.Get("/health", HealthHandler())
	r.Get("/ready", ReadinessHandler(db))
	r.Get("/live", LivenessHandler())
	r.Get("/metrics", MetricsHandler())
}

// HealthHandler returns basic health status
func HealthHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(HealthResponse{
			Status:    "up",
			Message:   "It's healthy",
			Timestamp: time.Now().UTC().Format(time.RFC3339),
		})
	}
}

// ReadinessHandler checks if the app is ready to serve traffic
func ReadinessHandler(db database.Service) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		checks := make(map[string]string)
		allHealthy := true

		// Check database connectivity
		health := db.Health()
		if health["status"] == "up" {
			checks["database"] = "ok"
		} else {
			checks["database"] = "failed"
			allHealthy = false
		}

		w.Header().Set("Content-Type", "application/json")

		status := "ready"
		httpStatus := http.StatusOK
		if !allHealthy {
			status = "not_ready"
			httpStatus = http.StatusServiceUnavailable
		}

		w.WriteHeader(httpStatus)
		json.NewEncoder(w).Encode(ReadinessResponse{
			Status:   status,
			Checks:   checks,
			Duration: time.Since(start).String(),
		})
	}
}

// LivenessHandler checks if the app is alive (for K8s)
func LivenessHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(LivenessResponse{
			Status: "alive",
			Uptime: time.Since(startTime).String(),
		})
	}
}

// MetricsHandler returns application metrics
func MetricsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var mem runtime.MemStats
		runtime.ReadMemStats(&mem)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(MetricsResponse{
			Uptime:       time.Since(startTime).String(),
			GoVersion:    runtime.Version(),
			NumGoroutine: runtime.NumGoroutine(),
			NumCPU:       runtime.NumCPU(),
			MemoryAlloc:  mem.Alloc,
			MemorySys:    mem.Sys,
			HeapAlloc:    mem.HeapAlloc,
			HeapSys:      mem.HeapSys,
			GCPauseNs:    mem.PauseNs[(mem.NumGC+255)%256],
			NumGC:        mem.NumGC,
		})
	}
}
