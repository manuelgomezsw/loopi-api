package entity

import "time"

// InventoryType represents the type/frequency of an inventory.
type InventoryType string

const (
	InventoryTypeDaily   InventoryType = "daily"
	InventoryTypeWeekly  InventoryType = "weekly"
	InventoryTypeMonthly InventoryType = "monthly"
)

// Schedule represents the time of day for a daily inventory.
type Schedule string

const (
	ScheduleOpening Schedule = "opening"
	ScheduleNoon    Schedule = "noon"
	ScheduleClosing Schedule = "closing"
)

// InventoryStatus represents the status of an inventory.
type InventoryStatus string

const (
	InventoryStatusInProgress InventoryStatus = "in_progress"
	InventoryStatusCompleted  InventoryStatus = "completed"
)

// Inventory represents an inventory session.
type Inventory struct {
	ID            uint32          `json:"id"`
	InventoryDate time.Time       `json:"inventory_date"`
	InventoryType InventoryType   `json:"inventory_type"`
	Schedule      *Schedule       `json:"schedule,omitempty"` // Nullable for weekly/monthly
	Status        InventoryStatus `json:"status"`
	ResponsibleID uint16          `json:"responsible_id"`
	StartedAt     time.Time       `json:"started_at"`
	CompletedAt   *time.Time      `json:"completed_at,omitempty"`
	CreatedAt     time.Time       `json:"created_at"`

	// Relationships (populated when needed)
	Responsible *Employee          `json:"responsible,omitempty"`
	Details     []*InventoryDetail `json:"details,omitempty"`
}

// IsCompleted returns true if the inventory is completed.
func (i *Inventory) IsCompleted() bool {
	return i.Status == InventoryStatusCompleted
}

// IsDaily returns true if the inventory type is daily.
func (i *Inventory) IsDaily() bool {
	return i.InventoryType == InventoryTypeDaily
}

// RequiresSalesAndPurchases returns true if the inventory requires sales and purchases input.
// This is true for all inventories except opening daily inventories.
func (i *Inventory) RequiresSalesAndPurchases() bool {
	if !i.IsDaily() {
		return true // Weekly and monthly always require sales/purchases
	}
	return i.Schedule != nil && *i.Schedule != ScheduleOpening
}

// GetScheduleString returns the schedule as a string, or empty if nil.
func (i *Inventory) GetScheduleString() string {
	if i.Schedule == nil {
		return ""
	}
	return string(*i.Schedule)
}
