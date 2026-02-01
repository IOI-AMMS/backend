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
	ErrPartNotFound  = errors.New("part not found")
	ErrStockNotFound = errors.New("stock not found")
)

// InventoryRepository handles parts and inventory data access
type InventoryRepository struct {
	db *pgxpool.Pool
}

// NewInventoryRepository creates a new inventory repository
func NewInventoryRepository(db *pgxpool.Pool) *InventoryRepository {
	return &InventoryRepository{db: db}
}

// ==================== PARTS CRUD ====================

// CreatePart inserts a new part
func (r *InventoryRepository) CreatePart(ctx context.Context, part *model.Part) error {
	query := `
		INSERT INTO parts (tenant_id, sku, name, category, uom, min_stock_level, is_stock_item)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id, created_at
	`

	return r.db.QueryRow(ctx, query,
		part.TenantID, part.SKU, part.Name, part.Category, part.UOM, part.MinStockLevel, part.IsStockItem,
	).Scan(&part.ID, &part.CreatedAt)
}

// FindPartByID retrieves a part by ID
func (r *InventoryRepository) FindPartByID(ctx context.Context, id string) (*model.Part, error) {
	query := `
		SELECT id, tenant_id, sku, name, category, uom, min_stock_level, is_stock_item, created_at
		FROM parts
		WHERE id = $1
	`

	var part model.Part
	err := r.db.QueryRow(ctx, query, id).Scan(
		&part.ID, &part.TenantID, &part.SKU, &part.Name, &part.Category,
		&part.UOM, &part.MinStockLevel, &part.IsStockItem, &part.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrPartNotFound
		}
		return nil, err
	}

	return &part, nil
}

// ListParts retrieves parts with filtering and pagination
func (r *InventoryRepository) ListParts(ctx context.Context, params model.PartListParams) (*model.PaginatedResult[model.Part], error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("tenant_id = $%d", argNum))
	args = append(args, params.TenantID)
	argNum++

	if params.Category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argNum))
		args = append(args, params.Category)
		argNum++
	}

	if params.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR sku ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+params.Search+"%")
		argNum++
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM parts WHERE %s", whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
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
		SELECT id, tenant_id, sku, name, category, uom, min_stock_level, is_stock_item, created_at
		FROM parts
		WHERE %s
		ORDER BY name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var parts []model.Part
	for rows.Next() {
		var part model.Part
		err := rows.Scan(
			&part.ID, &part.TenantID, &part.SKU, &part.Name, &part.Category,
			&part.UOM, &part.MinStockLevel, &part.IsStockItem, &part.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		parts = append(parts, part)
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.PaginatedResult[model.Part]{
		Data:       parts,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpdatePart modifies an existing part
func (r *InventoryRepository) UpdatePart(ctx context.Context, part *model.Part) error {
	query := `
		UPDATE parts
		SET sku = $2, name = $3, category = $4, uom = $5, min_stock_level = $6, is_stock_item = $7
		WHERE id = $1
	`

	_, err := r.db.Exec(ctx, query,
		part.ID, part.SKU, part.Name, part.Category, part.UOM, part.MinStockLevel, part.IsStockItem,
	)
	return err
}

// DeletePart removes a part
func (r *InventoryRepository) DeletePart(ctx context.Context, id string) error {
	query := `DELETE FROM parts WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrPartNotFound
	}
	return nil
}

// ==================== INVENTORY STOCK ====================

// ListStock retrieves inventory stock with filtering
func (r *InventoryRepository) ListStock(ctx context.Context, params model.InventoryStockListParams) (*model.PaginatedResult[model.InventoryStock], error) {
	var conditions []string
	var args []interface{}
	argNum := 1

	conditions = append(conditions, fmt.Sprintf("s.tenant_id = $%d", argNum))
	args = append(args, params.TenantID)
	argNum++

	if params.PartID != "" {
		conditions = append(conditions, fmt.Sprintf("s.part_id = $%d", argNum))
		args = append(args, params.PartID)
		argNum++
	}

	if params.LocationID != "" {
		conditions = append(conditions, fmt.Sprintf("s.location_id = $%d", argNum))
		args = append(args, params.LocationID)
		argNum++
	}

	if params.LowStock {
		conditions = append(conditions, "s.quantity_on_hand < p.min_stock_level")
	}

	whereClause := strings.Join(conditions, " AND ")

	// Count total
	countQuery := fmt.Sprintf(`
		SELECT COUNT(*) 
		FROM inventory_stock s
		JOIN parts p ON s.part_id = p.id
		WHERE %s
	`, whereClause)
	var total int
	err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, err
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
			s.id, s.tenant_id, s.part_id, s.location_id, s.quantity_on_hand, s.bin_label, s.updated_at,
			p.name as part_name, p.sku as part_sku,
			l.name as location_name
		FROM inventory_stock s
		JOIN parts p ON s.part_id = p.id
		LEFT JOIN locations l ON s.location_id = l.id
		WHERE %s
		ORDER BY p.name ASC
		LIMIT $%d OFFSET $%d
	`, whereClause, argNum, argNum+1)

	args = append(args, params.Limit, offset)

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stocks []model.InventoryStock
	for rows.Next() {
		var s model.InventoryStock
		err := rows.Scan(
			&s.ID, &s.TenantID, &s.PartID, &s.LocationID, &s.QuantityOnHand, &s.BinLabel, &s.UpdatedAt,
			&s.PartName, &s.PartSKU,
			&s.LocationName,
		)
		if err != nil {
			return nil, err
		}
		stocks = append(stocks, s)
	}

	totalPages := (total + params.Limit - 1) / params.Limit

	return &model.PaginatedResult[model.InventoryStock]{
		Data:       stocks,
		Total:      total,
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}, nil
}

// UpsertStock creates or updates stock at a location
func (r *InventoryRepository) UpsertStock(ctx context.Context, stock *model.InventoryStock) error {
	query := `
		INSERT INTO inventory_stock (tenant_id, part_id, location_id, quantity_on_hand, bin_label)
		VALUES ($1, $2, $3, $4, $5)
		ON CONFLICT (tenant_id, part_id, location_id) 
		DO UPDATE SET quantity_on_hand = EXCLUDED.quantity_on_hand, bin_label = EXCLUDED.bin_label, updated_at = NOW()
		RETURNING id, updated_at
	`

	return r.db.QueryRow(ctx, query,
		stock.TenantID, stock.PartID, stock.LocationID, stock.QuantityOnHand, stock.BinLabel,
	).Scan(&stock.ID, &stock.UpdatedAt)
}

// AdjustStock adjusts quantity at a location (positive or negative)
func (r *InventoryRepository) AdjustStock(ctx context.Context, tenantID, partID, locationID string, delta float64) error {
	query := `
		UPDATE inventory_stock
		SET quantity_on_hand = quantity_on_hand + $4, updated_at = NOW()
		WHERE tenant_id = $1 AND part_id = $2 AND location_id = $3
	`

	result, err := r.db.Exec(ctx, query, tenantID, partID, locationID, delta)
	if err != nil {
		return err
	}
	if result.RowsAffected() == 0 {
		return ErrStockNotFound
	}
	return nil
}

// ==================== INVENTORY WALLETS ====================

// ListWallets retrieves user inventory wallets
func (r *InventoryRepository) ListWallets(ctx context.Context, userID string) ([]model.InventoryWallet, error) {
	query := `
		SELECT 
			w.user_id, w.part_id, w.qty_held, w.last_updated_at,
			p.name as part_name, p.sku as part_sku
		FROM inventory_wallets w
		JOIN parts p ON w.part_id = p.id
		WHERE w.user_id = $1
		ORDER BY p.name ASC
	`

	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var wallets []model.InventoryWallet
	for rows.Next() {
		var w model.InventoryWallet
		err := rows.Scan(
			&w.UserID, &w.PartID, &w.QtyHeld, &w.LastUpdatedAt,
			&w.PartName, &w.PartSKU,
		)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, w)
	}

	return wallets, nil
}

// UpsertWallet creates or updates a user's wallet for a part
func (r *InventoryRepository) UpsertWallet(ctx context.Context, wallet *model.InventoryWallet) error {
	query := `
		INSERT INTO inventory_wallets (user_id, part_id, qty_held)
		VALUES ($1, $2, $3)
		ON CONFLICT (user_id, part_id) 
		DO UPDATE SET qty_held = EXCLUDED.qty_held, last_updated_at = NOW()
		RETURNING last_updated_at
	`

	return r.db.QueryRow(ctx, query,
		wallet.UserID, wallet.PartID, wallet.QtyHeld,
	).Scan(&wallet.LastUpdatedAt)
}

// AdjustWallet adjusts quantity in a user's wallet
func (r *InventoryRepository) AdjustWallet(ctx context.Context, userID, partID string, delta float64) error {
	query := `
		UPDATE inventory_wallets
		SET qty_held = qty_held + $3, last_updated_at = NOW()
		WHERE user_id = $1 AND part_id = $2
	`

	_, err := r.db.Exec(ctx, query, userID, partID, delta)
	return err
}
