package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"ioi-amms/internal/model"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrAssetNotFound = errors.New("asset not found")
)

// AssetRepository handles asset data access
type AssetRepository struct {
	db *pgxpool.Pool
}

// NewAssetRepository creates a new asset repository
func NewAssetRepository(db *pgxpool.Pool) *AssetRepository {
	return &AssetRepository{db: db}
}

// FindByID retrieves an asset by ID
func (r *AssetRepository) FindByID(ctx context.Context, id string) (*model.Asset, error) {
	query := `
		SELECT 
			a.id, a.tenant_id, a.parent_id, a.location_id, a.name, 
			a.status, a.criticality, a.last_inspection, a.created_at, a.updated_at,
			COALESCE(l.name, '') as location_name
		FROM assets a
		LEFT JOIN locations l ON a.location_id = l.id
		WHERE a.id = $1
	`

	var asset model.Asset
	err := r.db.QueryRow(ctx, query, id).Scan(
		&asset.ID, &asset.TenantID, &asset.ParentID, &asset.LocationID, &asset.Name,
		&asset.Status, &asset.Criticality, &asset.LastInspection, &asset.CreatedAt, &asset.UpdatedAt,
		&asset.LocationName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrAssetNotFound
		}
		return nil, err
	}

	return &asset, nil
}

// List retrieves assets with filtering and pagination
func (r *AssetRepository) List(ctx context.Context, params model.AssetListParams) (*model.PaginatedResult[model.Asset], error) {
	// Build dynamic query
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("a.tenant_id = $%d", argNum))
	args = append(args, params.TenantID)
	argNum++

	if len(params.Status) > 0 {
		placeholders := make([]string, len(params.Status))
		for i, s := range params.Status {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, s)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("a.status::text IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(params.Criticality) > 0 {
		placeholders := make([]string, len(params.Criticality))
		for i, c := range params.Criticality {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, c)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("a.criticality::text IN (%s)", strings.Join(placeholders, ",")))
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("a.name ILIKE $%d", argNum))
		args = append(args, "%"+params.Search+"%")
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM assets a WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Sorting
	sortBy := "a.created_at"
	if params.SortBy != "" {
		sortBy = "a." + params.SortBy
	}
	sortDir := "DESC"
	if params.SortDir == "asc" {
		sortDir = "ASC"
	}

	// Pagination
	if params.Limit == 0 {
		params.Limit = 10
	}
	if params.Page == 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit

	query := fmt.Sprintf(`
		SELECT 
			a.id, a.tenant_id, a.parent_id, a.location_id, a.name, 
			a.status, a.criticality, a.last_inspection, a.created_at, a.updated_at,
			COALESCE(l.name, '') as location_name
		FROM assets a
		LEFT JOIN locations l ON a.location_id = l.id
		WHERE %s
		ORDER BY %s %s
		LIMIT $%d OFFSET $%d
	`, whereClause, sortBy, sortDir, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var assets []model.Asset
	for rows.Next() {
		var asset model.Asset
		err := rows.Scan(
			&asset.ID, &asset.TenantID, &asset.ParentID, &asset.LocationID, &asset.Name,
			&asset.Status, &asset.Criticality, &asset.LastInspection, &asset.CreatedAt, &asset.UpdatedAt,
			&asset.LocationName,
		)
		if err != nil {
			return nil, err
		}
		assets = append(assets, asset)
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.PaginatedResult[model.Asset]{
		Data:       assets,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// Create inserts a new asset
func (r *AssetRepository) Create(ctx context.Context, asset *model.Asset) error {
	query := `
		INSERT INTO assets (tenant_id, parent_id, location_id, name, status, criticality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at, updated_at
	`

	return r.db.QueryRow(ctx, query,
		asset.TenantID, asset.ParentID, asset.LocationID, asset.Name, asset.Status, asset.Criticality,
	).Scan(&asset.ID, &asset.CreatedAt, &asset.UpdatedAt)
}

// Update modifies an existing asset
func (r *AssetRepository) Update(ctx context.Context, asset *model.Asset) error {
	query := `
		UPDATE assets
		SET parent_id = $2, location_id = $3, name = $4, status = $5, criticality = $6, updated_at = NOW()
		WHERE id = $1
		RETURNING updated_at
	`

	return r.db.QueryRow(ctx, query,
		asset.ID, asset.ParentID, asset.LocationID, asset.Name, asset.Status, asset.Criticality,
	).Scan(&asset.UpdatedAt)
}

// Delete removes an asset
func (r *AssetRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM assets WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrAssetNotFound
	}
	return nil
}
