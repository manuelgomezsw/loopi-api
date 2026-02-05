package request

import (
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// CreateInventoryRequest represents the request to create a new inventory.
type CreateInventoryRequest struct {
	Schedule string `json:"schedule"`
	Date     string `json:"date"` // Format: YYYY-MM-DD
}

// Validate validates the create inventory request.
func (r *CreateInventoryRequest) Validate() error {
	if r.Schedule == "" {
		return ErrScheduleRequired
	}
	if !isValidSchedule(r.Schedule) {
		return ErrInvalidSchedule
	}
	if r.Date == "" {
		return ErrDateRequired
	}
	if _, err := time.Parse("2006-01-02", r.Date); err != nil {
		return ErrInvalidDate
	}
	return nil
}

// GetSchedule returns the schedule as entity.Schedule.
func (r *CreateInventoryRequest) GetSchedule() entity.Schedule {
	return entity.Schedule(r.Schedule)
}

// GetDate returns the date as time.Time.
func (r *CreateInventoryRequest) GetDate() time.Time {
	t, _ := time.Parse("2006-01-02", r.Date)
	return t
}

// SaveInventoryDetailRequest represents the request to save an inventory detail.
type SaveInventoryDetailRequest struct {
	ItemID        uint16  `json:"item_id"`
	RealValue     uint16  `json:"real_value"`
	StockReceived *uint16 `json:"stock_received,omitempty"`
	UnitsSold     *uint16 `json:"units_sold,omitempty"`
}

// Validate validates the save inventory detail request.
func (r *SaveInventoryDetailRequest) Validate() error {
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
		"weekly":  true,
		"monthly": true,
	}
	return validSchedules[s]
}
