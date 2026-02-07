package response

import "time"

// SuggestedScheduleResponse represents the suggested schedule response.
type SuggestedScheduleResponse struct {
	InventoryType string  `json:"inventory_type"`
	Schedule      *string `json:"schedule,omitempty"`
	Date          string  `json:"date"`
}

// InventoryResponse represents an inventory in API responses.
type InventoryResponse struct {
	ID            uint32     `json:"id"`
	InventoryDate string     `json:"inventory_date"`
	InventoryType string     `json:"inventory_type"`
	Schedule      *string    `json:"schedule,omitempty"`
	Status        string     `json:"status"`
	ResponsibleID uint16     `json:"responsible_id"`
	StartedAt     time.Time  `json:"started_at"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

// InventoryItemResponse represents an item to inventory with suggested value.
type InventoryItemResponse struct {
	ItemID         uint16  `json:"item_id"`
	Name           string  `json:"name"`
	CategoryName   string  `json:"category_name,omitempty"`
	SuggestedValue *uint16 `json:"suggested_value,omitempty"`
	RealValue      *uint16 `json:"real_value,omitempty"`
	StockReceived  *uint16 `json:"stock_received,omitempty"`
	UnitsSold      *uint16 `json:"units_sold,omitempty"`
	IsComplete     bool    `json:"is_complete"`
}

// InventoryItemsResponse represents the list of items for an inventory.
type InventoryItemsResponse struct {
	InventoryID    uint32                  `json:"inventory_id"`
	InventoryType  string                  `json:"inventory_type"`
	Schedule       *string                 `json:"schedule,omitempty"`
	Date           string                  `json:"date"`
	RequiresSales  bool                    `json:"requires_sales"`
	TotalItems     int                     `json:"total_items"`
	CompletedItems int                     `json:"completed_items"`
	Items          []InventoryItemResponse `json:"items"`
}

// DiscrepancyItemResponse represents an item with discrepancy.
type DiscrepancyItemResponse struct {
	ItemID         uint16 `json:"item_id"`
	Name           string `json:"name"`
	SuggestedValue uint16 `json:"suggested_value"`
	RealValue      uint16 `json:"real_value"`
	Difference     int16  `json:"difference"`
	StockReceived  *uint16 `json:"stock_received,omitempty"`
	UnitsSold      *uint16 `json:"units_sold,omitempty"`
}

// DiscrepanciesResponse represents the list of items with discrepancies.
type DiscrepanciesResponse struct {
	InventoryID      uint32                    `json:"inventory_id"`
	InventoryType    string                    `json:"inventory_type"`
	Schedule         *string                   `json:"schedule,omitempty"`
	Date             string                    `json:"date"`
	RequiresSales    bool                      `json:"requires_sales"`
	TotalItems       int                       `json:"total_items"`
	HasDiscrepancies bool                      `json:"has_discrepancies"`
	Items            []DiscrepancyItemResponse `json:"items"`
}

// InventorySummaryItem represents an item in the inventory summary.
type InventorySummaryItem struct {
	ItemID         uint16  `json:"item_id"`
	Name           string  `json:"name"`
	SuggestedValue uint16  `json:"suggested_value"`
	RealValue      uint16  `json:"real_value"`
	Difference     int16   `json:"difference"`
	HasDiscrepancy bool    `json:"has_discrepancy"`
	StockReceived  *uint16 `json:"stock_received,omitempty"`
	UnitsSold      *uint16 `json:"units_sold,omitempty"`
}

// InventorySummaryResponse represents the inventory summary before completion.
type InventorySummaryResponse struct {
	InventoryID     uint32                 `json:"inventory_id"`
	InventoryType   string                 `json:"inventory_type"`
	Schedule        *string                `json:"schedule,omitempty"`
	Date            string                 `json:"date"`
	TotalItems      int                    `json:"total_items"`
	ItemsWithIssues int                    `json:"items_with_issues"`
	Items           []InventorySummaryItem `json:"items"`
	CanComplete     bool                   `json:"can_complete"`
	MissingItems    int                    `json:"missing_items"`
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

// InProgressInventoryResponse represents an in-progress inventory.
type InProgressInventoryResponse struct {
	ID            uint32    `json:"id"`
	InventoryDate string    `json:"inventory_date"`
	InventoryType string    `json:"inventory_type"`
	Schedule      *string   `json:"schedule,omitempty"`
	StartedAt     time.Time `json:"started_at"`
}

// InProgressInventoriesResponse represents the list of in-progress inventories.
type InProgressInventoriesResponse struct {
	Inventories []InProgressInventoryResponse `json:"inventories"`
	Count       int                           `json:"count"`
}
