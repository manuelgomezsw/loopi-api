package service

import (
	"context"
	"fmt"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
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

// SuggestedInventory contains suggested inventory type and schedule.
type SuggestedInventory struct {
	InventoryType entity.InventoryType `json:"inventory_type"`
	Schedule      *entity.Schedule     `json:"schedule,omitempty"`
	Date          time.Time            `json:"date"`
}

// GetSuggestedSchedule returns the suggested schedule based on current time.
func (s *InventoryService) GetSuggestedSchedule() SuggestedInventory {
	now := datetime.Now()
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

	return SuggestedInventory{
		InventoryType: entity.InventoryTypeDaily,
		Schedule:      &schedule,
		Date:          now,
	}
}

// GetLatestCompleted returns the most recent completed inventory.
func (s *InventoryService) GetLatestCompleted(ctx context.Context) (*entity.Inventory, error) {
	inv, err := s.inventoryRepo.FindLatestCompleted(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest inventory: %w", err)
	}
	return inv, nil
}

// GetInProgress returns in-progress inventories for an employee.
func (s *InventoryService) GetInProgress(ctx context.Context, employeeID uint16) ([]*entity.Inventory, error) {
	inventories, err := s.inventoryRepo.FindInProgressByEmployee(ctx, employeeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get in-progress inventories: %w", err)
	}
	return inventories, nil
}

// CreateInventory creates a new inventory and pre-populates items with suggested values.
func (s *InventoryService) CreateInventory(ctx context.Context, inventoryType entity.InventoryType, schedule *entity.Schedule, date time.Time, responsibleID uint16) (*entity.Inventory, error) {
	// Validate: daily inventories require a schedule
	if inventoryType == entity.InventoryTypeDaily && schedule == nil {
		return nil, apperrors.New(400, "schedule is required for daily inventories")
	}
	// Non-daily inventories should not have a schedule
	if inventoryType != entity.InventoryTypeDaily {
		schedule = nil
	}

	// Check if inventory already exists for this date, type and schedule
	existing, err := s.inventoryRepo.FindByDateTypeAndSchedule(ctx, date, inventoryType, schedule)
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
		InventoryType: inventoryType,
		Schedule:      schedule,
		Status:        entity.InventoryStatusInProgress,
		ResponsibleID: responsibleID,
		StartedAt:     datetime.Now(),
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
	// Get items based on inventory type
	items, err := s.itemRepo.FindActiveByInventoryType(ctx, inventory.InventoryType)
	if err != nil {
		return fmt.Errorf("failed to get items: %w", err)
	}

	// Get previous inventory to calculate suggested values
	previousInv, err := s.inventoryRepo.FindPreviousInventory(ctx, inventory.InventoryDate, inventory.InventoryType, inventory.Schedule)
	if err != nil {
		return fmt.Errorf("failed to get previous inventory: %w", err)
	}

	// Build a map of previous real value (sugerido siguiente = solo conteo anterior; mermas no restan)
	previousRealByItem := make(map[uint16]uint16)
	if previousInv != nil {
		prevDetails, err := s.inventoryDetailRepo.FindByInventoryID(ctx, previousInv.ID)
		if err != nil {
			return fmt.Errorf("failed to get previous inventory details: %w", err)
		}
		for _, d := range prevDetails {
			if d.RealValue == nil {
				continue
			}
			previousRealByItem[d.ItemID] = *d.RealValue
		}
	}

	// Create detail records
	details := make([]*entity.InventoryDetail, 0, len(items))
	for _, item := range items {
		detail := &entity.InventoryDetail{
			InventoryID: inventory.ID,
			ItemID:      item.ID,
		}

		// Set suggested value: real_anterior only (merma es control del periodo anterior, no resta del esperado del siguiente)
		if realPrev, ok := previousRealByItem[item.ID]; ok {
			detail.SuggestedValue = &realPrev
		}

		details = append(details, detail)
	}

	if err := s.inventoryDetailRepo.CreateBatch(ctx, details); err != nil {
		return fmt.Errorf("failed to create inventory details: %w", err)
	}

	return nil
}

// GetInventoryItems returns the items for an inventory with their current state.
// Suggested values are recomputed from the previous inventory (real only; mermas do not subtract from next period's expected).
func (s *InventoryService) GetInventoryItems(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	inventory, err := s.inventoryRepo.FindByID(ctx, inventoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return nil, nil, apperrors.ErrNotFound
	}

	details, err := s.inventoryDetailRepo.FindByInventoryIDWithItems(ctx, inventoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory details: %w", err)
	}

	s.enrichDetailsWithSuggestedFromPrevious(ctx, inventory, details)
	return inventory, details, nil
}

// enrichDetailsWithSuggestedFromPrevious recomputes suggested_value from the previous inventory
// (real_anterior only; mermas do not subtract from next period's expected) and overwrites each detail's SuggestedValue for response.
func (s *InventoryService) enrichDetailsWithSuggestedFromPrevious(ctx context.Context, inventory *entity.Inventory, details []*entity.InventoryDetail) {
	previousInv, err := s.inventoryRepo.FindPreviousInventory(ctx, inventory.InventoryDate, inventory.InventoryType, inventory.Schedule)
	if err != nil || previousInv == nil {
		return
	}
	prevDetails, err := s.inventoryDetailRepo.FindByInventoryID(ctx, previousInv.ID)
	if err != nil {
		return
	}
	// suggested = real_anterior (por ítem); merma es control/auditoría del periodo anterior
	realByItem := make(map[uint16]uint16)
	for _, d := range prevDetails {
		if d.RealValue != nil {
			realByItem[d.ItemID] = *d.RealValue
		}
	}
	for _, d := range details {
		if v, ok := realByItem[d.ItemID]; ok {
			d.SuggestedValue = &v
		}
	}
}

// ComputeExpectedAtEnd returns the expected count at end of period: suggested − units_sold + stock_received (clamped to 0).
// Used for summary and discrepancy logic when the employee has entered sales/purchases.
func (s *InventoryService) ComputeExpectedAtEnd(d *entity.InventoryDetail) uint16 {
	suggested := uint16(0)
	if d.SuggestedValue != nil {
		suggested = *d.SuggestedValue
	}
	unitsSold := uint16(0)
	if d.UnitsSold != nil {
		unitsSold = *d.UnitsSold
	}
	stockReceived := uint16(0)
	if d.StockReceived != nil {
		stockReceived = *d.StockReceived
	}
	expected := int32(suggested) + int32(stockReceived) - int32(unitsSold)
	if expected < 0 {
		expected = 0
	}
	return uint16(expected)
}

// HasDiscrepancyFromExpectedEnd returns true when real_value != expected_at_end (suggested − sold + received).
func (s *InventoryService) HasDiscrepancyFromExpectedEnd(d *entity.InventoryDetail) bool {
	if d.RealValue == nil {
		return false
	}
	expected := s.ComputeExpectedAtEnd(d)
	return *d.RealValue != expected
}

// GetDiscrepancies returns items with discrepancies (real_value != expected_at_end).
// Expected at end = suggested − units_sold + stock_received, so items that match after sales/purchases are not listed.
func (s *InventoryService) GetDiscrepancies(ctx context.Context, inventoryID uint32) (*entity.Inventory, []*entity.InventoryDetail, error) {
	inventory, err := s.inventoryRepo.FindByID(ctx, inventoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory: %w", err)
	}
	if inventory == nil {
		return nil, nil, apperrors.ErrNotFound
	}

	allDetails, err := s.inventoryDetailRepo.FindByInventoryIDWithItems(ctx, inventoryID)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get inventory details: %w", err)
	}
	s.enrichDetailsWithSuggestedFromPrevious(ctx, inventory, allDetails)

	details := make([]*entity.InventoryDetail, 0, len(allDetails))
	for _, d := range allDetails {
		if s.HasDiscrepancyFromExpectedEnd(d) {
			details = append(details, d)
		}
	}
	return inventory, details, nil
}

// SaveInventoryDetail saves or updates an inventory detail (physical count only).
func (s *InventoryService) SaveInventoryDetail(ctx context.Context, inventoryID uint32, itemID uint16, realValue uint16) (*entity.InventoryDetail, error) {
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

	// Update detail with physical count only
	detail.RealValue = &realValue

	if err := s.inventoryDetailRepo.Update(ctx, detail); err != nil {
		return nil, fmt.Errorf("failed to update inventory detail: %w", err)
	}

	return detail, nil
}

// SaveSalesAndPurchases saves sales and purchases for an inventory detail.
func (s *InventoryService) SaveSalesAndPurchases(ctx context.Context, inventoryID uint32, itemID uint16, stockReceived, unitsSold *uint16) (*entity.InventoryDetail, error) {
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
	if !inventory.RequiresSalesAndPurchases() {
		return nil, apperrors.New(400, "this inventory type does not require sales and purchases")
	}

	// Get existing detail
	detail, err := s.inventoryDetailRepo.FindByInventoryAndItem(ctx, inventoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory detail: %w", err)
	}
	if detail == nil {
		return nil, apperrors.New(400, "item not found in this inventory")
	}

	// Update sales and purchases
	detail.StockReceived = stockReceived
	detail.UnitsSold = unitsSold

	// Recalculate suggested value based on previous + received - sold
	if detail.SuggestedValue != nil {
		suggested := *detail.SuggestedValue
		if stockReceived != nil {
			suggested += *stockReceived
		}
		if unitsSold != nil {
			if suggested >= *unitsSold {
				suggested -= *unitsSold
			} else {
				suggested = 0
			}
		}
		detail.SuggestedValue = &suggested
	}

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

	// Create issues for discrepancies (real != expected_at_end; skip for initial inventories)
	var issues []*entity.InventoryIssue
	if !inventory.IsInitial() {
		for _, d := range details {
			if s.HasDiscrepancyFromExpectedEnd(d) {
				expectedAtEnd := s.ComputeExpectedAtEnd(d)
				diff := int16(*d.RealValue) - int16(expectedAtEnd)
				issue := &entity.InventoryIssue{
					InventoryDetailID: d.ID,
					Type:              entity.IssueTypeDiscrepancy,
					ExpectedValue:     &expectedAtEnd,
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
	}

	// Complete the inventory
	if err := s.inventoryRepo.Complete(ctx, inventoryID); err != nil {
		return 0, fmt.Errorf("failed to complete inventory: %w", err)
	}

	return len(issues), nil
}
