package model

import (
	"time"
)

// WorkOrder represents a maintenance work order
// Schema aligned with existing database structure
type WorkOrder struct {
	ID          string    `json:"id"`
	TenantID    string    `json:"tenantId"`
	AssetID     *string   `json:"assetId"`
	Status      string    `json:"status"`
	Origin      string    `json:"origin"`
	Priority    string    `json:"priority"`
	Description *string   `json:"description"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Computed/Joined fields
	AssetName string `json:"assetName,omitempty"`
}

// WorkOrder Status enum values
const (
	WOStatusDraft      = "Draft"
	WOStatusReady      = "Ready"
	WOStatusInProgress = "In_Progress"
	WOStatusClosed     = "Closed"
)

// WorkOrder Priority enum values
const (
	WOPriorityLow      = "Low"
	WOPriorityMedium   = "Medium"
	WOPriorityHigh     = "High"
	WOPriorityCritical = "Critical"
)

// WorkOrder Origin enum values
const (
	WOOriginPM     = "PM"
	WOOriginCM     = "CM"
	WOOriginDefect = "Defect"
)

// WorkOrderListParams for filtering and pagination
type WorkOrderListParams struct {
	TenantID string
	Status   []string
	Priority []string
	AssetID  string
	SortBy   string
	SortDir  string
	Page     int
	Limit    int
}
