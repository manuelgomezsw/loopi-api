package repository

import (
	"context"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

// EmployeeRepository defines the interface for employee data access.
type EmployeeRepository interface {
	// FindByID retrieves an employee by their ID.
	FindByID(ctx context.Context, id uint16) (*entity.Employee, error)

	// FindByUsername retrieves an employee by their username.
	FindByUsername(ctx context.Context, username string) (*entity.Employee, error)

	// FindAllActive retrieves all active employees.
	FindAllActive(ctx context.Context) ([]*entity.Employee, error)
}
