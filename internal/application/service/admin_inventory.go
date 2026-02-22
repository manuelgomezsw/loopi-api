package service

import (
	"context"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	invdomain "github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
	"golang.org/x/sync/errgroup"
)

// AdminInventoryService handles admin inventory operations (list, detail, update detail, initial inventory).
type AdminInventoryService struct {
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	employeeRepo        repository.EmployeeRepository
	itemRepo            repository.ItemRepository
	enricher            *invdomain.Enricher
}

// NewAdminInventoryService creates a new admin inventory service.
func NewAdminInventoryService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	employeeRepo repository.EmployeeRepository,
	itemRepo repository.ItemRepository,
	enricher *invdomain.Enricher,
) *AdminInventoryService {
	return &AdminInventoryService{
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		employeeRepo:        employeeRepo,
		itemRepo:            itemRepo,
		enricher:            enricher,
	}
}

// CountItemsWithDiscrepancy returns the count of details with real != expected (ExpectedForAdmin).
// It implements InventoryDiscrepancyCounter for the dashboard.
func (s *AdminInventoryService) CountItemsWithDiscrepancy(ctx context.Context, inv *entity.Inventory) int {
	details, err := s.inventoryDetailRepo.FindByInventoryID(ctx, inv.ID)
	if err != nil {
		return 0
	}
	s.enricher.Enrich(ctx, inv, details)
	_, withDiff := s.getInventoryCounts(details)
	return withDiff
}

// getInventoryCounts counts total items and items with discrepancy from pre-loaded details.
// It uses stored suggested_value directly (no Enricher) — suitable for the list view.
func (s *AdminInventoryService) getInventoryCounts(details []*entity.InventoryDetail) (totalItems, itemsWithDiff int) {
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

// ListInventories returns a paginated list of inventories with filters.
func (s *AdminInventoryService) ListInventories(ctx context.Context, filter InventoryFilter) (*InventoryListResult, error) {
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

	// Batch load all details for this page in a single query.
	ids := make([]uint32, len(inventories))
	for i, inv := range inventories {
		ids[i] = inv.ID
	}
	allDetails, err := s.inventoryDetailRepo.FindByInventoryIDs(ctx, ids)
	if err != nil {
		return nil, fmt.Errorf("failed to batch load inventory details: %w", err)
	}

	// Group details by inventory_id.
	detailsByInv := make(map[uint32][]*entity.InventoryDetail, len(inventories))
	for _, d := range allDetails {
		detailsByInv[d.InventoryID] = append(detailsByInv[d.InventoryID], d)
	}

	items := make([]InventoryListItem, 0, len(inventories))
	for _, inv := range inventories {
		totalItems, itemsWithDiff := s.getInventoryCounts(detailsByInv[inv.ID])
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

// GetInventoryDetail returns detailed information about an inventory.
// FindByIDWithEmployee and FindByInventoryIDWithItems are executed in parallel.
func (s *AdminInventoryService) GetInventoryDetail(ctx context.Context, inventoryID uint32) (*InventoryDetailView, error) {
	var inventory *entity.Inventory
	var details []*entity.InventoryDetail

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		inventory, err = s.inventoryRepo.FindByIDWithEmployee(gctx, inventoryID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		details, err = s.inventoryDetailRepo.FindByInventoryIDWithItems(gctx, inventoryID)
		if err != nil {
			return fmt.Errorf("failed to get inventory details: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, err
	}

	if inventory == nil {
		return nil, nil
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
// FindByID(inventory) and FindByID(detail) are executed in parallel.
func (s *AdminInventoryService) UpdateInventoryDetail(ctx context.Context, inventoryID uint32, detailID uint32, suggestedValue, realValue, stockReceived, unitsSold, shrinkage *uint16) error {
	var inv *entity.Inventory
	var detail *entity.InventoryDetail

	g, gctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		var err error
		inv, err = s.inventoryRepo.FindByID(gctx, inventoryID)
		if err != nil {
			return fmt.Errorf("failed to get inventory: %w", err)
		}
		return nil
	})
	g.Go(func() error {
		var err error
		detail, err = s.inventoryDetailRepo.FindByID(gctx, detailID)
		if err != nil {
			return fmt.Errorf("failed to get inventory detail: %w", err)
		}
		return nil
	})

	if err := g.Wait(); err != nil {
		return err
	}

	if inv == nil {
		return fmt.Errorf("inventory not found")
	}
	if detail == nil || detail.InventoryID != inventoryID {
		return fmt.Errorf("detail not found")
	}

	if suggestedValue != nil {
		detail.SuggestedValue = suggestedValue
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

// GetActiveInventoriesCount returns the count of in-progress inventories.
func (s *AdminInventoryService) GetActiveInventoriesCount(ctx context.Context) (int, error) {
	count, err := s.inventoryRepo.CountInProgress(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to count active inventories: %w", err)
	}
	return count, nil
}

// CreateInitialInventory creates an initial (baseline) inventory assigned to a specific employee.
func (s *AdminInventoryService) CreateInitialInventory(ctx context.Context, responsibleID uint16) (*entity.Inventory, error) {
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

	existing, err := s.inventoryRepo.FindLatestByType(ctx, entity.InventoryTypeInitial)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing initial inventory: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("an initial inventory already exists")
	}

	inProgress, err := s.inventoryRepo.FindAllInProgress(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check in-progress inventories: %w", err)
	}
	for _, inv := range inProgress {
		if inv.IsInitial() {
			return nil, fmt.Errorf("an initial inventory is already in progress")
		}
	}

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
