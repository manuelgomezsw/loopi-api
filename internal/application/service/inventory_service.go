package service

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	invdomain "github.com/manuelgomezsw/loopi-api/internal/domain/inventory"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
	"github.com/manuelgomezsw/loopi-api/pkg/logger"
)

// InventoryService handles inventory operations.
type InventoryService struct {
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	itemRepo            repository.ItemRepository
	enricher            *invdomain.Enricher
	movementsCounter    metric.Int64Counter
}

// NewInventoryService creates a new inventory service.
func NewInventoryService(
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
	itemRepo repository.ItemRepository,
	enricher *invdomain.Enricher,
) *InventoryService {
	meter := otel.GetMeterProvider().Meter("loopi-api")
	movementsCounter, _ := meter.Int64Counter(
		"loopi.inventory.movements",
		metric.WithDescription("Total de movimientos de inventario (create/complete)"),
		metric.WithUnit("{movement}"),
	)
	return &InventoryService{
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		itemRepo:            itemRepo,
		enricher:            enricher,
		movementsCounter:    movementsCounter,
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
		logger.FromContext(ctx).ErrorContext(ctx, "failed to create inventory",
			"operation", "inventory.Create", "error", err)
		return nil, fmt.Errorf("failed to create inventory: %w", err)
	}

	if err := s.prepopulateInventoryDetails(ctx, inventory); err != nil {
		return nil, fmt.Errorf("failed to prepopulate inventory details: %w", err)
	}

	logger.FromContext(ctx).InfoContext(ctx, "inventory created",
		"operation", "inventory.Create",
		"inventory_id", inventory.ID,
		"type", inventory.InventoryType,
		"responsible_id", inventory.ResponsibleID,
	)
	s.movementsCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("action", "create"),
			attribute.String("type", string(inventoryType)),
		),
	)
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

		// Set suggested value: real_anterior only
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

	s.enricher.Enrich(ctx, inventory, details)
	return inventory, details, nil
}

// ComputeExpectedAtEnd returns the expected count at end of period.
func (s *InventoryService) ComputeExpectedAtEnd(d *entity.InventoryDetail) uint16 {
	return invdomain.ExpectedAtEnd(d)
}

// HasDiscrepancyFromExpectedEnd returns true when real_value != expected_at_end.
func (s *InventoryService) HasDiscrepancyFromExpectedEnd(d *entity.InventoryDetail) bool {
	expected := invdomain.ExpectedAtEnd(d)
	return invdomain.HasDiscrepancyFromExpectedEnd(d, expected)
}

// GetDiscrepancies returns items with discrepancies (real_value != expected_at_end).
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
	s.enricher.Enrich(ctx, inventory, allDetails)

	details := make([]*entity.InventoryDetail, 0, len(allDetails))
	for _, d := range allDetails {
		expected := invdomain.ExpectedAtEnd(d)
		if invdomain.HasDiscrepancyFromExpectedEnd(d, expected) {
			details = append(details, d)
		}
	}
	return inventory, details, nil
}

// SaveInventoryDetail saves or updates an inventory detail (physical count only).
func (s *InventoryService) SaveInventoryDetail(ctx context.Context, inventoryID uint32, itemID uint16, realValue uint16) (*entity.InventoryDetail, error) {
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

	detail, err := s.inventoryDetailRepo.FindByInventoryAndItem(ctx, inventoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory detail: %w", err)
	}
	if detail == nil {
		return nil, apperrors.New(400, "item not found in this inventory")
	}

	detail.RealValue = &realValue

	if err := s.inventoryDetailRepo.Update(ctx, detail); err != nil {
		return nil, fmt.Errorf("failed to update inventory detail: %w", err)
	}

	return detail, nil
}

// SaveSalesAndPurchases saves sales and purchases for an inventory detail.
func (s *InventoryService) SaveSalesAndPurchases(ctx context.Context, inventoryID uint32, itemID uint16, stockReceived, unitsSold *uint16) (*entity.InventoryDetail, error) {
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

	detail, err := s.inventoryDetailRepo.FindByInventoryAndItem(ctx, inventoryID, itemID)
	if err != nil {
		return nil, fmt.Errorf("failed to get inventory detail: %w", err)
	}
	if detail == nil {
		return nil, apperrors.New(400, "item not found in this inventory")
	}

	unitsToApply := unitsSold
	if inventory.RequiresPurchasesOnly() {
		zero := uint16(0)
		unitsToApply = &zero
	}

	detail.StockReceived = stockReceived
	detail.UnitsSold = unitsToApply

	if detail.SuggestedValue != nil {
		suggested := *detail.SuggestedValue
		if stockReceived != nil {
			suggested += *stockReceived
		}
		if unitsToApply != nil && *unitsToApply > 0 {
			if suggested >= *unitsToApply {
				suggested -= *unitsToApply
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

// CompleteInventory marks an inventory as completed.
func (s *InventoryService) CompleteInventory(ctx context.Context, inventoryID uint32) (int, error) {
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

	details, err := s.inventoryDetailRepo.FindByInventoryID(ctx, inventoryID)
	if err != nil {
		return 0, fmt.Errorf("failed to get inventory details: %w", err)
	}

	s.enricher.Enrich(ctx, inventory, details)

	for _, d := range details {
		if !d.IsComplete() {
			return 0, apperrors.New(400, "not all items have been inventoried")
		}
	}

	discrepancyCount := 0
	if !inventory.IsInitial() {
		for _, d := range details {
			expectedAtEnd := invdomain.ExpectedAtEnd(d)
			if invdomain.HasDiscrepancyFromExpectedEnd(d, expectedAtEnd) {
				discrepancyCount++
			}
		}
	}

	if err := s.inventoryRepo.Complete(ctx, inventoryID); err != nil {
		logger.FromContext(ctx).ErrorContext(ctx, "failed to complete inventory",
			"operation", "inventory.Complete", "inventory_id", inventoryID, "error", err)
		return 0, fmt.Errorf("failed to complete inventory: %w", err)
	}

	logger.FromContext(ctx).InfoContext(ctx, "inventory completed",
		"operation", "inventory.Complete",
		"inventory_id", inventoryID,
		"discrepancies", discrepancyCount,
	)
	s.movementsCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("action", "complete"),
			attribute.String("type", string(inventory.InventoryType)),
		),
	)
	return discrepancyCount, nil
}
