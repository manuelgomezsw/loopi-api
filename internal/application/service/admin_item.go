package service

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
	"github.com/manuelgomezsw/loopi-api/pkg/logger"
)

// AdminItemService handles admin item operations (CRUD and add to active inventories).
type AdminItemService struct {
	itemRepo            repository.ItemRepository
	measurementUnitRepo repository.MeasurementUnitRepository
	inventoryRepo       repository.InventoryRepository
	inventoryDetailRepo repository.InventoryDetailRepository
	itemsCreatedCounter metric.Int64Counter
}

// NewAdminItemService creates a new admin item service.
func NewAdminItemService(
	itemRepo repository.ItemRepository,
	measurementUnitRepo repository.MeasurementUnitRepository,
	inventoryRepo repository.InventoryRepository,
	inventoryDetailRepo repository.InventoryDetailRepository,
) *AdminItemService {
	meter := otel.GetMeterProvider().Meter("loopi-api")
	itemsCreatedCounter, _ := meter.Int64Counter(
		"loopi.items.created",
		metric.WithDescription("Total de ítems creados por tipo"),
		metric.WithUnit("{item}"),
	)
	return &AdminItemService{
		itemRepo:            itemRepo,
		measurementUnitRepo: measurementUnitRepo,
		inventoryRepo:       inventoryRepo,
		inventoryDetailRepo: inventoryDetailRepo,
		itemsCreatedCounter: itemsCreatedCounter,
	}
}

func shouldIncludeItem(inventoryType entity.InventoryType, itemFrequency entity.InventoryFrequency) bool {
	switch inventoryType {
	case entity.InventoryTypeDaily:
		return itemFrequency == entity.InventoryFrequencyDaily
	case entity.InventoryTypeWeekly:
		return itemFrequency == entity.InventoryFrequencyWeekly
	case entity.InventoryTypeMonthly:
		return itemFrequency == entity.InventoryFrequencyMonthly
	case entity.InventoryTypeInitial:
		return true
	default:
		return false
	}
}

// ListItems retrieves items with filters and pagination.
func (s *AdminItemService) ListItems(ctx context.Context, filter ItemFilter) (*ItemListResult, error) {
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
func (s *AdminItemService) GetItem(ctx context.Context, id uint16) (*entity.Item, error) {
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
func (s *AdminItemService) CreateItem(ctx context.Context, req CreateItemRequest) (*entity.Item, error) {
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
	if req.MeasurementUnitID == 0 {
		return nil, fmt.Errorf("measurement unit is required")
	}
	unit, err := s.measurementUnitRepo.FindByID(ctx, req.MeasurementUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate measurement unit: %w", err)
	}
	if unit == nil {
		return nil, fmt.Errorf("invalid measurement unit")
	}

	item := &entity.Item{
		Type:               req.Type,
		Name:               req.Name,
		Active:             true,
		InventoryFrequency: req.InventoryFrequency,
		CategoryID:         req.CategoryID,
		SupplierID:         req.SupplierID,
		Cost:               req.Cost,
		MeasurementUnitID:  req.MeasurementUnitID,
		CreatedAt:          datetime.Now(),
		UpdatedAt:          datetime.Now(),
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to create item: %w", err)
	}

	s.itemsCreatedCounter.Add(ctx, 1,
		metric.WithAttributes(
			attribute.String("type", string(item.Type)),
			attribute.String("frequency", string(item.InventoryFrequency)),
		),
	)

	if req.AddToActiveInventories {
		if err := s.addItemToActiveInventories(ctx, item); err != nil {
			logger.FromContext(ctx).WarnContext(ctx, "failed to add item to active inventories",
				"operation", "adminItem.Create", "item_id", item.ID, "error", err)
		}
	}

	logger.FromContext(ctx).InfoContext(ctx, "item created",
		"operation", "adminItem.Create", "item_id", item.ID, "name", item.Name)
	return item, nil
}

func (s *AdminItemService) addItemToActiveInventories(ctx context.Context, item *entity.Item) error {
	inventories, err := s.inventoryRepo.FindAllInProgress(ctx)
	if err != nil {
		return fmt.Errorf("failed to find active inventories: %w", err)
	}

	for _, inv := range inventories {
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

// UpdateItem updates an existing item.
func (s *AdminItemService) UpdateItem(ctx context.Context, id uint16, req UpdateItemRequest) (*entity.Item, error) {
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
	if req.MeasurementUnitID == 0 {
		return nil, fmt.Errorf("measurement unit is required")
	}
	unit, err := s.measurementUnitRepo.FindByID(ctx, req.MeasurementUnitID)
	if err != nil {
		return nil, fmt.Errorf("failed to validate measurement unit: %w", err)
	}
	if unit == nil {
		return nil, fmt.Errorf("invalid measurement unit")
	}

	item.Type = req.Type
	item.Name = req.Name
	item.InventoryFrequency = req.InventoryFrequency
	item.Active = req.Active
	item.CategoryID = req.CategoryID
	item.SupplierID = req.SupplierID
	item.Cost = req.Cost
	item.MeasurementUnitID = req.MeasurementUnitID
	item.UpdatedAt = datetime.Now()

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, fmt.Errorf("failed to update item: %w", err)
	}

	return item, nil
}

// ListMeasurementUnits returns all measurement units ordered by name (alphabetical).
func (s *AdminItemService) ListMeasurementUnits(ctx context.Context) ([]*entity.MeasurementUnit, error) {
	return s.measurementUnitRepo.FindAll(ctx)
}

// UpdateItemStatus updates the active status of an item.
func (s *AdminItemService) UpdateItemStatus(ctx context.Context, id uint16, active bool) error {
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
