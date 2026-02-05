package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
)

// InventoryService handles inventory operations.
type InventoryService struct {
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	inventoryIssueRepo  repository.InventoryIssueRepository
	itemRepo            repository.ItemRepository
}

// NewInventoryService creates a new inventory service.
func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	inventoryIssueRepo repository.InventoryIssueRepository,
	itemRepo repository.ItemRepository,
) *InventoryService {
	return &InventoryService{
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		inventoryIssueRepo:  inventoryIssueRepo,
		itemRepo:            itemRepo,
	}
}

// GetSuggestedSchedule returns the suggested schedule based on current time.
func (s *InventoryService) GetSuggestedSchedule() (entity.Schedule, time.Time) {
	now := time.Now()
	hour := now.Hour()

	var schedule entity.Schedule
	switch {
	case hour >= 6 && hour < 11:
		schedule = entity.ScheduleOpening
	case hour >= 11 && hour < 16:
		schedule = entity.ScheduleNoon
	case hour >= 16 && hour < 22:
		schedule = entity.ScheduleClosing
	default:
		// Outside normal hours, default to opening
		schedule = entity.ScheduleOpening
	}

	return schedule, now
}

// GetLatestCompleted returns the most recent completed inventory.
func (s *InventoryService) GetLatestCompleted(ctx context.Context) (*entity.Inventory, error) {
	inv, err := s.inventoryRepo.FindLatestCompleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest inventory: %w", err)
	}
	return inv, nil
}

// CreateInventory creates a new inventory and pre-populates items with suggested values.
func (s *InventoryService) CreateInventory(ctx context.Context, schedule entity.Schedule, date time.Time, responsibleID uint16) (*entity.Inventory, error) {
	// Check if inventory already exists for this date and schedule
	existing, err := s.inventoryRepo.FindByDateAndSchedule(ctx, date, schedule)
	if err != nil {
		return nil, fmt.Errorf("failed to check existing inventory: %w", err)
	}
	if existing != nil {
		// Return existing inventory if it's still in progress
		if existing.Status == entity.InventoryStatusInProgress {
			return existing, nil
		}
		return nil, apperrors.ErrConflict
	}

	// Create new inventory
	inventory := &entity.Inventory{
		InventoryDate: date,
		Schedule:      schedule,
		Status:        entity.InventoryStatusInProgress,
		ResponsibleID: responsibleID,
		StartedAt:     time.Now(),
	}

	if err := s.inventoryRepo.Create(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}

	// Pre-populate inventory details with items and suggested values
	if err := s.prepopulateInventoryDetails(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to prepopulate inventory details: %w", err)
	}

	return inventory, nil
}

// prepopulateInventoryDetails creates inventory detail records for all applicable items.
func (s *InventoryService) prepopulateInventoryDetails(ctx context.Context, inventory *entity.Inventory) error {
	// Determine which items to include based on schedule
	var items []*entity.Item
	var err error

	switch inventory.Schedule {
	case entity.ScheduleOpening, entity.ScheduleNoon, entity.ScheduleClosing:
		// Daily inventories only include products
		items, err = s.itemRepo.FindActiveByType(ctx, entity.ItemTypeProduct)
	case entity.ScheduleWeekly:
		// Weekly inventories only include supplies
		items, err = s.itemRepo.FindActiveByType(ctx, entity.ItemTypeSupply)
	case entity.ScheduleMonthly:
		// Monthly inventories include all items
		items, err = s.itemRepo.FindAllActive(ctx)
	}

	if err != nil {
		return fmt.Errorf("failed to get items: %w", err)
	}

	// Get previous inventory to calculate suggested values
	previousInv, err := s.inventoryRepo.FindPreviousSchedule(ctx, inventory.InventoryDate, inventory.Schedule)
	if err != nil {
		return fmt.Errorf("failed to get previous inventory: %w", err)
	}

	// Build a map of previous values if available
	previousValues := make(map[uint16]uint16)
	if previousInv != nil {
		prevDetails, err := s.inventoryDetailRepo.FindByInventoryID(ctx, previousInv.ID)
		if err != nil {
			return fmt.Errorf("failed to get previous inventory details: %w", err)
		}
		for _, d := range prevDetails {
			if d.RealValue != nil {
				previousValues[d.ItemID] = *d.RealValue
			}
		}
	}

	// Create detail records
	details := make([]*entity.InventoryDetail, 0, len(items))
	for _, item := range items {
		detail := &entity.InventoryDetail{
			InventoryID: inventory.ID,
			ItemID:      item.ID,
		}

		// Set suggested value from previous inventory
		if prevValue, ok := previousValues[item.ID]; ok {
			detail.SuggestedValue = &prevValue
		}

		details = append(details, detail)
	}

	if err := s.inventoryDetailRepo.CreateBatch(ctx, details); err != nil {
		return fmt.Errorf("failed to create inventory details: %w", err)
	}

	return nil
}

// GetInventoryItems returns the items for an inventory with their current state.
func (s *InventoryService) GetInventoryItems(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	// Get inventory
	inventory, err := s.inventoryRepo.FindByID(ctx, inventoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return nil, nil, apperrors.ErrNotFound
	}

	// Get details with items
	details, err := s.inventoryDetailRepo.FindByInventoryIDWithItems(ctx, inventoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory details: %w", err)
	}

	return inventory, details, nil
}

// SaveInventoryDetail saves or updates an inventory detail.
func (s *InventoryService) SaveInventoryDetail(ctx context.Context, inventoryID uint32, itemID uint16, realValue uint16, stockReceived, unitsSold *uint16) (*entity.InventoryDetail, error) {
	// Get inventory to verify it exists and is in progress
	inventory, err := s.inventoryRepo.FindByID(ctx, inventoryID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return nil, apperrors.ErrNotFound
	}
	if inventory.IsCompleted() {
		return nil, apperrors.New(400, "inventory is already completed")
	}

	// Get existing detail
	detail, err := s.inventoryDetailRepo.FindByInventoryAndItem(ctx, inventoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory detail: %w", err)
	}
	if detail == nil {
		return nil, apperrors.New(400, "item not found in this inventory")
	}

	// Calculate suggested value for noon/closing based on previous values + purchases - sales
	if inventory.RequiresSalesAndPurchases() && detail.SuggestedValue != nil {
		// Recalculate suggested value: previous + received - sold
		suggested := *detail.SuggestedValue
		if stockReceived != nil {
			suggested += *stockReceived
		}
		if unitsSold != nil {
			suggested -= *unitsSold
		}
		detail.SuggestedValue = &suggested
	}

	// Update detail
	detail.RealValue = &realValue
	detail.StockReceived = stockReceived
	detail.UnitsSold = unitsSold

	if err := s.inventoryDetailRepo.Update(ctx, detail); err != nil {
		return nil, fmt.Errorf("failed to update inventory detail: %w", err)
	}

	return detail, nil
}

// GetInventorySummary returns a summary of the inventory before completion.
func (s *InventoryService) GetInventorySummary(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	return s.GetInventoryItems(ctx, inventoryID)
}

// CompleteInventory marks an inventory as completed and creates issues for discrepancies.
func (s *InventoryService) CompleteInventory(ctx context.Context, inventoryID uint32) (int, error) {
	// Get inventory
	inventory, err := s.inventoryRepo.FindByID(ctx, inventoryID)
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return 0, apperrors.ErrNotFound
	}
	if inventory.IsCompleted() {
		return 0, apperrors.New(400, "inventory is already completed")
	}

	// Get all details
	details, err := s.inventoryDetailRepo.FindByInventoryID(ctx, inventoryID)
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory details: %w", err)
	}

	// Check if all items have been completed
	for _, d := range details {
		if !d.IsComplete() {
			return 0, apperrors.New(400, "not all items have been inventoried")
		}
	}

	// Create issues for discrepancies
	var issues []*entity.InventoryIssue
	for _, d := range details {
		if d.HasDiscrepancy() {
			diff := d.Difference()
			issue := &entity.InventoryIssue{
				InventoryDetailID: d.ID,
				Type:              entity.IssueTypeDiscrepancy,
				ExpectedValue:     d.SuggestedValue,
				ActualValue:       d.RealValue,
				Difference:        &diff,
				Status:            entity.IssueStatusOpen,
			}
			issues = append(issues, issue)
		}
	}

	// Save issues
	if len(issues) > 0 {
		if err := s.inventoryIssueRepo.CreateBatch(ctx, issues); err != nil {
			return 0, fmt.Errorf("failed to create issues: %w", err)
		}
	}

	// Complete the inventory
	if err := s.inventoryRepo.Complete(ctx, inventoryID); err != nil {
		return 0, fmt.Errorf("failed to complete inventory: %w", err)
	}

	return len(issues), nil
}
