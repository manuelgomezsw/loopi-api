package entity

import "time"

// IssueType represents the type of inventory issue.
type IssueType string

const (
	IssueTypeDiscrepancy     IssueType = "discrepancy"
	IssueTypeSkippedSchedule IssueType = "skipped_schedule"
)

// IssueStatus represents the status of an inventory issue.
type IssueStatus string

const (
	IssueStatusOpen     IssueStatus = "open"
	IssueStatusResolved IssueStatus = "resolved"
)

// InventoryIssue represents a discrepancy or anomaly in an inventory.
type InventoryIssue struct {
	ID                uint32      `json:"id"`
	InventoryDetailID uint32      `json:"inventory_detail_id"`
	Type              IssueType   `json:"type"`
	ExpectedValue     *uint16     `json:"expected_value,omitempty"`
	ActualValue       *uint16     `json:"actual_value,omitempty"`
	Difference        *int16      `json:"difference,omitempty"`
	Status            IssueStatus `json:"status"`
	ResolutionNotes   *string     `json:"resolution_notes,omitempty"`
	ResolvedBy        *uint16     `json:"resolved_by,omitempty"`
	ResolvedAt        *time.Time  `json:"resolved_at,omitempty"`
	CreatedAt         time.Time   `json:"created_at"`

	// Relationships (populated when needed)
	InventoryDetail *InventoryDetail `json:"inventory_detail,omitempty"`
	Resolver        *Employee        `json:"resolver,omitempty"`
}

// IsResolved returns true if the issue has been resolved.
func (i *InventoryIssue) IsResolved() bool {
	return i.Status == IssueStatusResolved
}
