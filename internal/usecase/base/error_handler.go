package base

import (
	"errors"
	"fmt"
	"strings"

	appErr "loopi-api/internal/common/errors"
	mysqlErrors "loopi-api/internal/repository/mysql"
)

// UseCaseError represents errors specific to use case layer
type UseCaseError struct {
	Code      string
	Message   string
	Status    int
	Operation string
	Entity    string
}

func (e *UseCaseError) Error() string {
	return fmt.Sprintf("usecase error [%s.%s]: %s", e.Entity, e.Operation, e.Message)
}

// ErrorHandler provides standardized error handling for use cases
type ErrorHandler struct {
	entityName string
}

// NewErrorHandler creates a new error handler for a specific entity
func NewErrorHandler(entityName string) *ErrorHandler {
	return &ErrorHandler{entityName: entityName}
}

// HandleRepositoryError converts repository errors to appropriate use case errors
func (h *ErrorHandler) HandleRepositoryError(operation string, err error) error {
	if err == nil {
		return nil
	}

	// Handle repository-specific errors
	switch {
	case errors.Is(err, mysqlErrors.ErrNotFound):
		return h.HandleNotFound(operation, fmt.Sprintf("%s not found", h.entityName))
	case errors.Is(err, mysqlErrors.ErrDuplicateKey):
		return h.HandleConflict(operation, fmt.Sprintf("%s already exists", h.entityName))
	case errors.Is(err, mysqlErrors.ErrForeignKey):
		return h.HandleConflict(operation, fmt.Sprintf("Cannot perform operation: related %s exists", h.entityName))
	case isValidationError(err):
		return h.HandleValidationError(operation, err)
	default:
		return h.HandleInternalError(operation, err)
	}
}

// HandleValidationError handles validation errors
func (h *ErrorHandler) HandleValidationError(operation string, err error) error {
	return appErr.NewDomainError(400, fmt.Sprintf("Validation failed for %s.%s: %s", h.entityName, operation, err.Error()))
}

// HandleNotFound handles not found errors
func (h *ErrorHandler) HandleNotFound(operation string, message string) error {
	if message == "" {
		message = fmt.Sprintf("%s not found", h.entityName)
	}
	return appErr.NewDomainError(404, message)
}

// HandleConflict handles conflict errors (duplicate, foreign key, etc.)
func (h *ErrorHandler) HandleConflict(operation string, message string) error {
	return appErr.NewDomainError(409, message)
}

// HandleInternalError handles internal server errors
func (h *ErrorHandler) HandleInternalError(operation string, err error) error {
	return appErr.NewDomainError(500, fmt.Sprintf("Internal error in %s.%s: %s", h.entityName, operation, err.Error()))
}

// HandleBusinessRuleViolation handles business rule violations
func (h *ErrorHandler) HandleBusinessRuleViolation(operation string, rule string, message string) error {
	return appErr.NewDomainError(422, fmt.Sprintf("Business rule violation in %s.%s (%s): %s", h.entityName, operation, rule, message))
}

// HandleUnauthorized handles unauthorized access
func (h *ErrorHandler) HandleUnauthorized(operation string, message string) error {
	if message == "" {
		message = fmt.Sprintf("Unauthorized access to %s.%s", h.entityName, operation)
	}
	return appErr.NewDomainError(401, message)
}

// HandleForbidden handles forbidden access
func (h *ErrorHandler) HandleForbidden(operation string, message string) error {
	if message == "" {
		message = fmt.Sprintf("Forbidden access to %s.%s", h.entityName, operation)
	}
	return appErr.NewDomainError(403, message)
}

// Helper functions
func isValidationError(err error) bool {
	errStr := strings.ToLower(err.Error())
	validationKeywords := []string{"validation", "invalid", "required", "missing", "format"}

	for _, keyword := range validationKeywords {
		if strings.Contains(errStr, keyword) {
			return true
		}
	}
	return false
}

// CreateCustomError creates a custom domain error
func (h *ErrorHandler) CreateCustomError(status int, operation string, message string) error {
	return appErr.NewDomainError(status, fmt.Sprintf("%s.%s: %s", h.entityName, operation, message))
}
