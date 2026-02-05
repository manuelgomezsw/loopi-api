package entity

import "time"

// Schedule represents the time of day for an inventory.
type Schedule string

const (
	ScheduleOpening Schedule = "opening"
	ScheduleNoon    Schedule = "noon"
	ScheduleClosing Schedule = "closing"
	ScheduleWeekly  Schedule = "weekly"
	ScheduleMonthly Schedule = "monthly"
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
	Schedule      Schedule        `json:"schedule"`
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

// IsDaily returns true if the schedule is a daily schedule.
func (i *Inventory) IsDaily() bool {
	return i.Schedule == ScheduleOpening || i.Schedule == ScheduleNoon || i.Schedule == ScheduleClosing
}

// RequiresSalesAndPurchases returns true if the schedule requires sales and purchases input.
func (i *Inventory) RequiresSalesAndPurchases() bool {
	return i.Schedule != ScheduleOpening
}
