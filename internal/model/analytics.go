package model

// DashboardStats aggregates high-level KPIs for the dashboard
type DashboardStats struct {
	TotalAssets    int `json:"totalAssets"`
	MaintenanceDue int `json:"maintenanceDue"`
	OpenFaults     int `json:"openFaults"`
	CriticalAlerts int `json:"criticalAlerts"`
	OpenWorkOrders int `json:"openWorkOrders"`
	CompletionRate int `json:"completionRate"`
}

// Alert represents a high-priority system notification
type Alert struct {
	ID        string `json:"id"`
	Title     string `json:"title"`
	Severity  string `json:"severity"` // Critical, Warning, Info
	AssetID   string `json:"assetId,omitempty"`
	AssetName string `json:"assetName,omitempty"`
	Timestamp string `json:"timestamp"`
}

// AnalyticsResponse wraps the full dashboard data payload
type AnalyticsResponse struct {
	Stats  DashboardStats `json:"stats"`
	Alerts []Alert        `json:"alerts"`
}
