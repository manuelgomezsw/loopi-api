package repository

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// CategoryRepository defines the interface for category data access.
type CategoryRepository interface {
	// FindAll retrieves all categories ordered by display_order.
	FindAll(ctx context.Context) ([]*entity.Category, error)

	// FindAllActive retrieves all active categories ordered by display_order.
	FindAllActive(ctx context.Context) ([]*entity.Category, error)

	// FindByID retrieves a category by its ID.
	FindByID(ctx context.Context, id uint16) (*entity.Category, error)

	// FindByName retrieves a category by its name.
	FindByName(ctx context.Context, name string) (*entity.Category, error)

	// Create creates a new category.
	Create(ctx context.Context, category *entity.Category) error

	// Update updates an existing category.
	Update(ctx context.Context, category *entity.Category) error

	// UpdateStatus updates the active status of a category.
	UpdateStatus(ctx context.Context, id uint16, active bool) error

	// UpdateDisplayOrders updates the display order for multiple categories.
	UpdateDisplayOrders(ctx context.Context, orders map[uint16]int) error

	// GetMaxDisplayOrder returns the maximum display order value.
	GetMaxDisplayOrder(ctx context.Context) (int, error)
}
