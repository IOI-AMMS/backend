package model

import (
	"time"
)

// WorkOrder represents a maintenance work order (v1.1 schema)
type WorkOrder struct {
	ID             string     `json:"id"`
	TenantID       string     `json:"tenantId"`
	ReadableID     *int       `json:"readableId,omitempty"`
	AssetID        *string    `json:"assetId,omitempty"`
	AssignedUserID *string    `json:"assignedUserId,omitempty"`
	Status         string     `json:"status"`
	Origin         string     `json:"origin"`
	Priority       string     `json:"priority"`
	Title          string     `json:"title"`
	Description    *string    `json:"description,omitempty"`
	StartedAt      *time.Time `json:"startedAt,omitempty"`
	CompletedAt    *time.Time `json:"completedAt,omitempty"`
	CreatedAt      time.Time  `json:"createdAt"`

	// Computed/Joined fields
	AssetName string `json:"assetName,omitempty"`
}

// WorkOrder Status constants (v1.1 schema)
const (
	WOStatusRequested    = "Requested"
	WOStatusApproved     = "Approved"
	WOStatusInProgress   = "In_Progress"
	WOStatusWorkComplete = "Work_Complete"
	WOStatusClosed       = "Closed"
	WOStatusCancelled    = "Cancelled"
)

// WorkOrder Priority constants
const (
	WOPriorityLow      = "Low"
	WOPriorityMedium   = "Medium"
	WOPriorityHigh     = "High"
	WOPriorityCritical = "Critical"
)

// WorkOrder Origin constants (v1.1 schema)
const (
	WOOriginPreventiveAuto = "Preventive_Auto"
	WOOriginManualRequest  = "Manual_Request"
	WOOriginDefectFollowup = "Defect_Followup"
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
