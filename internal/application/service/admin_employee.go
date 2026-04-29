package service

import (
	"context"
	"fmt"

	employeedomain "github.com/manuelgomezsw/loopi-api/internal/domain/employee"
	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/pkg/datetime"
	"github.com/manuelgomezsw/loopi-api/pkg/logger"
	"golang.org/x/crypto/bcrypt"
)

// AdminEmployeeService handles admin employee operations (CRUD, reset password, list active).
type AdminEmployeeService struct {
	employeeRepo repository.EmployeeRepository
}

// NewAdminEmployeeService creates a new admin employee service.
func NewAdminEmployeeService(employeeRepo repository.EmployeeRepository) *AdminEmployeeService {
	return &AdminEmployeeService{employeeRepo: employeeRepo}
}

func hashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// ListEmployees retrieves employees with filters and pagination.
func (s *AdminEmployeeService) ListEmployees(ctx context.Context, filter EmployeeFilter) (*EmployeeListResult, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 || filter.PageSize > 100 {
		filter.PageSize = 20
	}

	employees, total, err := s.employeeRepo.FindAllWithFilters(ctx, filter.Role, filter.Active, filter.Search, filter.Page, filter.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to list employees: %w", err)
	}

	totalPages := total / filter.PageSize
	if total%filter.PageSize > 0 {
		totalPages++
	}

	return &EmployeeListResult{
		Employees:  employees,
		Total:      total,
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		TotalPages: totalPages,
	}, nil
}

// GetEmployee retrieves a single employee by ID.
func (s *AdminEmployeeService) GetEmployee(ctx context.Context, id uint16) (*entity.Employee, error) {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}
	return employee, nil
}

// CreateEmployee creates a new employee.
func (s *AdminEmployeeService) CreateEmployee(ctx context.Context, req CreateEmployeeRequest) (*entity.Employee, error) {
	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Password == "" {
		return nil, fmt.Errorf("password is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
	}
	if req.Role != entity.RoleEmployee && req.Role != entity.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}

	existing, err := s.employeeRepo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("username already exists")
	}

	passwordHash, err := hashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	employee := &entity.Employee{
		Username:       req.Username,
		PasswordHash:   passwordHash,
		Name:           req.Name,
		LastName:       req.LastName,
		DocumentType:   req.DocumentType,
		DocumentNumber: req.DocumentNumber,
		Phone:          req.Phone,
		Email:          req.Email,
		BirthDate:      req.BirthDate,
		Role:           req.Role,
		Active:         true,
		CreatedAt:      datetime.Now(),
		UpdatedAt:      datetime.Now(),
	}

	if err := s.employeeRepo.Create(ctx, employee); err != nil {
		logger.FromContext(ctx).ErrorContext(ctx, "failed to create employee",
			"operation", "adminEmployee.Create", "username", req.Username, "error", err)
		return nil, fmt.Errorf("failed to create employee: %w", err)
	}

	logger.FromContext(ctx).InfoContext(ctx, "employee created",
		"operation", "adminEmployee.Create", "employee_id", employee.ID, "username", employee.Username, "role", employee.Role)
	return employee, nil
}

// UpdateEmployee updates an existing employee.
func (s *AdminEmployeeService) UpdateEmployee(ctx context.Context, id uint16, req UpdateEmployeeRequest) (*entity.Employee, error) {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return nil, fmt.Errorf("employee not found")
	}

	if req.Username == "" {
		return nil, fmt.Errorf("username is required")
	}
	if req.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if req.LastName == "" {
		return nil, fmt.Errorf("last name is required")
	}
	if req.Role != entity.RoleEmployee && req.Role != entity.RoleAdmin {
		return nil, fmt.Errorf("invalid role")
	}

	employee.Username = req.Username
	employee.Name = req.Name
	employee.LastName = req.LastName
	employee.DocumentType = req.DocumentType
	employee.DocumentNumber = req.DocumentNumber
	employee.Phone = req.Phone
	employee.Email = req.Email
	employee.BirthDate = req.BirthDate
	employee.Role = req.Role
	employee.Active = req.Active
	employee.UpdatedAt = datetime.Now()

	if err := s.employeeRepo.Update(ctx, employee); err != nil {
		return nil, fmt.Errorf("failed to update employee: %w", err)
	}

	return employee, nil
}

// UpdateEmployeeStatus updates the active status of an employee.
func (s *AdminEmployeeService) UpdateEmployeeStatus(ctx context.Context, id uint16, active bool) error {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return fmt.Errorf("employee not found")
	}

	if err := s.employeeRepo.UpdateStatus(ctx, id, active); err != nil {
		return fmt.Errorf("failed to update employee status: %w", err)
	}

	return nil
}

// ResetEmployeePassword resets an employee's password to default (domain rule).
func (s *AdminEmployeeService) ResetEmployeePassword(ctx context.Context, id uint16) error {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get employee: %w", err)
	}
	if employee == nil {
		return fmt.Errorf("employee not found")
	}

	defaultPassword := employeedomain.DefaultPasswordForReset(employee)
	passwordHash, err := hashPassword(defaultPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	if err := s.employeeRepo.UpdatePassword(ctx, id, passwordHash); err != nil {
		logger.FromContext(ctx).ErrorContext(ctx, "failed to reset employee password",
			"operation", "adminEmployee.ResetPassword", "employee_id", id, "error", err)
		return fmt.Errorf("failed to update password: %w", err)
	}

	logger.FromContext(ctx).InfoContext(ctx, "employee password reset",
		"operation", "adminEmployee.ResetPassword", "employee_id", id)
	return nil
}

// ListAllActiveEmployees retrieves all active employees for dropdowns.
func (s *AdminEmployeeService) ListAllActiveEmployees(ctx context.Context) ([]*entity.Employee, error) {
	employees, err := s.employeeRepo.FindAllActive(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list active employees: %w", err)
	}
	return employees, nil
}
