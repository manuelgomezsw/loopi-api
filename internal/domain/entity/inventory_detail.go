package entity

import "time"

// InventoryDetail represents the detail of an inventory for a specific item.
type InventoryDetail struct {
	ID             uint32    `json:"id"`
	InventoryID    uint32    `json:"inventory_id"`
	ItemID         uint16    `json:"item_id"`
	SuggestedValue *uint16   `json:"suggested_value,omitempty"`
	RealValue      *uint16   `json:"real_value,omitempty"`
	StockReceived  *uint16   `json:"stock_received,omitempty"`
	UnitsSold      *uint16   `json:"units_sold,omitempty"`
	Shrinkage      *uint16   `json:"shrinkage,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships (populated when needed)
	Item      *Item      `json:"item,omitempty"`
	Inventory *Inventory `json:"inventory,omitempty"`
}

// IsComplete returns true if the real value has been set.
func (d *InventoryDetail) IsComplete() bool {
	return d.RealValue != nil
}
