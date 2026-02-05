package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
)

// mysqlEmployeeRepository implements repository.EmployeeRepository.
type mysqlEmployeeRepository struct {
	db *sql.DB
}

// NewMySQLEmployeeRepository creates a new MySQL employee repository.
func NewMySQLEmployeeRepository(db *sql.DB) repository.EmployeeRepository {
	return &mysqlEmployeeRepository{db: db}
}

// FindByID retrieves an employee by their ID.
func (r *mysqlEmployeeRepository) FindByID(ctx context.Context, id uint16) (*entity.Employee, error) {
	query := `
		SELECT id, username, password_hash, name, last_name, role, active, created_at, updated_at
		FROM employees
		WHERE id = ?
	`

	var emp entity.Employee
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&emp.ID,
		&emp.Username,
		&emp.PasswordHash,
		&emp.Name,
		&emp.LastName,
		&emp.Role,
		&emp.Active,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find employee by id: %w", err)
	}

	return &emp, nil
}

// FindByUsername retrieves an employee by their username.
func (r *mysqlEmployeeRepository) FindByUsername(ctx context.Context, username string) (*entity.Employee, error) {
	query := `
		SELECT id, username, password_hash, name, last_name, role, active, created_at, updated_at
		FROM employees
		WHERE username = ? AND active = 1
	`

	var emp entity.Employee
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&emp.ID,
		&emp.Username,
		&emp.PasswordHash,
		&emp.Name,
		&emp.LastName,
		&emp.Role,
		&emp.Active,
		&emp.CreatedAt,
		&emp.UpdatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to find employee by username: %w", err)
	}

	return &emp, nil
}

// FindAllActive retrieves all active employees.
func (r *mysqlEmployeeRepository) FindAllActive(ctx context.Context) ([]*entity.Employee, error) {
	query := `
		SELECT id, username, password_hash, name, last_name, role, active, created_at, updated_at
		FROM employees
		WHERE active = 1
		ORDER BY name, last_name
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to find active employees: %w", err)
	}
	defer rows.Close()

	var employees []*entity.Employee
	for rows.Next() {
		var emp entity.Employee
		if err := rows.Scan(
			&emp.ID,
			&emp.Username,
			&emp.PasswordHash,
			&emp.Name,
			&emp.LastName,
			&emp.Role,
			&emp.Active,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan employee: %w", err)
		}
		employees = append(employees, &emp)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating employees: %w", err)
	}

	return employees, nil
}
