package repository

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// MeasurementUnitRepository defines the interface for measurement unit data access.
type MeasurementUnitRepository interface {
	// FindAll retrieves all measurement units ordered by name (alphabetical).
	FindAll(ctx context.Context) ([]*entity.MeasurementUnit, error)

	// FindByID retrieves a measurement unit by its ID.
	FindByID(ctx context.Context, id uint16) (*entity.MeasurementUnit, error)
}
