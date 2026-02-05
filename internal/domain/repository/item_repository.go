package repository

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// ItemRepository defines the interface for item data access.
type ItemRepository interface {
	// FindByID retrieves an item by its ID.
	FindByID(ctx context.Context, id uint16) (*entity.Item, error)

	// FindAllActive retrieves all active items.
	FindAllActive(ctx context.Context) ([]*entity.Item, error)

	// FindActiveByType retrieves all active items of a specific type.
	FindActiveByType(ctx context.Context, itemType entity.ItemType) ([]*entity.Item, error)

	// FindActiveByInventoryType retrieves active items based on inventory type.
	// - daily: items with inventory_frequency = 'daily'
	// - weekly: items with inventory_frequency IN ('daily', 'weekly')
	// - monthly: all active items
	FindActiveByInventoryType(ctx context.Context, inventoryType entity.InventoryType) ([]*entity.Item, error)
}
