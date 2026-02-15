package request

import (
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// CreateInventoryRequest represents the request to create a new inventory.
type CreateInventoryRequest struct {
	InventoryType string  `json:"inventory_type"` // daily, weekly, monthly
	Schedule      *string `json:"schedule,omitempty"` // opening, noon, closing (required for daily)
	Date          string  `json:"date"` // Format: YYYY-MM-DD
}

// Validate validates the create inventory request.
func (r *CreateInventoryRequest) Validate() error {
	if r.InventoryType == "" {
		return ErrInventoryTypeRequired
	}
	if !isValidInventoryType(r.InventoryType) {
		return ErrInvalidInventoryType
	}
	if r.InventoryType == "daily" {
		if r.Schedule == nil || *r.Schedule == "" {
			return ErrScheduleRequired
		}
		if !isValidSchedule(*r.Schedule) {
			return ErrInvalidSchedule
		}
	}
	if r.Date == "" {
		return ErrDateRequired
	}
	if _, err := time.Parse("2006-01-02", r.Date); err != nil {
		return ErrInvalidDate
	}
	return nil
}

// GetInventoryType returns the inventory type as entity.InventoryType.
func (r *CreateInventoryRequest) GetInventoryType() entity.InventoryType {
	return entity.InventoryType(r.InventoryType)
}

// GetSchedule returns the schedule as *entity.Schedule (nil for weekly/monthly).
func (r *CreateInventoryRequest) GetSchedule() *entity.Schedule {
	if r.Schedule == nil || r.InventoryType != "daily" {
		return nil
	}
	s := entity.Schedule(*r.Schedule)
	return &s
}

// GetDate returns the date as time.Time.
func (r *CreateInventoryRequest) GetDate() time.Time {
	t, _ := time.Parse("2006-01-02", r.Date)
	return t
}

// SaveInventoryDetailRequest represents the request to save a physical count.
type SaveInventoryDetailRequest struct {
	ItemID    uint16 `json:"item_id"`
	RealValue uint16 `json:"real_value"`
}

// Validate validates the save inventory detail request.
func (r *SaveInventoryDetailRequest) Validate() error {
	if r.ItemID == 0 {
		return ErrItemIDRequired
	}
	return nil
}

// SaveSalesRequest represents the request to save sales and purchases.
type SaveSalesRequest struct {
	ItemID        uint16  `json:"item_id"`
	StockReceived *uint16 `json:"stock_received,omitempty"`
	UnitsSold     *uint16 `json:"units_sold,omitempty"`
}

// Validate validates the save sales request.
func (r *SaveSalesRequest) Validate() error {
	if r.ItemID == 0 {
		return ErrItemIDRequired
	}
	return nil
}

// isValidSchedule checks if the schedule value is valid.
func isValidSchedule(s string) bool {
	validSchedules := map[string]bool{
		"opening": true,
		"noon":    true,
		"closing": true,
	}
	return validSchedules[s]
}

// CreateInitialInventoryRequest represents the request to create an initial (baseline) inventory.
type CreateInitialInventoryRequest struct {
	ResponsibleID uint16 `json:"responsible_id"`
}

// Validate validates the create initial inventory request.
func (r *CreateInitialInventoryRequest) Validate() error {
	if r.ResponsibleID == 0 {
		return ErrResponsibleIDRequired
	}
	return nil
}

// isValidInventoryType checks if the inventory type value is valid.
func isValidInventoryType(t string) bool {
	validTypes := map[string]bool{
		"daily":   true,
		"weekly":  true,
		"monthly": true,
		"initial": true,
	}
	return validTypes[t]
}
