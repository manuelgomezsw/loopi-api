package entity

import "time"

// Supplier represents a product supplier.
type Supplier struct {
	ID           uint16    `json:"id"`
	BusinessName string    `json:"business_name"`
	TaxID        string    `json:"tax_id"` // NIT
	ContactName  string    `json:"contact_name"`
	ContactPhone string    `json:"contact_phone"`
	ContactEmail string    `json:"contact_email"`
	Active       bool      `json:"active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`

	// Computed fields
	ItemCount int `json:"item_count,omitempty"`
}
