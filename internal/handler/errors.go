package handler

import (
	"encoding/json"
	"net/http"
)

// APIError represents a structured error response
type APIError struct {
	Error   string                 `json:"error"`             // HTTP status text
	Code    string                 `json:"code,omitempty"`    // Machine-readable error code
	Message string                 `json:"message"`           // Human-readable description
	Details map[string]interface{} `json:"details,omitempty"` // Additional context
}

// Common error codes for frontend handling
const (
	ErrCodeValidation      = "VALIDATION_ERROR"
	ErrCodeNotFound        = "NOT_FOUND"
	ErrCodeUnauthorized    = "UNAUTHORIZED"
	ErrCodeForbidden       = "FORBIDDEN"
	ErrCodeConflict        = "CONFLICT"
	ErrCodeInternal        = "INTERNAL_ERROR"
	ErrCodeBadRequest      = "BAD_REQUEST"
	ErrCodeRateLimited     = "RATE_LIMITED"
	ErrCodeFileTooLarge    = "FILE_TOO_LARGE"
	ErrCodeInvalidInput    = "INVALID_INPUT"
	ErrCodeDatabaseError   = "DATABASE_ERROR"
	ErrCodeExternalService = "EXTERNAL_SERVICE_ERROR"
)

// Simple error response (backward compatible)
func errorResponse(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Error:   http.StatusText(status),
		Message: message,
	})
}

// Enhanced error response with code and details
func errorResponseWithCode(w http.ResponseWriter, status int, code, message string, details map[string]interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(APIError{
		Error:   http.StatusText(status),
		Code:    code,
		Message: message,
		Details: details,
	})
}

// JSON response helper
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Convenience functions for common errors

func badRequest(w http.ResponseWriter, message string, details map[string]interface{}) {
	errorResponseWithCode(w, http.StatusBadRequest, ErrCodeBadRequest, message, details)
}

func validationError(w http.ResponseWriter, message string, fields map[string]string) {
	details := make(map[string]interface{})
	for k, v := range fields {
		details[k] = v
	}
	errorResponseWithCode(w, http.StatusBadRequest, ErrCodeValidation, message, details)
}

func notFoundError(w http.ResponseWriter, resource string) {
	errorResponseWithCode(w, http.StatusNotFound, ErrCodeNotFound,
		resource+" not found",
		map[string]interface{}{"resource": resource})
}

func unauthorizedError(w http.ResponseWriter, message string) {
	errorResponseWithCode(w, http.StatusUnauthorized, ErrCodeUnauthorized, message, nil)
}

func forbiddenError(w http.ResponseWriter, message string) {
	errorResponseWithCode(w, http.StatusForbidden, ErrCodeForbidden, message, nil)
}

func conflictError(w http.ResponseWriter, message string, details map[string]interface{}) {
	errorResponseWithCode(w, http.StatusConflict, ErrCodeConflict, message, details)
}

func internalError(w http.ResponseWriter, message string) {
	errorResponseWithCode(w, http.StatusInternalServerError, ErrCodeInternal, message, nil)
}

func databaseError(w http.ResponseWriter, operation string) {
	errorResponseWithCode(w, http.StatusInternalServerError, ErrCodeDatabaseError,
		"Database operation failed",
		map[string]interface{}{"operation": operation})
}
