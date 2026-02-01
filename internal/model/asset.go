package model

import (
	"encoding/json"
	"time"
)

// Asset represents a physical asset in the system (v1.1 schema)
type Asset struct {
	ID              string          `json:"id"`
	TenantID        string          `json:"tenantId"`
	ParentID        *string         `json:"parentId,omitempty"`
	LocationID      *string         `json:"locationId,omitempty"`
	OrgUnitID       *string         `json:"orgUnitId,omitempty"`
	Name            string          `json:"name"`
	Status          string          `json:"status"`
	IsFieldRelated  bool            `json:"isFieldRelated"`
	IsFieldVerified bool            `json:"isFieldVerified"`
	Manufacturer    *string         `json:"manufacturer,omitempty"`
	ModelNumber     *string         `json:"modelNumber,omitempty"`
	Specs           json.RawMessage `json:"specs,omitempty"`
	CreatedAt       time.Time       `json:"createdAt"`
	UpdatedAt       time.Time       `json:"updatedAt"`

	// Computed/Joined fields
	LocationName string `json:"location,omitempty"`
	OrgUnitName  string `json:"orgUnit,omitempty"`
}

// AssetStatus constants (v1.1)
const (
	AssetStatusDraft    = "Draft"
	AssetStatusActive   = "Active"
	AssetStatusDown     = "Down"
	AssetStatusArchived = "Archived"
	AssetStatusRedTag   = "Red_Tag"
)

// AssetListParams for filtering and pagination
type AssetListParams struct {
	TenantID  string
	Status    []string
	OrgUnitID string
	Search    string
	SortBy    string
	SortDir   string
	Page      int
	Limit     int
}

// PaginatedResult wraps paginated data
type PaginatedResult[T any] struct {
	Data       []T `json:"data"`
	Total      int `json:"total"`
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	TotalPages int `json:"totalPages"`
}

// AssetFinance represents sensitive financial data (1:1 with Asset)
type AssetFinance struct {
	AssetID             string     `json:"assetId"`
	PurchaseDate        *time.Time `json:"purchaseDate,omitempty"`
	PlacedInServiceDate *time.Time `json:"placedInServiceDate,omitempty"`
	WarrantyExpiryDate  *time.Time `json:"warrantyExpiryDate,omitempty"`
	Currency            string     `json:"currency"`
	AcquisitionCost     *float64   `json:"acquisitionCost,omitempty"`
	CurrentBookValue    *float64   `json:"currentBookValue,omitempty"`
	TotalDepreciation   *float64   `json:"totalDepreciation,omitempty"`
	VendorName          *string    `json:"vendorName,omitempty"`
	ErpFixedAssetID     *string    `json:"erpFixedAssetId,omitempty"`
	ErpStatus           *string    `json:"erpStatus,omitempty"`
	LastSyncedAt        *time.Time `json:"lastSyncedAt,omitempty"`
}

// AssetIdentity represents a searchable identifier for an asset
type AssetIdentity struct {
	AssetID   string `json:"assetId"`
	Type      string `json:"type"` // Client_Code, QR_Token, Serial_Number, Barcode
	Value     string `json:"value"`
	IsPrimary bool   `json:"isPrimary"`
}

// AssetMeter tracks operational stats for an asset
type AssetMeter struct {
	AssetID           string     `json:"assetId"`
	CurrentRunHours   float64    `json:"currentRunHours"`
	CurrentOdometerKm float64    `json:"currentOdometerKm"`
	LastUpdatedAt     *time.Time `json:"lastUpdatedAt,omitempty"`
	UpdatedByUserID   *string    `json:"updatedByUserId,omitempty"`
}
