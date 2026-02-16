package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	invdomain "github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
	"golang.org/x/crypto/bcrypt"
)

// ErrShrinkageOnlyWhenCompleted is returned when trying to edit shrinkage on a non-completed inventory.
var ErrShrinkageOnlyWhenCompleted = errors.New("shrinkage can only be edited on completed inventories")

// AdminService handles admin-specific operations.
type AdminService struct {
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	employeeRepo        repository.EmployeeRepository
	itemRepo            repository.ItemRepository
	categoryRepo        repository.CategoryRepository
	supplierRepo        repository.SupplierRepository
	enricher            *invdomain.Enricher
}

// NewAdminService creates a new admin service.
func NewAdminService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	employeeRepo repository.EmployeeRepository,
	itemRepo repository.ItemRepository,
	categoryRepo repository.CategoryRepository,
	supplierRepo repository.SupplierRepository,
	enricher *invdomain.Enricher,
) *AdminService {
	return &AdminService{
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		employeeRepo:        employeeRepo,
		itemRepo:            itemRepo,
		categoryRepo:        categoryRepo,
		supplierRepo:        supplierRepo,
		enricher:            enricher,
	}
}

// DashboardStats contains dashboard statistics.
type DashboardStats struct {
	TodayInventories     int `json:"today_inventories"`
	WithDiscrepancies    int `json:"with_discrepancies"`
	WithoutDiscrepancies int `json:"without_discrepancies"`
	PendingInventories   int `json:"pending_inventories"`
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
	Stats               DashboardStats       `json:"stats"`
	RecentDiscrepancies []DiscrepancySummary `json:"recent_discrepancies"`
}

// GetDashboard returns dashboard data (stats and recent discrepancies computed in service using domain rules).
func (s *AdminService) GetDashboard(ctx context.Context, days int) (*DashboardData, error) {
	today := datetime.Today()

	todayCount, err := s.inventoryRepo.CountInventoriesByDate(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("failed to count today inventories: %w", err)
	}

	pending, err := s.inventoryRepo.CountInProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to count pending: %w", err)
	}

	completedToday, err := s.inventoryRepo.FindCompletedInventoriesByDate(ctx, today)
	if err != nil {
		return nil, fmt.Errorf("failed to find completed inventories by date: %w", err)
	}

	withDiscrepancies := 0
	for _, inv := range completedToday {
		if s.countItemsWithDiscrepancy(ctx, inv) > 0 {
			withDiscrepancies++
		}
	}
	withoutDiscrepancies := len(completedToday) - withDiscrepancies
	if withoutDiscrepancies < 0 {
		withoutDiscrepancies = 0
	}

	// Recent details (raw); enrich per inventory, filter by domain rule, build summary (limit 20).
	recentDetails, err := s.inventoryDetailRepo.FindRecentDetailsWithInventory(ctx, days, 100)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent details: %w", err)
	}

	// Group by inventory ID and enrich each group, then collect items with discrepancy.
	type key struct{ invID uint32; itemID uint16 }
	seen := make(map[key]struct{})
	recentDiscrepancies := make([]DiscrepancySummary, 0, 20)

	invDetails := make(map[uint32][]*entity.InventoryDetail)
	for _, d := range recentDetails {
		invDetails[d.InventoryID] = append(invDetails[d.InventoryID], d)
	}

	for _, details := range invDetails {
		if len(details) == 0 {
			continue
		}
		inv := details[0].Inventory
		if inv == nil {
			continue
		}
		s.enricher.Enrich(ctx, inv, details)
		for _, d := range details {
			expected := invdomain.ExpectedAtEnd(d)
			if !invdomain.HasDiscrepancyFromExpectedEnd(d, expected) {
				continue
			}
			k := key{d.InventoryID, d.ItemID}
			if _, ok := seen[k]; ok {
				continue
			}
			seen[k] = struct{}{}
			var actual uint16
			if d.RealValue != nil {
				actual = *d.RealValue
			}
			recentDiscrepancies = append(recentDiscrepancies, DiscrepancySummary{
				InventoryID:   d.InventoryID,
				ItemID:        d.ItemID,
				ItemName:      d.Item.Name,
				ExpectedValue: expected,
				ActualValue:   actual,
				Difference:    invdomain.DifferenceFromExpected(d, expected),
				InventoryDate: inv.InventoryDate,
				InventoryType: string(inv.InventoryType),
			})
			if len(recentDiscrepancies) >= 20 {
				break
			}
		}
		if len(recentDiscrepancies) >= 20 {
			break
		}
	}

	return &DashboardData{
		Stats: DashboardStats{
			TodayInventories:     todayCount,
			WithDiscrepancies:    withDiscrepancies,
			WithoutDiscrepancies: withoutDiscrepancies,
			PendingInventories:   pending,
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

// ListInventories returns a paginated list of inventories with filters.
func (s *AdminService) ListInventories(ctx context.Context, filter InventoryFilter) (*InventoryListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	inventories, total, err := s.inventoryRepo.FindAllWithFilters(ctx, filter.DateFrom, filter.DateTo, filter.InventoryType, filter.EmployeeID, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list inventories: %w", err)
	}

	items := make([]InventoryListItem, 0, len(inventories))
	for _, inv := range inventories {
		totalItems, itemsWithDiff := s.getInventoryCounts(ctx, inv)
		item := InventoryListItem{
			ID:            inv.ID,
			InventoryDate: inv.InventoryDate,
			InventoryType: inv.InventoryType,
			Schedule:      inv.Schedule,
			Status:        inv.Status,
			EmployeeID:    inv.ResponsibleID,
			EmployeeName:  inv.Employee.FullName(),
			TotalItems:    totalItems,
			ItemsWithDiff: itemsWithDiff,
			StartedAt:     inv.StartedAt,
			CompletedAt:   inv.CompletedAt,
		}
		if filter.HasDiscrepancies != nil {
			hasDiff := itemsWithDiff > 0
			if *filter.HasDiscrepancies != hasDiff {
				continue
			}
		}
		items = append(items, item)
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
// ExpectedValue is computed as suggested_value − shrinkage + stock_received − units_sold,
// and is used for discrepancy and difference so that admin adjustments (mermas, compras, ventas) resolve differences.
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
	ExpectedValue  uint16  `json:"expected_value"` // sugerido − mermas + compras − ventas (used for discrepancy)
	Difference     int16   `json:"difference"`     // real_value − expected_value
	HasDiscrepancy bool    `json:"has_discrepancy"`
}

// countItemsWithDiscrepancy returns the count of details with real != expected (ExpectedForAdmin).
func (s *AdminService) countItemsWithDiscrepancy(ctx context.Context, inv *entity.Inventory) int {
	total, withDiff := s.getInventoryCounts(ctx, inv)
	_ = total
	return withDiff
}

// getInventoryCounts returns totalItems and itemsWithDiff for an inventory (enrich + domain rule).
func (s *AdminService) getInventoryCounts(ctx context.Context, inv *entity.Inventory) (totalItems, itemsWithDiff int) {
	details, err := s.inventoryDetailRepo.FindByInventoryID(ctx, inv.ID)
	if err != nil {
		return 0, 0
	}
	s.enricher.Enrich(ctx, inv, details)
	totalItems = len(details)
	for _, d := range details {
		if d.RealValue == nil {
			continue
		}
		expected := invdomain.ExpectedForAdmin(d)
		if invdomain.HasDiscrepancyFromExpectedEnd(d, expected) {
			itemsWithDiff++
		}
	}
	return totalItems, itemsWithDiff
}

// GetInventoryDetail returns detailed information about an inventory.
// Suggested values are recomputed from the previous inventory (real only) so expected and discrepancies match the employee view.
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

	s.enricher.Enrich(ctx, inventory, details)

	detailItems := make([]InventoryDetailItem, 0, len(details))
	itemsWithDiff := 0
	for _, d := range details {
		expected := invdomain.ExpectedForAdmin(d)
		diff := invdomain.DifferenceFromExpected(d, expected)
		hasDiscrepancy := invdomain.HasDiscrepancyFromExpectedEnd(d, expected)
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
			Shrinkage:      d.Shrinkage,
			ExpectedValue:  expected,
			Difference:     diff,
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
// Shrinkage can only be set or updated when the inventory is completed.
func (s *AdminService) UpdateInventoryDetail(ctx context.Context, inventoryID uint32, detailID uint32, realValue, stockReceived, unitsSold, shrinkage *uint16) error {
	inv, err := s.inventoryRepo.FindByID(ctx, inventoryID)
	if err != nil {
		return fmt.Errorf("failed to get inventory: %w", err)
	}
	if inv == nil {
		return fmt.Errorf("inventory not found")
	}

	if shrinkage != nil && inv.Status != entity.InventoryStatusCompleted {
		return ErrShrinkageOnlyWhenCompleted
	}

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
	if shrinkage != nil {
		detail.Shrinkage = shrinkage
	}

	if err := s.inventoryDetailRepo.Update(ctx, detail); err != nil {
		return fmt.Errorf("failed to update inventory detail: %w", err)
	}

	return nil
}

// --- Item Management ---

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
}

// ListItems retrieves items with filters and pagination.
func (s *AdminService) ListItems(ctx context.Context, filter ItemFilter) (*ItemListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	items, total, err := s.itemRepo.FindAllWithFilters(ctx, filter.Type, filter.Frequency, filter.Active, filter.Search, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list items: %w", err)
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	return &ItemListResult{
		Items:      items,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetItem retrieves a single item by ID.
func (s *AdminService) GetItem(ctx context.Context, id uint16) (*entity.Item, error) {
	item, err := s.itemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return nil, fmt.Errorf("item not found")
	}
	return item, nil
}

// CreateItem creates a new item and optionally adds it to all active inventories.
func (s *AdminService) CreateItem(ctx context.Context, req CreateItemRequest) (*entity.Item, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Type != entity.ItemTypeProduct && req.Type != entity.ItemTypeSupply {
		return nil, fmt.Errorf("invalid item type")
	}
	if req.InventoryFrequency != entity.InventoryFrequencyDaily &&
		req.InventoryFrequency != entity.InventoryFrequencyWeekly &&
		req.InventoryFrequency != entity.InventoryFrequencyMonthly {
		return nil, fmt.Errorf("invalid inventory frequency")
	}
	if req.CategoryID == 0 {
		return nil, fmt.Errorf("category is required")
	}

	item := &entity.Item{
		Type:               req.Type,
		Name:               req.Name,
		Active:             true,
		InventoryFrequency: req.InventoryFrequency,
		CategoryID:         req.CategoryID,
		SupplierID:         req.SupplierID,
		Cost:               req.Cost,
		CreatedAt:          datetime.Now(),
		UpdatedAt:          datetime.Now(),
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	// Add item to all active inventories if requested
	if req.AddToActiveInventories {
		if err := s.addItemToActiveInventories(ctx, item); err != nil {
			// Log the error but don't fail the item creation
			fmt.Printf("warning: failed to add item to active inventories: %v\n", err)
		}
	}

	return item, nil
}

// addItemToActiveInventories adds an item to all in-progress inventories
// that match the item's inventory frequency.
func (s *AdminService) addItemToActiveInventories(ctx context.Context, item *entity.Item) error {
	inventories, err := s.inventoryRepo.FindAllInProgress(ctx)
	if err != nil {
		return fmt.Errorf("failed to find active inventories: %w", err)
	}

	for _, inv := range inventories {
		// Check if the item should be included based on inventory type and frequency
		if !shouldIncludeItem(inv.InventoryType, item.InventoryFrequency) {
			continue
		}

		detail := &entity.InventoryDetail{
			InventoryID: inv.ID,
			ItemID:      item.ID,
		}

		if err := s.inventoryDetailRepo.Create(ctx, detail); err != nil {
			return fmt.Errorf("failed to add item %d to inventory %d: %w", item.ID, inv.ID, err)
		}
	}

	return nil
}

// shouldIncludeItem checks if an item should be included in an inventory
// based on the inventory type and item frequency.
func shouldIncludeItem(inventoryType entity.InventoryType, itemFrequency entity.InventoryFrequency) bool {
	switch inventoryType {
	case entity.InventoryTypeDaily:
		return itemFrequency == entity.InventoryFrequencyDaily
	case entity.InventoryTypeWeekly:
		return itemFrequency == entity.InventoryFrequencyDaily || itemFrequency == entity.InventoryFrequencyWeekly
	case entity.InventoryTypeMonthly, entity.InventoryTypeInitial:
		// Monthly and initial inventories include all items
		return true
	default:
		return false
	}
}

// GetActiveInventoriesCount returns the count of in-progress inventories.
func (s *AdminService) GetActiveInventoriesCount(ctx context.Context) (int, error) {
	count, err := s.inventoryRepo.CountInProgress(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count active inventories: %w", err)
	}
	return count, nil
}

// UpdateItem updates an existing item.
func (s *AdminService) UpdateItem(ctx context.Context, id uint16, req UpdateItemRequest) (*entity.Item, error) {
	item, err := s.itemRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return nil, fmt.Errorf("item not found")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.Type != entity.ItemTypeProduct && req.Type != entity.ItemTypeSupply {
		return nil, fmt.Errorf("invalid item type")
	}
	if req.InventoryFrequency != entity.InventoryFrequencyDaily &&
		req.InventoryFrequency != entity.InventoryFrequencyWeekly &&
		req.InventoryFrequency != entity.InventoryFrequencyMonthly {
		return nil, fmt.Errorf("invalid inventory frequency")
	}
	if req.CategoryID == 0 {
		return nil, fmt.Errorf("category is required")
	}

	item.Type = req.Type
	item.Name = req.Name
	item.InventoryFrequency = req.InventoryFrequency
	item.Active = req.Active
	item.CategoryID = req.CategoryID
	item.SupplierID = req.SupplierID
	item.Cost = req.Cost
	item.UpdatedAt = datetime.Now()

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return item, nil
}

// UpdateItemStatus updates the active status of an item.
func (s *AdminService) UpdateItemStatus(ctx context.Context, id uint16, active bool) error {
	item, err := s.itemRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}
	if item == nil {
		return fmt.Errorf("item not found")
	}

	if err := s.itemRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update item status: %w", err)
	}

	return nil
}

// --- Employee Management ---

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

// ListEmployees retrieves employees with filters and pagination.
func (s *AdminService) ListEmployees(ctx context.Context, filter EmployeeFilter) (*EmployeeListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	employees, total, err := s.employeeRepo.FindAllWithFilters(ctx, filter.Role, filter.Active, filter.Search, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	return &EmployeeListResult{
		Employees:  employees,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetEmployee retrieves a single employee by ID.
func (s *AdminService) GetEmployee(ctx context.Context, id uint16) (*entity.Employee, error) {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}
	return employee, nil
}

// CreateEmployee creates a new employee.
func (s *AdminService) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*entity.Employee, error) {
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
	}
	if req.Role != entity.RoleEmployee && req.Role != entity.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}

	// Check if username already exists
	existing, err := s.employeeRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	// Hash password
	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	employee := &entity.Employee{
		Username:       req.Username,
		PasswordHash:   passwordHash,
		Name:           req.Name,
		LastName:       req.LastName,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
		BirthDate:      req.BirthDate,
		Role:           req.Role,
		Active:         true,
		CreatedAt:      datetime.Now(),
		UpdatedAt:      datetime.Now(),
	}

	if err := s.employeeRepo.Create(ctx, employee); err != nil {
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	return employee, nil
}

// UpdateEmployee updates an existing employee.
func (s *AdminService) UpdateEmployee(ctx context.Context, id uint16, req UpdateEmployeeRequest) (*entity.Employee, error) {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}

	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
	}
	if req.Role != entity.RoleEmployee && req.Role != entity.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}

	employee.Username = req.Username
	employee.Name = req.Name
	employee.LastName = req.LastName
	employee.DocumentType = req.DocumentType
	employee.DocumentNumber = req.DocumentNumber
	employee.Phone = req.Phone
	employee.Email = req.Email
	employee.BirthDate = req.BirthDate
	employee.Role = req.Role
	employee.Active = req.Active
	employee.UpdatedAt = datetime.Now()

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	return employee, nil
}

// UpdateEmployeeStatus updates the active status of an employee.
func (s *AdminService) UpdateEmployeeStatus(ctx context.Context, id uint16, active bool) error {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return fmt.Errorf("employee not found")
	}

	if err := s.employeeRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update employee status: %w", err)
	}

	return nil
}

// ResetEmployeePassword resets an employee's password to default (document_number + birth_year).
func (s *AdminService) ResetEmployeePassword(ctx context.Context, id uint16) error {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return fmt.Errorf("employee not found")
	}

	// Generate default password: document_number + birth_year
	var defaultPassword string
	if employee.DocumentNumber != nil && employee.BirthDate != nil {
		defaultPassword = *employee.DocumentNumber + fmt.Sprintf("%d", employee.BirthDate.Year())
	} else if employee.DocumentNumber != nil {
		defaultPassword = *employee.DocumentNumber
	} else {
		defaultPassword = "password123" // Fallback if no document
	}

	passwordHash, err := hashPassword(defaultPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.employeeRepo.UpdatePassword(ctx, id, passwordHash); err != nil {
		return fmt.Errorf("failed to update password: %w", err)
	}

	return nil
}

// hashPassword hashes a password using bcrypt.
func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// --- Category Management ---

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

// ListCategories retrieves all categories.
func (s *AdminService) ListCategories(ctx context.Context) ([]*entity.Category, error) {
	categories, err := s.categoryRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list categories: %w", err)
	}
	return categories, nil
}

// GetCategory retrieves a single category by ID.
func (s *AdminService) GetCategory(ctx context.Context, id uint16) (*entity.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}
	return category, nil
}

// CreateCategory creates a new category.
func (s *AdminService) CreateCategory(ctx context.Context, req CreateCategoryRequest) (*entity.Category, error) {
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Check if name already exists
	existing, err := s.categoryRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check category name: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("category name already exists")
	}

	// Get next display order
	maxOrder, err := s.categoryRepo.GetMaxDisplayOrder(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get max display order: %w", err)
	}

	category := &entity.Category{
		Name:         req.Name,
		DisplayOrder: maxOrder + 1,
		Active:       true,
	}

	if err := s.categoryRepo.Create(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to create category: %w", err)
	}

	return category, nil
}

// UpdateCategory updates an existing category.
func (s *AdminService) UpdateCategory(ctx context.Context, id uint16, req UpdateCategoryRequest) (*entity.Category, error) {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return nil, fmt.Errorf("category not found")
	}

	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	// Check if name already exists for another category
	existing, err := s.categoryRepo.FindByName(ctx, req.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check category name: %w", err)
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("category name already exists")
	}

	category.Name = req.Name
	category.Active = req.Active

	if err := s.categoryRepo.Update(ctx, category); err != nil {
		return nil, fmt.Errorf("failed to update category: %w", err)
	}

	return category, nil
}

// UpdateCategoryStatus updates the active status of a category.
func (s *AdminService) UpdateCategoryStatus(ctx context.Context, id uint16, active bool) error {
	category, err := s.categoryRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get category: %w", err)
	}
	if category == nil {
		return fmt.Errorf("category not found")
	}

	if err := s.categoryRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update category status: %w", err)
	}

	return nil
}

// ReorderCategories updates the display order of multiple categories.
func (s *AdminService) ReorderCategories(ctx context.Context, req ReorderCategoryRequest) error {
	if len(req.Orders) == 0 {
		return nil
	}

	orders := make(map[uint16]int)
	for _, item := range req.Orders {
		orders[item.ID] = item.DisplayOrder
	}

	if err := s.categoryRepo.UpdateDisplayOrders(ctx, orders); err != nil {
		return fmt.Errorf("failed to reorder categories: %w", err)
	}

	return nil
}

// --- Supplier Management ---

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

// ListSuppliers retrieves suppliers with filters and pagination.
func (s *AdminService) ListSuppliers(ctx context.Context, filter SupplierFilter) (*SupplierListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	suppliers, total, err := s.supplierRepo.FindAllWithFilters(ctx, filter.Active, filter.Search, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list suppliers: %w", err)
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	return &SupplierListResult{
		Suppliers:  suppliers,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// ListAllActiveEmployees retrieves all active employees for dropdowns.
func (s *AdminService) ListAllActiveEmployees(ctx context.Context) ([]*entity.Employee, error) {
	employees, err := s.employeeRepo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active employees: %w", err)
	}
	return employees, nil
}

// CreateInitialInventory creates an initial (baseline) inventory assigned to a specific employee.
// Only one completed initial inventory can exist at a time.
func (s *AdminService) CreateInitialInventory(ctx context.Context, responsibleID uint16) (*entity.Inventory, error) {
	// Validate the employee exists and is active
	employee, err := s.employeeRepo.FindByID(ctx, responsibleID)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}
	if !employee.Active {
		return nil, fmt.Errorf("employee is not active")
	}

	// Check if an initial inventory already exists (in_progress or completed)
	existing, err := s.inventoryRepo.FindLatestByType(ctx, entity.InventoryTypeInitial)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing initial inventory: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("an initial inventory already exists")
	}

	// Also check for in-progress initial inventories
	inProgress, err := s.inventoryRepo.FindAllInProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check in-progress inventories: %w", err)
	}
	for _, inv := range inProgress {
		if inv.IsInitial() {
			return nil, fmt.Errorf("an initial inventory is already in progress")
		}
	}

	// Create the initial inventory
	now := datetime.Now()
	inventory := &entity.Inventory{
		InventoryDate: now,
		InventoryType: entity.InventoryTypeInitial,
		Schedule:      nil,
		Status:        entity.InventoryStatusInProgress,
		ResponsibleID: responsibleID,
		StartedAt:     now,
	}

	if err := s.inventoryRepo.Create(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to create initial inventory: %w", err)
	}

	// Pre-populate with ALL active items (no suggested values)
	items, err := s.itemRepo.FindActiveByInventoryType(ctx, entity.InventoryTypeInitial)
	if err != nil {
		return nil, fmt.Errorf("failed to get items: %w", err)
	}

	details := make([]*entity.InventoryDetail, 0, len(items))
	for _, item := range items {
		details = append(details, &entity.InventoryDetail{
			InventoryID: inventory.ID,
			ItemID:      item.ID,
		})
	}

	if err := s.inventoryDetailRepo.CreateBatch(ctx, details); err != nil {
		return nil, fmt.Errorf("failed to create inventory details: %w", err)
	}

	return inventory, nil
}

// ListAllActiveSuppliers retrieves all active suppliers for dropdowns.
func (s *AdminService) ListAllActiveSuppliers(ctx context.Context) ([]*entity.Supplier, error) {
	suppliers, err := s.supplierRepo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active suppliers: %w", err)
	}
	return suppliers, nil
}

// GetSupplier retrieves a single supplier by ID.
func (s *AdminService) GetSupplier(ctx context.Context, id uint16) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return nil, fmt.Errorf("supplier not found")
	}
	return supplier, nil
}

// CreateSupplier creates a new supplier.
func (s *AdminService) CreateSupplier(ctx context.Context, req CreateSupplierRequest) (*entity.Supplier, error) {
	if req.BusinessName == "" {
		return nil, fmt.Errorf("business_name is required")
	}
	if req.TaxID == "" {
		return nil, fmt.Errorf("tax_id is required")
	}

	// Check if tax_id already exists
	existing, err := s.supplierRepo.FindByTaxID(ctx, req.TaxID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tax_id: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("tax_id already exists")
	}

	supplier := &entity.Supplier{
		BusinessName: req.BusinessName,
		TaxID:        req.TaxID,
		ContactName:  req.ContactName,
		ContactPhone: req.ContactPhone,
		ContactEmail: req.ContactEmail,
		Active:       true,
	}

	if err := s.supplierRepo.Create(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to create supplier: %w", err)
	}

	return supplier, nil
}

// UpdateSupplier updates an existing supplier.
func (s *AdminService) UpdateSupplier(ctx context.Context, id uint16, req UpdateSupplierRequest) (*entity.Supplier, error) {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return nil, fmt.Errorf("supplier not found")
	}

	if req.BusinessName == "" {
		return nil, fmt.Errorf("business_name is required")
	}
	if req.TaxID == "" {
		return nil, fmt.Errorf("tax_id is required")
	}

	// Check if tax_id already exists for another supplier
	existing, err := s.supplierRepo.FindByTaxID(ctx, req.TaxID)
	if err != nil {
		return nil, fmt.Errorf("failed to check tax_id: %w", err)
	}
	if existing != nil && existing.ID != id {
		return nil, fmt.Errorf("tax_id already exists")
	}

	supplier.BusinessName = req.BusinessName
	supplier.TaxID = req.TaxID
	supplier.ContactName = req.ContactName
	supplier.ContactPhone = req.ContactPhone
	supplier.ContactEmail = req.ContactEmail
	supplier.Active = req.Active

	if err := s.supplierRepo.Update(ctx, supplier); err != nil {
		return nil, fmt.Errorf("failed to update supplier: %w", err)
	}

	return supplier, nil
}

// UpdateSupplierStatus updates the active status of a supplier.
func (s *AdminService) UpdateSupplierStatus(ctx context.Context, id uint16, active bool) error {
	supplier, err := s.supplierRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get supplier: %w", err)
	}
	if supplier == nil {
		return fmt.Errorf("supplier not found")
	}

	if err := s.supplierRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update supplier status: %w", err)
	}

	return nil
}
