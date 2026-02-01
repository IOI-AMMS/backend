package model

import (
	"time"
)

// User represents a user in the system (v1.1 schema)
type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	OrgUnitID    *string   `json:"orgUnitId,omitempty"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never serialize
	FullName     string    `json:"fullName"`
	Role         string    `json:"role"`
	IsActive     bool      `json:"isActive"`
	CreatedAt    time.Time `json:"createdAt"`
}

// UserResponse is the API response for a user (without sensitive data)
type UserResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	OrgUnitID *string   `json:"orgUnitId,omitempty"`
	Email     string    `json:"email"`
	FullName  string    `json:"fullName"`
	Role      string    `json:"role"`
	IsActive  bool      `json:"isActive"`
	CreatedAt time.Time `json:"createdAt"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		TenantID:  u.TenantID,
		OrgUnitID: u.OrgUnitID,
		Email:     u.Email,
		FullName:  u.FullName,
		Role:      u.Role,
		IsActive:  u.IsActive,
		CreatedAt: u.CreatedAt,
	}
}

// User Role constants
const (
	RoleAdmin      = "Admin"
	RoleManager    = "Manager"
	RoleSupervisor = "Supervisor"
	RoleTechnician = "Technician"
	RoleStoreman   = "Storeman"
	RoleViewer     = "Viewer"
)
