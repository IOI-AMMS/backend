package handler

import (
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"ioi-amms/internal/config"
	"ioi-amms/internal/middleware"
	"ioi-amms/internal/storage"

	"github.com/go-chi/chi/v5"
)

const (
	maxUploadSize = 10 << 20 // 10 MB
	bucketName    = "attachments"
)

// FileHandler handles file upload/download requests
type FileHandler struct {
	storage storage.Service
	cfg     *config.Config
}

// NewFileHandler creates a new file handler
func NewFileHandler(storage storage.Service, cfg *config.Config) *FileHandler {
	return &FileHandler{
		storage: storage,
		cfg:     cfg,
	}
}

// RegisterRoutes registers file routes
func (h *FileHandler) RegisterRoutes(r chi.Router) {
	r.Post("/upload", h.Upload)
	r.Get("/files/{bucket}/{*objectPath}", h.Download)
	r.Delete("/files/{bucket}/{*objectPath}", h.Delete)
}

// Upload handles POST /upload
func (h *FileHandler) Upload(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		unauthorizedError(w, "User not authenticated")
		return
	}

	// Limit upload size
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)

	// Parse multipart form
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		if strings.Contains(err.Error(), "request body too large") {
			errorResponseWithCode(w, http.StatusRequestEntityTooLarge, ErrCodeFileTooLarge,
				fmt.Sprintf("File exceeds maximum size of %d MB", maxUploadSize>>20),
				map[string]interface{}{"maxSizeBytes": maxUploadSize})
			return
		}
		badRequest(w, "Invalid multipart form data", map[string]interface{}{"details": err.Error()})
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		badRequest(w, "No file provided in request", map[string]interface{}{"field": "file"})
		return
	}
	defer file.Close()

	// Get associated asset ID (optional)
	assetID := r.FormValue("assetId")
	if assetID == "" {
		assetID = "general"
	}

	// Generate unique object name
	objectName := storage.GenerateObjectName(claims.TenantID, assetID, header.Filename)
	contentType := storage.GetContentType(header.Filename)

	// Upload to MinIO
	result, err := h.storage.Upload(r.Context(), bucketName, objectName, file, header.Size, contentType)
	if err != nil {
		slog.Error("Failed to upload file",
			slog.String("error", err.Error()),
			slog.String("filename", header.Filename),
			slog.String("tenantId", claims.TenantID))
		internalError(w, "Failed to upload file to storage")
		return
	}

	slog.Info("File uploaded",
		slog.String("objectName", result.ObjectName),
		slog.Int64("size", result.Size),
		slog.String("tenantId", claims.TenantID))

	jsonResponse(w, http.StatusCreated, map[string]interface{}{
		"success":      true,
		"objectName":   result.ObjectName,
		"size":         result.Size,
		"contentType":  contentType,
		"originalName": header.Filename,
		"downloadUrl":  fmt.Sprintf("/api/v1/files/%s/%s", result.Bucket, result.ObjectName),
	})
}

// Download handles GET /files/{bucket}/{objectPath}
func (h *FileHandler) Download(w http.ResponseWriter, r *http.Request) {
	bucket := chi.URLParam(r, "bucket")
	objectPath := chi.URLParam(r, "objectPath")

	if bucket == "" || objectPath == "" {
		badRequest(w, "Bucket and object path are required", nil)
		return
	}

	// Get presigned URL for download (valid for 1 hour)
	url, err := h.storage.GetPresignedURL(r.Context(), bucket, objectPath, time.Hour)
	if err != nil {
		slog.Error("Failed to generate presigned URL",
			slog.String("error", err.Error()),
			slog.String("object", objectPath))
		internalError(w, "Failed to generate download URL")
		return
	}

	// Redirect to presigned URL
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

// Delete handles DELETE /files/{bucket}/{objectPath}
func (h *FileHandler) Delete(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetUserFromContext(r)
	if !ok {
		unauthorizedError(w, "User not authenticated")
		return
	}

	bucket := chi.URLParam(r, "bucket")
	objectPath := chi.URLParam(r, "objectPath")

	// Verify tenant owns this file (path should start with their tenant ID)
	if !strings.HasPrefix(objectPath, "tenants/"+claims.TenantID+"/") {
		forbiddenError(w, "You can only delete files owned by your tenant")
		return
	}

	if err := h.storage.Delete(r.Context(), bucket, objectPath); err != nil {
		slog.Error("Failed to delete file",
			slog.String("error", err.Error()),
			slog.String("object", objectPath))
		internalError(w, "Failed to delete file")
		return
	}

	slog.Info("File deleted",
		slog.String("object", objectPath),
		slog.String("tenantId", claims.TenantID))

	w.WriteHeader(http.StatusNoContent)
}
