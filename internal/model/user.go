package model

import (
	"time"
)

// User represents a user in the system
type User struct {
	ID           string    `json:"id"`
	TenantID     string    `json:"tenantId"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Never serialize
	Role         string    `json:"role"`
	FirstName    string    `json:"firstName"`
	LastName     string    `json:"lastName"`
	CreatedAt    time.Time `json:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt"`
}

// UserResponse is the API response for a user (without sensitive data)
type UserResponse struct {
	ID        string    `json:"id"`
	TenantID  string    `json:"tenantId"`
	Email     string    `json:"email"`
	Role      string    `json:"role"`
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	Name      string    `json:"name"` // Computed
	CreatedAt time.Time `json:"createdAt"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() UserResponse {
	return UserResponse{
		ID:        u.ID,
		TenantID:  u.TenantID,
		Email:     u.Email,
		Role:      u.Role,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Name:      u.FirstName + " " + u.LastName,
		CreatedAt: u.CreatedAt,
	}
}
