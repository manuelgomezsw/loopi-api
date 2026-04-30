package service

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
	"github.com/manuelgomezsw/loopi-api/internal/domain/repository"
	"github.com/manuelgomezsw/loopi-api/internal/infrastructure/auth"
	apperrors "github.com/manuelgomezsw/loopi-api/pkg/errors"
	"github.com/manuelgomezsw/loopi-api/pkg/logger"
)

// AuthService handles authentication operations.
type AuthService struct {
	employeeRepo repository.EmployeeRepository
	jwtManager   *auth.JWTManager
	loginCounter metric.Int64Counter
}

// NewAuthService creates a new auth service.
func NewAuthService(employeeRepo repository.EmployeeRepository, jwtManager *auth.JWTManager) *AuthService {
	meter := otel.GetMeterProvider().Meter("loopi-api")
	loginCounter, _ := meter.Int64Counter(
		"loopi.auth.logins",
		metric.WithDescription("Total de intentos de login"),
		metric.WithUnit("{login}"),
	)
	return &AuthService{
		employeeRepo: employeeRepo,
		jwtManager:   jwtManager,
		loginCounter: loginCounter,
	}
}

// LoginResult contains the result of a successful login.
type LoginResult struct {
	Token    string           `json:"token"`
	Employee *entity.Employee `json:"employee"`
}

// Login authenticates an employee and returns a JWT token.
func (s *AuthService) Login(ctx context.Context, username, password string) (*LoginResult, error) {
	log := logger.FromContext(ctx)

	employee, err := s.employeeRepo.FindByUsername(ctx, username)
	if err != nil {
		log.ErrorContext(ctx, "find employee failed", "operation", "auth.Login", "error", err)
		s.loginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "error")))
		return nil, fmt.Errorf("failed to find employee: %w", err)
	}
	if employee == nil || !auth.CheckPassword(password, employee.PasswordHash) {
		log.WarnContext(ctx, "invalid credentials attempt", "operation", "auth.Login", "username", username)
		s.loginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "invalid_credentials")))
		return nil, apperrors.ErrInvalidCredentials
	}

	token, err := s.jwtManager.GenerateToken(employee)
	if err != nil {
		log.ErrorContext(ctx, "token generation failed", "operation", "auth.Login", "employee_id", employee.ID, "error", err)
		s.loginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "error")))
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	log.InfoContext(ctx, "user logged in", "operation", "auth.Login", "employee_id", employee.ID, "role", employee.Role)
	s.loginCounter.Add(ctx, 1, metric.WithAttributes(attribute.String("result", "success")))
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
