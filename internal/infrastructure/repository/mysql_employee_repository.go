package repository

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
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
		SELECT id, username, password_hash, name, last_name, document_type, document_number, 
		       phone, email, birth_date, role, active, created_at, updated_at
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
		&emp.DocumentType,
		&emp.DocumentNumber,
		&emp.Phone,
		&emp.Email,
		&emp.BirthDate,
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
		SELECT id, username, password_hash, name, last_name, document_type, document_number,
		       phone, email, birth_date, role, active, created_at, updated_at
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
		&emp.DocumentType,
		&emp.DocumentNumber,
		&emp.Phone,
		&emp.Email,
		&emp.BirthDate,
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
		SELECT id, username, password_hash, name, last_name, document_type, document_number,
		       phone, email, birth_date, role, active, created_at, updated_at
		FROM employees
		WHERE active = 1
		ORDER BY name, last_name
	`

	return r.queryEmployees(ctx, query)
}

// FindAllWithFilters retrieves employees with optional filters and pagination.
func (r *mysqlEmployeeRepository) FindAllWithFilters(ctx context.Context, role *entity.Role, active *bool, search string, page, pageSize int) ([]*entity.Employee, int, error) {
	baseQuery := " FROM employees WHERE 1=1"
	var args []interface{}

	if role != nil {
		baseQuery += " AND role = ?"
		args = append(args, *role)
	}
	if active != nil {
		baseQuery += " AND active = ?"
		args = append(args, *active)
	}
	if search != "" {
		baseQuery += " AND (name LIKE ? OR last_name LIKE ? OR username LIKE ?)"
		searchPattern := "%" + search + "%"
		args = append(args, searchPattern, searchPattern, searchPattern)
	}

	// Count total
	var total int
	countQuery := "SELECT COUNT(*)" + baseQuery
	if err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count employees: %w", err)
	}

	// Get paginated results
	selectQuery := `SELECT id, username, password_hash, name, last_name, document_type, document_number,
	                       phone, email, birth_date, role, active, created_at, updated_at` + baseQuery +
		` ORDER BY name, last_name LIMIT ? OFFSET ?`
	args = append(args, pageSize, (page-1)*pageSize)

	employees, err := r.queryEmployees(ctx, selectQuery, args...)
	if err != nil {
		return nil, 0, err
	}

	return employees, total, nil
}

// Create creates a new employee.
func (r *mysqlEmployeeRepository) Create(ctx context.Context, employee *entity.Employee) error {
	query := `
		INSERT INTO employees (username, password_hash, name, last_name, document_type, document_number,
		                       phone, email, birth_date, role, active)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`

	result, err := r.db.ExecContext(ctx, query,
		employee.Username,
		employee.PasswordHash,
		employee.Name,
		employee.LastName,
		employee.DocumentType,
		employee.DocumentNumber,
		employee.Phone,
		employee.Email,
		employee.BirthDate,
		employee.Role,
		employee.Active,
	)
	if err != nil {
		return fmt.Errorf("failed to create employee: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert id: %w", err)
	}

	employee.ID = uint16(id)
	return nil
}

// Update updates an existing employee.
func (r *mysqlEmployeeRepository) Update(ctx context.Context, employee *entity.Employee) error {
	query := `
		UPDATE employees
		SET username = ?, name = ?, last_name = ?, document_type = ?, document_number = ?,
		    phone = ?, email = ?, birth_date = ?, role = ?, active = ?, updated_at = ?
		WHERE id = ?
	`

	_, err := r.db.ExecContext(ctx, query,
		employee.Username,
		employee.Name,
		employee.LastName,
		employee.DocumentType,
		employee.DocumentNumber,
		employee.Phone,
		employee.Email,
		employee.BirthDate,
		employee.Role,
		employee.Active,
		datetime.Now(),
		employee.ID,
	)
	if err != nil {
		return fmt.Errorf("failed to update employee: %w", err)
	}

	return nil
}

// UpdateStatus updates the active status of an employee.
func (r *mysqlEmployeeRepository) UpdateStatus(ctx context.Context, id uint16, active bool) error {
	query := `UPDATE employees SET active = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, active, datetime.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update employee status: %w", err)
	}

	return nil
}

// UpdatePassword updates the password hash of an employee.
func (r *mysqlEmployeeRepository) UpdatePassword(ctx context.Context, id uint16, passwordHash string) error {
	query := `UPDATE employees SET password_hash = ?, updated_at = ? WHERE id = ?`

	_, err := r.db.ExecContext(ctx, query, passwordHash, datetime.Now(), id)
	if err != nil {
		return fmt.Errorf("failed to update employee password: %w", err)
	}

	return nil
}

// queryEmployees is a helper function to execute employee queries.
func (r *mysqlEmployeeRepository) queryEmployees(ctx context.Context, query string, args ...interface{}) ([]*entity.Employee, error) {
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query employees: %w", err)
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
			&emp.DocumentType,
			&emp.DocumentNumber,
			&emp.Phone,
			&emp.Email,
			&emp.BirthDate,
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
