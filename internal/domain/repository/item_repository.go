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
}
