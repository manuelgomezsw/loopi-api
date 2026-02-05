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
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`

	// Relationships (populated when needed)
	Item *Item `json:"item,omitempty"`
}

// HasDiscrepancy returns true if the real value differs from the suggested value.
func (d *InventoryDetail) HasDiscrepancy() bool {
	if d.SuggestedValue == nil || d.RealValue == nil {
		return false
	}
	return *d.SuggestedValue != *d.RealValue
}

// Difference returns the difference between real and suggested values.
// Positive means surplus, negative means missing.
func (d *InventoryDetail) Difference() int16 {
	if d.SuggestedValue == nil || d.RealValue == nil {
		return 0
	}
	return int16(*d.RealValue) - int16(*d.SuggestedValue)
}

// IsComplete returns true if the real value has been set.
func (d *InventoryDetail) IsComplete() bool {
	return d.RealValue != nil
}
