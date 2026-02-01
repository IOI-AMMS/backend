package model

import (
	"time"
)

// Part represents a spare part in the parts catalog (v1.1 schema)
type Part struct {
	ID            string    `json:"id"`
	TenantID      string    `json:"tenantId"`
	SKU           string    `json:"sku"`
	Name          string    `json:"name"`
	Category      *string   `json:"category,omitempty"`
	UOM           string    `json:"uom"`
	MinStockLevel float64   `json:"minStockLevel"`
	IsStockItem   bool      `json:"isStockItem"`
	CreatedAt     time.Time `json:"createdAt"`
}

// PartListParams for filtering parts
type PartListParams struct {
	TenantID string
	Category string
	Search   string
	Page     int
	Limit    int
}

// InventoryStock represents stock at a location (v1.1 schema)
type InventoryStock struct {
	ID             string    `json:"id"`
	TenantID       string    `json:"tenantId"`
	PartID         string    `json:"partId"`
	LocationID     string    `json:"locationId"`
	QuantityOnHand float64   `json:"quantityOnHand"`
	BinLabel       *string   `json:"binLabel,omitempty"`
	UpdatedAt      time.Time `json:"updatedAt"`

	// Joined fields
	PartName     string `json:"partName,omitempty"`
	PartSKU      string `json:"partSku,omitempty"`
	LocationName string `json:"locationName,omitempty"`
}

// InventoryStockListParams for filtering inventory
type InventoryStockListParams struct {
	TenantID   string
	PartID     string
	LocationID string
	LowStock   bool // Filter items below min_stock_level
	Page       int
	Limit      int
}

// InventoryWallet represents technician-held inventory (v1.1 schema)
type InventoryWallet struct {
	UserID        string    `json:"userId"`
	PartID        string    `json:"partId"`
	QtyHeld       float64   `json:"qtyHeld"`
	LastUpdatedAt time.Time `json:"lastUpdatedAt"`

	// Joined fields
	PartName string `json:"partName,omitempty"`
	PartSKU  string `json:"partSku,omitempty"`
	UserName string `json:"userName,omitempty"`
}

// StockTransfer represents a transfer between locations or to/from wallets
type StockTransfer struct {
	PartID         string  `json:"partId"`
	FromLocationID *string `json:"fromLocationId,omitempty"`
	ToLocationID   *string `json:"toLocationId,omitempty"`
	FromUserID     *string `json:"fromUserId,omitempty"`
	ToUserID       *string `json:"toUserId,omitempty"`
	Quantity       float64 `json:"quantity"`
	Notes          *string `json:"notes,omitempty"`
}
