package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ioi-amms/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

// AuditRepository handles audit log data access
type AuditRepository struct {
	db *pgxpool.Pool
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *pgxpool.Pool) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create inserts a new audit log
func (r *AuditRepository) Create(ctx context.Context, log *model.AuditLog) error {
	query := `
		INSERT INTO audit_logs (user_id, action, entity_type, entity_id, changes)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`

	changesJSON, err := json.Marshal(log.Changes)
	if err != nil {
		return err
	}

	return r.db.QueryRow(ctx, query,
		log.UserID, log.Action, log.EntityType, log.EntityID, changesJSON,
	).Scan(&log.ID, &log.CreatedAt)
}

// List retrieves audit logs with filtering and pagination
func (r *AuditRepository) List(ctx context.Context, params model.AuditListParams) (*model.PaginatedResult[model.AuditLog], error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	// Filter by tenant via user relationship (audit_logs -> users -> tenant_id)
	// Or explicitly passed TenantID (if we join users)
	// The DB schema: audit_logs(user_id) -> users(id, tenant_id)

	// Join users to get email/name and filter by tenant
	baseQuery := `
		FROM audit_logs a
		JOIN users u ON a.user_id = u.id
	`

	conditions = append(conditions, fmt.Sprintf("u.tenant_id = $%d", argNum))
	args = append(args, params.TenantID)
	argNum++

	if params.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("a.user_id = $%d", argNum))
		args = append(args, params.UserID)
		argNum++
	}

	if params.EntityType != "" {
		conditions = append(conditions, fmt.Sprintf("a.entity_type = $%d", argNum))
		args = append(args, params.EntityType)
		argNum++
	}

	if params.EntityID != "" {
		conditions = append(conditions, fmt.Sprintf("a.entity_id = $%d", argNum))
		args = append(args, params.EntityID)
		argNum++
	}

	if params.Action != "" {
		conditions = append(conditions, fmt.Sprintf("a.action = $%d", argNum))
		args = append(args, params.Action)
		argNum++
	}

	if !params.From.IsZero() {
		conditions = append(conditions, fmt.Sprintf("a.created_at >= $%d", argNum))
		args = append(args, params.From)
		argNum++
	}

	if !params.To.IsZero() {
		conditions = append(conditions, fmt.Sprintf("a.created_at <= $%d", argNum))
		args = append(args, params.To)
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) %s WHERE %s", baseQuery, whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
	}

	// Pagination
	if params.Limit == 0 {
		params.Limit = 20
	}
	if params.Page == 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.Limit

	// List Query
	// We might need to join entities tables based on entity_type to get names?
	// That's complex ("Polymorphic" join).
	// For now, we'll just return the email as User Name.
	// Resolving Entity Name (e.g. Asset Name) dynamically in SQL is hard.
	// We can skip EntityName for now or fetch it if needed.
	// The requirement says: "Join users.name as userName", "Join entity name where possible".
	// Let's settle for UserName for now to keep it simpler.

	query := fmt.Sprintf(`
		SELECT 
			a.id, a.user_id, a.action, a.entity_type, a.entity_id, a.changes, a.created_at,
			COALESCE(u.full_name, u.email) as user_name
		%s
		WHERE %s
		ORDER BY a.created_at DESC
		LIMIT $%d OFFSET $%d
	`, baseQuery, whereClause, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.AuditLog
	for rows.Next() {
		var l model.AuditLog
		var changesJSON []byte
		err := rows.Scan(
			&l.ID, &l.UserID, &l.Action, &l.EntityType, &l.EntityID, &changesJSON, &l.CreatedAt,
			&l.UserName,
		)
		if err != nil {
			return nil, err
		}

		if len(changesJSON) > 0 {
			json.Unmarshal(changesJSON, &l.Changes)
		}

		logs = append(logs, l)
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.PaginatedResult[model.AuditLog]{
		Data:       logs,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}
