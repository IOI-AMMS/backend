package model

import (
	"time"
)

// Asset represents a physical asset in the system
// Schema aligned with existing database structure
type Asset struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenantId"`
	ParentID       *string    `json:"parentId"`
	LocationID     *string    `json:"locationId"`
	Name           string     `json:"name"`
	Status         string     `json:"status"`
	Criticality    string     `json:"criticality"`
	LastInspection *time.Time `json:"lastInspection"`
	CreatedAt      time.Time  `json:"createdAt"`
	UpdatedAt      time.Time  `json:"updatedAt"`

	// Computed/Joined fields
	LocationName string `json:"location,omitempty"`
}

// AssetStatus enum values
const (
	AssetStatusOperational    = "operational"
	AssetStatusMaintenance    = "maintenance"
	AssetStatusDecommissioned = "decommissioned"
	AssetStatusPending        = "pending"
)

// AssetCriticality enum values
const (
	AssetCriticalityHigh   = "high"
	AssetCriticalityMedium = "medium"
	AssetCriticalityLow    = "low"
)

// AssetListParams for filtering and pagination
type AssetListParams struct {
	TenantID    string
	Status      []string
	Criticality []string
	Search      string
	SortBy      string
	SortDir     string
	Page        int
	Limit       int
}

// PaginatedResult wraps paginated data
type PaginatedResult[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}
