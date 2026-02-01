package repository

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrTenantNotFound = errors.New("tenant not found")
)

// TenantRepository handles tenant data access
type TenantRepository struct {
	db *pgxpool.Pool
}

// NewTenantRepository creates a new tenant repository
func NewTenantRepository(db *pgxpool.Pool) *TenantRepository {
	return &TenantRepository{db: db}
}

// CategoryDefault defines default values for a specific asset category
type CategoryDefault struct {
	MaintenanceCycleDays int    `json:"maintenanceCycleDays"`
	WarrantyPeriodMonths int    `json:"warrantyPeriodMonths"`
	DepreciationMethod   string `json:"depreciationMethod"` // e.g. "Straight Line"
}

// TenantSettings represents the JSONB settings structure
type TenantSettings struct {
	// Global Security Policies
	SessionTimeoutMinutes int  `json:"sessionTimeoutMinutes"`
	RequireMFA            bool `json:"requireMFA"`
	PasswordExpiryDays    int  `json:"passwordExpiryDays"`

	// API Integration (Client ERP)
	ERPConfig struct {
		Enabled      bool   `json:"enabled"`
		Provider     string `json:"provider"`
		BaseURL      string `json:"baseURL"`
		SyncSchedule string `json:"syncSchedule"`
	} `json:"erpConfig"`

	// Dynamic Dropdown Lists
	AssetCategories []string `json:"assetCategories"` // ["Hardware", "Vehicles", "HVAC"]
	Departments     []string `json:"departments"`     // ["IT", "Operations", "Main Plant"]

	// Per-Category Defaults
	CategoryDefaults map[string]CategoryDefault `json:"categoryDefaults"`
}

// Tenant represents the tenant model subset for settings
type Tenant struct {
	ID       string                 `json:"id"`
	Name     string                 `json:"name"`
	Settings map[string]interface{} `json:"settings"`
}

// GetSettings retrieves tenant settings
func (r *TenantRepository) GetSettings(ctx context.Context, id string) (*Tenant, error) {
	query := `
		SELECT id, name, settings
		FROM tenants
		WHERE id = $1
	`

	var t Tenant
	var settingsJSON []byte

	err := r.db.QueryRow(ctx, query, id).Scan(&t.ID, &t.Name, &settingsJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	if len(settingsJSON) > 0 {
		if err := json.Unmarshal(settingsJSON, &t.Settings); err != nil {
			return nil, err
		}
	} else {
		t.Settings = make(map[string]interface{})
	}

	return &t, nil
}

// UpdateSettings updates tenant settings
func (r *TenantRepository) UpdateSettings(ctx context.Context, id string, settings map[string]interface{}) (*Tenant, error) {
	settingsJSON, err := json.Marshal(settings)
	if err != nil {
		return nil, err
	}

	query := `
		UPDATE tenants
		SET settings = settings || $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, name, settings
	`

	var t Tenant
	var updatedSettingsJSON []byte

	err = r.db.QueryRow(ctx, query, id, settingsJSON).Scan(&t.ID, &t.Name, &updatedSettingsJSON)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTenantNotFound
		}
		return nil, err
	}

	if len(updatedSettingsJSON) > 0 {
		if err := json.Unmarshal(updatedSettingsJSON, &t.Settings); err != nil {
			return nil, err
		}
	}

	return &t, nil
}
