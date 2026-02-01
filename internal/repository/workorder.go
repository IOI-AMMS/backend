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
	ErrWorkOrderNotFound = errors.New("work order not found")
)

// WorkOrderRepository handles work order data access
type WorkOrderRepository struct {
	db *pgxpool.Pool
}

// NewWorkOrderRepository creates a new work order repository
func NewWorkOrderRepository(db *pgxpool.Pool) *WorkOrderRepository {
	return &WorkOrderRepository{db: db}
}

// FindByID retrieves a work order by ID (v1.1 schema)
func (r *WorkOrderRepository) FindByID(ctx context.Context, id string) (*model.WorkOrder, error) {
	query := `
		SELECT 
			w.id, w.tenant_id, w.readable_id, w.asset_id, w.assigned_user_id,
			w.status, w.origin, w.priority, w.title,
			w.description, w.started_at, w.completed_at, w.created_at,
			COALESCE(a.name, '') as asset_name
		FROM work_orders w
		LEFT JOIN assets a ON w.asset_id = a.id
		WHERE w.id = $1
	`

	var wo model.WorkOrder
	err := r.db.QueryRow(ctx, query, id).Scan(
		&wo.ID, &wo.TenantID, &wo.ReadableID, &wo.AssetID, &wo.AssignedUserID,
		&wo.Status, &wo.Origin, &wo.Priority, &wo.Title,
		&wo.Description, &wo.StartedAt, &wo.CompletedAt, &wo.CreatedAt,
		&wo.AssetName,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrWorkOrderNotFound
		}
		return nil, err
	}

	return &wo, nil
}

// List retrieves work orders with filtering and pagination (v1.1 schema)
func (r *WorkOrderRepository) List(ctx context.Context, params model.WorkOrderListParams) (*model.PaginatedResult[model.WorkOrder], error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("w.tenant_id = $%d", argNum))
	args = append(args, params.TenantID)
	argNum++

	if len(params.Status) > 0 {
		placeholders := make([]string, len(params.Status))
		for i, s := range params.Status {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, s)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("w.status IN (%s)", strings.Join(placeholders, ",")))
	}

	if len(params.Priority) > 0 {
		placeholders := make([]string, len(params.Priority))
		for i, p := range params.Priority {
			placeholders[i] = fmt.Sprintf("$%d", argNum)
			args = append(args, p)
			argNum++
		}
		conditions = append(conditions, fmt.Sprintf("w.priority IN (%s)", strings.Join(placeholders, ",")))
	}

	if params.AssetID != "" {
		conditions = append(conditions, fmt.Sprintf("w.asset_id = $%d", argNum))
		args = append(args, params.AssetID)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM work_orders w WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Sorting
	sortBy := "w.created_at"
	if params.SortBy != "" {
		sortBy = "w." + params.SortBy
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
			w.id, w.tenant_id, w.readable_id, w.asset_id, w.assigned_user_id,
			w.status, w.origin, w.priority, w.title,
			w.description, w.started_at, w.completed_at, w.created_at,
			COALESCE(a.name, '') as asset_name
		FROM work_orders w
		LEFT JOIN assets a ON w.asset_id = a.id
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

	var workOrders []model.WorkOrder
	for rows.Next() {
		var wo model.WorkOrder
		err := rows.Scan(
			&wo.ID, &wo.TenantID, &wo.ReadableID, &wo.AssetID, &wo.AssignedUserID,
			&wo.Status, &wo.Origin, &wo.Priority, &wo.Title,
			&wo.Description, &wo.StartedAt, &wo.CompletedAt, &wo.CreatedAt,
			&wo.AssetName,
		)
		if err != nil {
			return nil, err
		}
		workOrders = append(workOrders, wo)
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.PaginatedResult[model.WorkOrder]{
		Data:       workOrders,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// Create inserts a new work order (v1.1 schema)
func (r *WorkOrderRepository) Create(ctx context.Context, wo *model.WorkOrder) error {
	query := `
		INSERT INTO work_orders (tenant_id, asset_id, assigned_user_id, status, origin, priority, title, description)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, readable_id, created_at
	`

	return r.db.QueryRow(ctx, query,
		wo.TenantID, wo.AssetID, wo.AssignedUserID, wo.Status, wo.Origin, wo.Priority, wo.Title, wo.Description,
	).Scan(&wo.ID, &wo.ReadableID, &wo.CreatedAt)
}

// Update modifies an existing work order (v1.1 schema)
func (r *WorkOrderRepository) Update(ctx context.Context, wo *model.WorkOrder) error {
	query := `
		UPDATE work_orders
		SET status = $2, priority = $3, description = $4, title = $5
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query, wo.ID, wo.Status, wo.Priority, wo.Description, wo.Title)
	return err
}

// UpdateStatus changes a work order's status
func (r *WorkOrderRepository) UpdateStatus(ctx context.Context, id, status string) error {
	query := `UPDATE work_orders SET status = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id, status)
	return err
}
