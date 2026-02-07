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

	// FindAllWithFilters retrieves employees with optional filters and pagination.
	FindAllWithFilters(ctx context.Context, role *entity.Role, active *bool, search string, page, pageSize int) ([]*entity.Employee, int, error)

	// Create creates a new employee.
	Create(ctx context.Context, employee *entity.Employee) error

	// Update updates an existing employee.
	Update(ctx context.Context, employee *entity.Employee) error

	// UpdateStatus updates the active status of an employee.
	UpdateStatus(ctx context.Context, id uint16, active bool) error

	// UpdatePassword updates the password hash of an employee.
	UpdatePassword(ctx context.Context, id uint16, passwordHash string) error
}
