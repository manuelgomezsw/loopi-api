package entity

import "time"

// Category represents a product category for organizing items.
type Category struct {
	ID           uint16    `json:"id"`
	Name         string    `json:"name"`
	DisplayOrder int       `json:"display_order"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Computed fields
	ItemCount int `json:"item_count,omitempty"`
}
