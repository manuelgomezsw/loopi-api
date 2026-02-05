package entity

import "time"

// ItemType represents the type of an item.
type ItemType string

const (
	ItemTypeProduct ItemType = "product"
	ItemTypeSupply  ItemType = "supply"
)

// InventoryFrequency represents how often an item should be inventoried.
type InventoryFrequency string

const (
	InventoryFrequencyDaily   InventoryFrequency = "daily"
	InventoryFrequencyWeekly  InventoryFrequency = "weekly"
	InventoryFrequencyMonthly InventoryFrequency = "monthly"
)

// Item represents a product or supply in the inventory.
type Item struct {
	ID                 uint16             `json:"id"`
	Type               ItemType           `json:"type"`
	Name               string             `json:"name"`
	Active             bool               `json:"active"`
	InventoryFrequency InventoryFrequency `json:"inventory_frequency"`
	CreatedAt          time.Time          `json:"created_at"`
	UpdatedAt          time.Time          `json:"updated_at"`
}
