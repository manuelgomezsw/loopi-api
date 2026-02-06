package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// AdminService handles admin-specific operations.
type AdminService struct {
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	employeeRepo        repository.EmployeeRepository
}

// NewAdminService creates a new admin service.
func NewAdminService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	employeeRepo repository.EmployeeRepository,
) *AdminService {
	return &AdminService{
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		employeeRepo:        employeeRepo,
	}
}

// DashboardStats contains dashboard statistics.
type DashboardStats struct {
	TodayInventories      int `json:"today_inventories"`
	WithDiscrepancies     int `json:"with_discrepancies"`
	WithoutDiscrepancies  int `json:"without_discrepancies"`
	PendingInventories    int `json:"pending_inventories"`
}

// DiscrepancySummary contains a summary of a discrepancy for the dashboard.
type DiscrepancySummary struct {
	InventoryID   uint32    `json:"inventory_id"`
	ItemID        uint16    `json:"item_id"`
	ItemName      string    `json:"item_name"`
	ExpectedValue uint16    `json:"expected_value"`
	ActualValue   uint16    `json:"actual_value"`
	Difference    int16     `json:"difference"`
	InventoryDate time.Time `json:"inventory_date"`
	InventoryType string    `json:"inventory_type"`
}

// DashboardData contains all dashboard data.
type DashboardData struct {
	Stats                 DashboardStats       `json:"stats"`
	RecentDiscrepancies   []DiscrepancySummary `json:"recent_discrepancies"`
}

// GetDashboard returns dashboard data.
func (s *AdminService) GetDashboard(ctx context.Context, days int) (*DashboardData, error) {
	stats, err := s.inventoryRepo.GetDashboardStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get dashboard stats: %w", err)
	}

	discrepancies, err := s.inventoryDetailRepo.GetRecentDiscrepancies(ctx, days)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent discrepancies: %w", err)
	}

	// Convert to DiscrepancySummary
	recentDiscrepancies := make([]DiscrepancySummary, 0, len(discrepancies))
	for _, d := range discrepancies {
		var expected, actual uint16
		if d.SuggestedValue != nil {
			expected = *d.SuggestedValue
		}
		if d.RealValue != nil {
			actual = *d.RealValue
		}
		
		recentDiscrepancies = append(recentDiscrepancies, DiscrepancySummary{
			InventoryID:   d.InventoryID,
			ItemID:        d.ItemID,
			ItemName:      d.Item.Name,
			ExpectedValue: expected,
			ActualValue:   actual,
			Difference:    d.Difference(),
			InventoryDate: d.Inventory.InventoryDate,
			InventoryType: string(d.Inventory.InventoryType),
		})
	}

	return &DashboardData{
		Stats: DashboardStats{
			TodayInventories:     stats.TodayInventories,
			WithDiscrepancies:    stats.WithDiscrepancies,
			WithoutDiscrepancies: stats.WithoutDiscrepancies,
			PendingInventories:   stats.PendingInventories,
		},
		RecentDiscrepancies: recentDiscrepancies,
	}, nil
}

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
	ID               uint32              `json:"id"`
	InventoryDate    time.Time           `json:"inventory_date"`
	InventoryType    entity.InventoryType `json:"inventory_type"`
	Schedule         *entity.Schedule    `json:"schedule,omitempty"`
	Status           entity.InventoryStatus `json:"status"`
	EmployeeID       uint16              `json:"employee_id"`
	EmployeeName     string              `json:"employee_name"`
	TotalItems       int                 `json:"total_items"`
	ItemsWithDiff    int                 `json:"items_with_diff"`
	StartedAt        time.Time           `json:"started_at"`
	CompletedAt      *time.Time          `json:"completed_at,omitempty"`
}

// InventoryListResult contains the paginated list result.
type InventoryListResult struct {
	Items      []InventoryListItem `json:"items"`
	Total      int                 `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
}

// ListInventories returns a paginated list of inventories with filters.
func (s *AdminService) ListInventories(ctx context.Context, filter InventoryFilter) (*InventoryListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	inventories, total, err := s.inventoryRepo.FindAllWithFilters(ctx, filter.DateFrom, filter.DateTo, filter.InventoryType, filter.EmployeeID, filter.HasDiscrepancies, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}

	items := make([]InventoryListItem, 0, len(inventories))
	for _, inv := range inventories {
		items = append(items, InventoryListItem{
			ID:            inv.ID,
			InventoryDate: inv.InventoryDate,
			InventoryType: inv.InventoryType,
			Schedule:      inv.Schedule,
			Status:        inv.Status,
			EmployeeID:    inv.ResponsibleID,
			EmployeeName:  inv.Employee.FullName(),
			TotalItems:    inv.TotalItems,
			ItemsWithDiff: inv.ItemsWithDiff,
			StartedAt:     inv.StartedAt,
			CompletedAt:   inv.CompletedAt,
		})
	}

	totalPages := (total + filter.PageSize - 1) / filter.PageSize

	return &InventoryListResult{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// InventoryDetailView represents the detailed view of an inventory.
type InventoryDetailView struct {
	ID            uint32                `json:"id"`
	InventoryDate time.Time             `json:"inventory_date"`
	InventoryType entity.InventoryType  `json:"inventory_type"`
	Schedule      *entity.Schedule      `json:"schedule,omitempty"`
	Status        entity.InventoryStatus `json:"status"`
	EmployeeID    uint16                `json:"employee_id"`
	EmployeeName  string                `json:"employee_name"`
	StartedAt     time.Time             `json:"started_at"`
	CompletedAt   *time.Time            `json:"completed_at,omitempty"`
	TotalItems    int                   `json:"total_items"`
	ItemsWithDiff int                   `json:"items_with_diff"`
	Details       []InventoryDetailItem `json:"details"`
}

// InventoryDetailItem represents a single item in the inventory detail.
type InventoryDetailItem struct {
	DetailID      uint32  `json:"detail_id"`
	ItemID        uint16  `json:"item_id"`
	ItemName      string  `json:"item_name"`
	ItemType      string  `json:"item_type"`
	SuggestedValue *uint16 `json:"suggested_value"`
	RealValue     *uint16 `json:"real_value"`
	StockReceived *uint16 `json:"stock_received"`
	UnitsSold     *uint16 `json:"units_sold"`
	Difference    int16   `json:"difference"`
	HasDiscrepancy bool   `json:"has_discrepancy"`
}

// GetInventoryDetail returns detailed information about an inventory.
func (s *AdminService) GetInventoryDetail(ctx context.Context, inventoryID uint32) (*InventoryDetailView, error) {
	inventory, err := s.inventoryRepo.FindByIDWithEmployee(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return nil, nil
	}

	details, err := s.inventoryDetailRepo.FindByInventoryIDWithItems(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory details: %w", err)
	}

	detailItems := make([]InventoryDetailItem, 0, len(details))
	itemsWithDiff := 0
	for _, d := range details {
		hasDiscrepancy := d.HasDiscrepancy()
		if hasDiscrepancy {
			itemsWithDiff++
		}
		
		detailItems = append(detailItems, InventoryDetailItem{
			DetailID:       d.ID,
			ItemID:         d.ItemID,
			ItemName:       d.Item.Name,
			ItemType:       string(d.Item.Type),
			SuggestedValue: d.SuggestedValue,
			RealValue:      d.RealValue,
			StockReceived:  d.StockReceived,
			UnitsSold:      d.UnitsSold,
			Difference:     d.Difference(),
			HasDiscrepancy: hasDiscrepancy,
		})
	}

	return &InventoryDetailView{
		ID:            inventory.ID,
		InventoryDate: inventory.InventoryDate,
		InventoryType: inventory.InventoryType,
		Schedule:      inventory.Schedule,
		Status:        inventory.Status,
		EmployeeID:    inventory.ResponsibleID,
		EmployeeName:  inventory.Employee.FullName(),
		StartedAt:     inventory.StartedAt,
		CompletedAt:   inventory.CompletedAt,
		TotalItems:    len(details),
		ItemsWithDiff: itemsWithDiff,
		Details:       detailItems,
	}, nil
}

// UpdateInventoryDetail updates a specific inventory detail (admin can edit closed inventories).
func (s *AdminService) UpdateInventoryDetail(ctx context.Context, inventoryID uint32, detailID uint32, realValue, stockReceived, unitsSold *uint16) error {
	detail, err := s.inventoryDetailRepo.FindByID(ctx, detailID)
	if err != nil {
		return fmt.Errorf("failed to get inventory detail: %w", err)
	}
	if detail == nil || detail.InventoryID != inventoryID {
		return fmt.Errorf("detail not found")
	}

	detail.RealValue = realValue
	detail.StockReceived = stockReceived
	detail.UnitsSold = unitsSold

	if err := s.inventoryDetailRepo.Update(ctx, detail); err != nil {
		return fmt.Errorf("failed to update inventory detail: %w", err)
	}

	return nil
}
