package response

import "time"

// SuggestedScheduleResponse represents the suggested schedule response.
type SuggestedScheduleResponse struct {
	Schedule string `json:"schedule"`
	Date     string `json:"date"`
}

// InventoryResponse represents an inventory in API responses.
type InventoryResponse struct {
	ID            uint32     `json:"id"`
	InventoryDate string     `json:"inventory_date"`
	Schedule      string     `json:"schedule"`
	Status        string     `json:"status"`
	ResponsibleID uint16     `json:"responsible_id"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// InventoryItemResponse represents an item to inventory with suggested value.
type InventoryItemResponse struct {
	ItemID         uint16  `json:"item_id"`
	Name           string  `json:"name"`
	SuggestedValue *uint16 `json:"suggested_value,omitempty"`
	RealValue      *uint16 `json:"real_value,omitempty"`
	StockReceived  *uint16 `json:"stock_received,omitempty"`
	UnitsSold      *uint16 `json:"units_sold,omitempty"`
	RequiresSales  bool    `json:"requires_sales"`
	IsComplete     bool    `json:"is_complete"`
}

// InventoryItemsResponse represents the list of items for an inventory.
type InventoryItemsResponse struct {
	InventoryID   uint32                  `json:"inventory_id"`
	Schedule      string                  `json:"schedule"`
	Date          string                  `json:"date"`
	TotalItems    int                     `json:"total_items"`
	CompletedItems int                    `json:"completed_items"`
	Items         []InventoryItemResponse `json:"items"`
}

// InventorySummaryItem represents an item in the inventory summary.
type InventorySummaryItem struct {
	ItemID         uint16 `json:"item_id"`
	Name           string `json:"name"`
	SuggestedValue uint16 `json:"suggested_value"`
	RealValue      uint16 `json:"real_value"`
	Difference     int16  `json:"difference"`
	HasDiscrepancy bool   `json:"has_discrepancy"`
}

// InventorySummaryResponse represents the inventory summary before completion.
type InventorySummaryResponse struct {
	InventoryID      uint32                 `json:"inventory_id"`
	Schedule         string                 `json:"schedule"`
	Date             string                 `json:"date"`
	TotalItems       int                    `json:"total_items"`
	ItemsWithIssues  int                    `json:"items_with_issues"`
	Items            []InventorySummaryItem `json:"items"`
	CanComplete      bool                   `json:"can_complete"`
	MissingItems     int                    `json:"missing_items"`
}

// CompleteInventoryResponse represents the response after completing an inventory.
type CompleteInventoryResponse struct {
	Completed     bool `json:"completed"`
	IssuesCreated int  `json:"issues_created"`
}

// SaveDetailResponse represents the response after saving a detail.
type SaveDetailResponse struct {
	Saved          bool    `json:"saved"`
	SuggestedValue *uint16 `json:"suggested_value,omitempty"`
}
