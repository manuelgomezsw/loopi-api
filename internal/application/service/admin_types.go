package service

import (
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// --- Dashboard ---

// DashboardStats contains dashboard statistics.
type DashboardStats struct {
	TodayInventories     int `json:"today_inventories"`
	WithDiscrepancies    int `json:"with_discrepancies"`
	WithoutDiscrepancies int `json:"without_discrepancies"`
	PendingInventories   int `json:"pending_inventories"`
}

// DashboardData contains dashboard data (stats only).
type DashboardData struct {
	Stats DashboardStats `json:"stats"`
}

// --- Inventory ---

// InventoryFilter contains filter options for inventory list.
type InventoryFilter struct {
	DateFrom         *time.Time
	DateTo           *time.Time
	InventoryType    *entity.InventoryType
	EmployeeID       *uint16
	HasDiscrepancies *bool
	Page             int
	PageSize         int
}

// InventoryListItem represents an inventory in the list view.
type InventoryListItem struct {
	ID            uint32                 `json:"id"`
	InventoryDate time.Time              `json:"inventory_date"`
	InventoryType entity.InventoryType   `json:"inventory_type"`
	Schedule      *entity.Schedule       `json:"schedule,omitempty"`
	Status        entity.InventoryStatus `json:"status"`
	EmployeeID    uint16                 `json:"employee_id"`
	EmployeeName  string                 `json:"employee_name"`
	TotalItems    int                    `json:"total_items"`
	ItemsWithDiff int                    `json:"items_with_diff"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
}

// InventoryListResult contains the paginated list result.
type InventoryListResult struct {
	Items      []InventoryListItem `json:"items"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// InventoryDetailView represents the detailed view of an inventory.
type InventoryDetailView struct {
	ID            uint32                 `json:"id"`
	InventoryDate time.Time              `json:"inventory_date"`
	InventoryType entity.InventoryType   `json:"inventory_type"`
	Schedule      *entity.Schedule       `json:"schedule,omitempty"`
	Status        entity.InventoryStatus `json:"status"`
	EmployeeID    uint16                 `json:"employee_id"`
	EmployeeName  string                 `json:"employee_name"`
	StartedAt     time.Time              `json:"started_at"`
	CompletedAt   *time.Time             `json:"completed_at,omitempty"`
	TotalItems    int                    `json:"total_items"`
	ItemsWithDiff int                    `json:"items_with_diff"`
	Details       []InventoryDetailItem  `json:"details"`
}

// InventoryDetailItem represents a single item in the inventory detail.
type InventoryDetailItem struct {
	DetailID       uint32  `json:"detail_id"`
	ItemID         uint16  `json:"item_id"`
	ItemName       string  `json:"item_name"`
	ItemType       string  `json:"item_type"`
	SuggestedValue *uint16 `json:"suggested_value"`
	RealValue      *uint16 `json:"real_value"`
	StockReceived  *uint16 `json:"stock_received"`
	UnitsSold      *uint16 `json:"units_sold"`
	Shrinkage      *uint16 `json:"shrinkage"`
	ExpectedValue  uint16  `json:"expected_value"`
	Difference     int16   `json:"difference"`
	HasDiscrepancy bool    `json:"has_discrepancy"`
}

// --- Item ---

// ItemFilter contains filters for listing items.
type ItemFilter struct {
	Type      *entity.ItemType           `json:"type"`
	Frequency *entity.InventoryFrequency `json:"frequency"`
	Active    *bool                      `json:"active"`
	Search    string                     `json:"search"`
	Page      int                        `json:"page"`
	PageSize  int                        `json:"page_size"`
}

// ItemListResult contains paginated item results.
type ItemListResult struct {
	Items      []*entity.Item `json:"items"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
}

// CreateItemRequest contains data for creating an item.
type CreateItemRequest struct {
	Type                   entity.ItemType           `json:"type"`
	Name                   string                    `json:"name"`
	InventoryFrequency     entity.InventoryFrequency `json:"inventory_frequency"`
	CategoryID             uint16                    `json:"category_id"`
	SupplierID             *uint16                   `json:"supplier_id"`
	Cost                   uint32                    `json:"cost"`
	MeasurementUnitID      uint16                    `json:"measurement_unit_id"`
	AddToActiveInventories bool                      `json:"add_to_active_inventories"`
}

// UpdateItemRequest contains data for updating an item.
type UpdateItemRequest struct {
	Type               entity.ItemType           `json:"type"`
	Name               string                    `json:"name"`
	InventoryFrequency entity.InventoryFrequency `json:"inventory_frequency"`
	Active             bool                      `json:"active"`
	CategoryID         uint16                    `json:"category_id"`
	SupplierID         *uint16                   `json:"supplier_id"`
	Cost               uint32                    `json:"cost"`
	MeasurementUnitID  uint16                    `json:"measurement_unit_id"`
}

// --- Employee ---

// EmployeeFilter contains filters for listing employees.
type EmployeeFilter struct {
	Role     *entity.Role `json:"role"`
	Active   *bool        `json:"active"`
	Search   string       `json:"search"`
	Page     int          `json:"page"`
	PageSize int          `json:"page_size"`
}

// EmployeeListResult contains paginated employee results.
type EmployeeListResult struct {
	Employees  []*entity.Employee `json:"employees"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// CreateEmployeeRequest contains data for creating an employee.
type CreateEmployeeRequest struct {
	Username       string      `json:"username"`
	Password       string      `json:"password"`
	Name           string      `json:"name"`
	LastName       string      `json:"last_name"`
	DocumentType   *string     `json:"document_type"`
	DocumentNumber *string     `json:"document_number"`
	Phone          *string     `json:"phone"`
	Email          *string     `json:"email"`
	BirthDate      *time.Time  `json:"birth_date"`
	Role           entity.Role `json:"role"`
}

// UpdateEmployeeRequest contains data for updating an employee.
type UpdateEmployeeRequest struct {
	Username       string      `json:"username"`
	Name           string      `json:"name"`
	LastName       string      `json:"last_name"`
	DocumentType   *string     `json:"document_type"`
	DocumentNumber *string     `json:"document_number"`
	Phone          *string     `json:"phone"`
	Email          *string     `json:"email"`
	BirthDate      *time.Time  `json:"birth_date"`
	Role           entity.Role `json:"role"`
	Active         bool        `json:"active"`
}

// --- Category ---

// CreateCategoryRequest contains data for creating a category.
type CreateCategoryRequest struct {
	Name string `json:"name"`
}

// UpdateCategoryRequest contains data for updating a category.
type UpdateCategoryRequest struct {
	Name   string `json:"name"`
	Active bool   `json:"active"`
}

// ReorderCategoryRequest contains data for reordering categories.
type ReorderCategoryRequest struct {
	Orders []CategoryOrderItem `json:"orders"`
}

// CategoryOrderItem represents a single category order update.
type CategoryOrderItem struct {
	ID           uint16 `json:"id"`
	DisplayOrder int    `json:"display_order"`
}

// --- Supplier ---

// SupplierFilter contains filters for listing suppliers.
type SupplierFilter struct {
	Active   *bool  `json:"active"`
	Search   string `json:"search"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

// SupplierListResult contains paginated supplier results.
type SupplierListResult struct {
	Suppliers  []*entity.Supplier `json:"suppliers"`
	Total      int                `json:"total"`
	Page       int                `json:"page"`
	PageSize   int                `json:"page_size"`
	TotalPages int                `json:"total_pages"`
}

// CreateSupplierRequest contains data for creating a supplier.
type CreateSupplierRequest struct {
	BusinessName string `json:"business_name"`
	TaxID        string `json:"tax_id"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
}

// UpdateSupplierRequest contains data for updating a supplier.
type UpdateSupplierRequest struct {
	BusinessName string `json:"business_name"`
	TaxID        string `json:"tax_id"`
	ContactName  string `json:"contact_name"`
	ContactPhone string `json:"contact_phone"`
	ContactEmail string `json:"contact_email"`
	Active       bool   `json:"active"`
}
