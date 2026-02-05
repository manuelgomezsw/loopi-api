package repository

import (
	"context"
	"time"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// InventoryRepository defines the interface for inventory data access.
type InventoryRepository interface {
	// FindByID retrieves an inventory by its ID.
	FindByID(ctx context.Context, id uint32) (*entity.Inventory, error)

	// FindByDateTypeAndSchedule retrieves an inventory by date, type and schedule.
	FindByDateTypeAndSchedule(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error)

	// FindLatestCompleted retrieves the most recent completed inventory.
	FindLatestCompleted(ctx context.Context) (*entity.Inventory, error)

	// FindLatestByType retrieves the most recent completed inventory for a specific type.
	FindLatestByType(ctx context.Context, inventoryType entity.InventoryType) (*entity.Inventory, error)

	// FindPreviousInventory retrieves the previous inventory for calculating suggested values.
	FindPreviousInventory(ctx context.Context, date time.Time, inventoryType entity.InventoryType, schedule *entity.Schedule) (*entity.Inventory, error)

	// Create creates a new inventory.
	Create(ctx context.Context, inventory *entity.Inventory) error

	// Update updates an existing inventory.
	Update(ctx context.Context, inventory *entity.Inventory) error

	// Complete marks an inventory as completed.
	Complete(ctx context.Context, id uint32) error
}

// InventoryDetailRepository defines the interface for inventory detail data access.
type InventoryDetailRepository interface {
	// FindByInventoryID retrieves all details for an inventory.
	FindByInventoryID(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error)

	// FindByInventoryIDWithItems retrieves all details with item information.
	FindByInventoryIDWithItems(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error)

	// FindByInventoryAndItem retrieves a specific detail by inventory and item.
	FindByInventoryAndItem(ctx context.Context, inventoryID uint32, itemID uint16) (*entity.InventoryDetail, error)

	// FindDiscrepancies retrieves details where real_value differs from suggested_value.
	FindDiscrepancies(ctx context.Context, inventoryID uint32) ([]*entity.InventoryDetail, error)

	// Create creates a new inventory detail.
	Create(ctx context.Context, detail *entity.InventoryDetail) error

	// Update updates an existing inventory detail.
	Update(ctx context.Context, detail *entity.InventoryDetail) error

	// Upsert creates or updates an inventory detail.
	Upsert(ctx context.Context, detail *entity.InventoryDetail) error

	// CreateBatch creates multiple inventory details at once.
	CreateBatch(ctx context.Context, details []*entity.InventoryDetail) error
}
