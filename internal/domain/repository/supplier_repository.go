package repository

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// SupplierRepository defines the interface for supplier data access.
type SupplierRepository interface {
	// FindAll retrieves all suppliers ordered by business_name.
	FindAll(ctx context.Context) ([]*entity.Supplier, error)

	// FindAllActive retrieves all active suppliers ordered by business_name.
	FindAllActive(ctx context.Context) ([]*entity.Supplier, error)

	// FindAllWithFilters retrieves suppliers with pagination and filters.
	FindAllWithFilters(ctx context.Context, active *bool, search string, page, pageSize int) ([]*entity.Supplier, int, error)

	// FindByID retrieves a supplier by its ID.
	FindByID(ctx context.Context, id uint16) (*entity.Supplier, error)

	// FindByTaxID retrieves a supplier by its tax ID (NIT).
	FindByTaxID(ctx context.Context, taxID string) (*entity.Supplier, error)

	// Create creates a new supplier.
	Create(ctx context.Context, supplier *entity.Supplier) error

	// Update updates an existing supplier.
	Update(ctx context.Context, supplier *entity.Supplier) error

	// UpdateStatus updates the active status of a supplier.
	UpdateStatus(ctx context.Context, id uint16, active bool) error
}
