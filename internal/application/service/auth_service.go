package service

import (
	"context"
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/auth"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
)

// AuthService handles authentication operations.
type AuthService struct {
	employeeRepo repository.EmployeeRepository
	jwtManager   *auth.JWTManager
}

// NewAuthService creates a new auth service.
func NewAuthService(employeeRepo repository.EmployeeRepository, jwtManager *auth.JWTManager) *AuthService {
	return &AuthService{
		employeeRepo: employeeRepo,
		jwtManager:   jwtManager,
	}
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	Token    string           `json:"token"`
	Employee *entity.Employee `json:"employee"`
}

// Login authenticates an employee and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	// Find employee by username
	employee, err := s.employeeRepo.FindByUsername(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to find employee: %w", err)
	}
	if employee == nil {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Verify password
	if !auth.CheckPassword(password, employee.PasswordHash) {
		return nil, apperrors.ErrInvalidCredentials
	}

	// Generate JWT token
	token, err := s.jwtManager.GenerateToken(employee)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResult{
		Token:    token,
		Employee: employee,
	}, nil
}

// GetEmployeeByID retrieves an employee by ID.
func (s *AuthService) GetEmployeeByID(ctx context.Context, id uint16) (*entity.Employee, error) {
	employee, err := s.employeeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to find employee: %w", err)
	}
	if employee == nil {
		return nil, apperrors.ErrNotFound
	}
	return employee, nil
}
