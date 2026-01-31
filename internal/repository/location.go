package repository

import (
	"context"
	"errors"

	"ioi-amms/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrLocationNotFound = errors.New("location not found")
)

// LocationRepository handles location data access
type LocationRepository struct {
	db *pgxpool.Pool
}

// NewLocationRepository creates a new location repository
func NewLocationRepository(db *pgxpool.Pool) *LocationRepository {
	return &LocationRepository{db: db}
}

// FindByID retrieves a location by ID
func (r *LocationRepository) FindByID(ctx context.Context, id string) (*model.Location, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, type, created_at, updated_at
		FROM locations
		WHERE id = $1
	`

	var loc model.Location
	err := r.db.QueryRow(ctx, query, id).Scan(
		&loc.ID, &loc.TenantID, &loc.ParentID, &loc.Name, &loc.Type, &loc.CreatedAt, &loc.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrLocationNotFound
		}
		return nil, err
	}

	return &loc, nil
}

// ListByTenant retrieves all locations for a tenant
func (r *LocationRepository) ListByTenant(ctx context.Context, tenantID string) ([]model.Location, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, type, created_at, updated_at
		FROM locations
		WHERE tenant_id = $1
		ORDER BY type, name
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var loc model.Location
		err := rows.Scan(
			&loc.ID, &loc.TenantID, &loc.ParentID, &loc.Name, &loc.Type, &loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// ListChildren retrieves child locations of a parent
func (r *LocationRepository) ListChildren(ctx context.Context, parentID string) ([]model.Location, error) {
	query := `
		SELECT id, tenant_id, parent_id, name, type, created_at, updated_at
		FROM locations
		WHERE parent_id = $1
		ORDER BY name
	`

	rows, err := r.db.Query(ctx, query, parentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var locations []model.Location
	for rows.Next() {
		var loc model.Location
		err := rows.Scan(
			&loc.ID, &loc.TenantID, &loc.ParentID, &loc.Name, &loc.Type, &loc.CreatedAt, &loc.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		locations = append(locations, loc)
	}

	return locations, nil
}

// Create inserts a new location
func (r *LocationRepository) Create(ctx context.Context, loc *model.Location) error {
	query := `
		INSERT INTO locations (tenant_id, parent_id, name, type)
		VALUES ($1, $2, $3, $4)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		loc.TenantID, loc.ParentID, loc.Name, loc.Type,
	).Scan(&loc.ID, &loc.CreatedAt, &loc.UpdatedAt)
}

// Update modifies an existing location
func (r *LocationRepository) Update(ctx context.Context, loc *model.Location) error {
	query := `
		UPDATE locations
		SET parent_id = $2, name = $3, type = $4, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	return r.db.QueryRow(ctx, query,
		loc.ID, loc.ParentID, loc.Name, loc.Type,
	).Scan(&loc.UpdatedAt)
}

// Delete removes a location
func (r *LocationRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM locations WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrLocationNotFound
	}
	return nil
}
