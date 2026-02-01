package model

import (
	"time"
)

// AuditLog represents an audit trail entry
type AuditLog struct {
	ID         string                 `json:"id"`
	UserID     string                 `json:"userId"`
	UserName   string                 `json:"userName,omitempty"` // Joined field
	Action     string                 `json:"action"`
	EntityType string                 `json:"entityType"`
	EntityID   string                 `json:"entityId"`
	EntityName string                 `json:"entityName,omitempty"` // Joined field
	Changes    map[string]interface{} `json:"changes,omitempty"`
	CreatedAt  time.Time              `json:"createdAt"`
}

// AuditListParams for filtering logs
type AuditListParams struct {
	TenantID   string
	UserID     string
	EntityType string
	EntityID   string
	Action     string
	From       time.Time
	To         time.Time
	Page       int
	Limit      int
}

// Common Audit Actions
const (
	AuditActionCreate       = "create"
	AuditActionUpdate       = "update"
	AuditActionDelete       = "delete"
	AuditActionStatusChange = "status_change"
	AuditActionLogin        = "login"
	AuditActionAssign       = "assign"
)

// Common Entity Types
const (
	AuditEntityAsset     = "asset"
	AuditEntityWorkOrder = "work_order"
	AuditEntityUser      = "user"
	AuditEntityLocation  = "location"
	AuditEntityTenant    = "tenant"
)
