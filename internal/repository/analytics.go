package repository

import (
	"context"
	"time"

	"ioi-amms/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type AnalyticsRepository struct {
	db *pgxpool.Pool
}

func NewAnalyticsRepository(db *pgxpool.Pool) *AnalyticsRepository {
	return &AnalyticsRepository{db: db}
}

// GetDashboardStats aggregates counts for the dashboard
func (r *AnalyticsRepository) GetDashboardStats(ctx context.Context, tenantID string) (*model.DashboardStats, error) {
	stats := &model.DashboardStats{}

	// 1. Total Assets
	err := r.db.QueryRow(ctx, "SELECT COUNT(*) FROM assets WHERE tenant_id = $1", tenantID).Scan(&stats.TotalAssets)
	if err != nil {
		return nil, err
	}

	// 2. Open Work Orders (Anything not Completed or Cancelled)
	err = r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM work_orders WHERE tenant_id = $1 AND status NOT IN ('Completed', 'Cancelled')",
		tenantID).Scan(&stats.OpenWorkOrders)
	if err != nil {
		return nil, err
	}

	// 3. Maintenance Due (Work orders with 'High' or 'Critical' priority that are open)
	err = r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM work_orders WHERE tenant_id = $1 AND status NOT IN ('Completed', 'Cancelled') AND priority IN ('High', 'Critical')",
		tenantID).Scan(&stats.MaintenanceDue)
	if err != nil {
		return nil, err
	}

	// 4. Critical Alerts (Assets with 'Critical' status)
	err = r.db.QueryRow(ctx,
		"SELECT COUNT(*) FROM assets WHERE tenant_id = $1 AND status = 'Critical'",
		tenantID).Scan(&stats.CriticalAlerts)
	if err != nil {
		return nil, err
	}

	// 5. Completion Rate (Completed WOs / Total WOs)
	var totalWOs, completedWOs int
	err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM work_orders WHERE tenant_id = $1", tenantID).Scan(&totalWOs)
	if err != nil {
		return nil, err
	}

	if totalWOs > 0 {
		err = r.db.QueryRow(ctx, "SELECT COUNT(*) FROM work_orders WHERE tenant_id = $1 AND status = 'Completed'", tenantID).Scan(&completedWOs)
		if err != nil {
			return nil, err
		}
		stats.CompletionRate = (completedWOs * 100) / totalWOs
	} else {
		stats.CompletionRate = 100 // Default to 100% if no work orders exist yet
	}

	stats.OpenFaults = stats.CriticalAlerts // Mapping "Faults" to Critical Assets for now

	return stats, nil
}

// GetTopAlerts returns a list of critical items for the dashboard feed
func (r *AnalyticsRepository) GetTopAlerts(ctx context.Context, tenantID string) ([]model.Alert, error) {
	// For now, we generate alerts based on Critical assets
	query := `
		SELECT id, name, updated_at 
		FROM assets 
		WHERE tenant_id = $1 AND status = 'Critical'
		ORDER BY updated_at DESC
		LIMIT 5
	`

	rows, err := r.db.Query(ctx, query, tenantID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []model.Alert
	for rows.Next() {
		var id, name string
		var updatedAt time.Time
		if err := rows.Scan(&id, &name, &updatedAt); err != nil {
			return nil, err
		}

		alerts = append(alerts, model.Alert{
			ID:        id,
			Title:     "Asset Reported Critical Status",
			Severity:  "Critical",
			AssetID:   id,
			AssetName: name,
			Timestamp: updatedAt.Format(time.RFC3339),
		})
	}

	return alerts, nil
}
