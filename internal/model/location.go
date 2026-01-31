package model

import (
	"time"
)

// Location represents a hierarchical location in the system
type Location struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	ParentID  *string   `json:"parentId"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// Computed fields
	Children []Location `json:"children,omitempty"`
	Path     string     `json:"path,omitempty"`
}

// LocationType enum values
const (
	LocationTypeSite     = "Site"
	LocationTypeBuilding = "Building"
	LocationTypeRoom     = "Room"
	LocationTypeZone     = "Zone"
)
