package model

import (
	"time"
)

// Location represents a hierarchical location in the system (v1.1 schema)
type Location struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	ParentID  *string   `json:"parentId,omitempty"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`

	// Computed fields
	Children []Location `json:"children,omitempty"`
	Path     string     `json:"path,omitempty"`
}

// LocationType constants
const (
	LocationTypeSite     = "Site"
	LocationTypeBuilding = "Building"
	LocationTypeRoom     = "Room"
	LocationTypeZone     = "Zone"
)
