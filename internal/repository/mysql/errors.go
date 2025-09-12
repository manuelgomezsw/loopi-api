package mysql

import (
	"errors"
	"fmt"
	"strings"
)

// Common repository errors
var (
	ErrNotFound          = errors.New("record not found")
	ErrDuplicateKey      = errors.New("duplicate key violation")
	ErrForeignKey        = errors.New("foreign key constraint violation")
	ErrInvalidInput      = errors.New("invalid input")
	ErrConnectionFailed  = errors.New("database connection failed")
	ErrTransactionFailed = errors.New("transaction failed")
)

// RepositoryError wraps database errors with context
type RepositoryError struct {
	Operation string
	Table     string
	ID        interface{}
	Err       error
}

func (e *RepositoryError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("repository error: %s operation failed on table %s for ID %v: %v",
			e.Operation, e.Table, e.ID, e.Err)
	}
	return fmt.Sprintf("repository error: %s operation failed on table %s: %v",
		e.Operation, e.Table, e.Err)
}

func (e *RepositoryError) Unwrap() error {
	return e.Err
}

// NewRepositoryError creates a new repository error
func NewRepositoryError(operation, table string, id interface{}, err error) *RepositoryError {
	return &RepositoryError{
		Operation: operation,
		Table:     table,
		ID:        id,
		Err:       err,
	}
}

// ErrorHandler provides standardized error handling for repositories
type ErrorHandler struct {
	tableName string
}

// NewErrorHandler creates a new error handler for a specific table
func NewErrorHandler(tableName string) *ErrorHandler {
	return &ErrorHandler{tableName: tableName}
}

// HandleError processes and wraps database errors appropriately
func (h *ErrorHandler) HandleError(operation string, err error, id ...interface{}) error {
	if err == nil {
		return nil
	}

	// Handle specific database errors
	switch {
	case isDuplicateKeyError(err):
		return NewRepositoryError(operation, h.tableName, getID(id), ErrDuplicateKey)
	case isForeignKeyError(err):
		return NewRepositoryError(operation, h.tableName, getID(id), ErrForeignKey)
	case isConnectionError(err):
		return NewRepositoryError(operation, h.tableName, getID(id), ErrConnectionFailed)
	default:
		return NewRepositoryError(operation, h.tableName, getID(id), err)
	}
}

// HandleNotFound returns ErrNotFound for record not found scenarios
func (h *ErrorHandler) HandleNotFound(operation string, id ...interface{}) error {
	return NewRepositoryError(operation, h.tableName, getID(id), ErrNotFound)
}

// Validation helpers
func isDuplicateKeyError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "duplicate") ||
		strings.Contains(errStr, "unique constraint") ||
		strings.Contains(errStr, "1062")
}

func isForeignKeyError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "foreign key") ||
		strings.Contains(errStr, "constraint") ||
		strings.Contains(errStr, "1452")
}

func isConnectionError(err error) bool {
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "connection") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "refused")
}

func getID(ids []interface{}) interface{} {
	if len(ids) > 0 {
		return ids[0]
	}
	return nil
}

// LogError logs repository errors with context (can be extended with proper logging)
func (h *ErrorHandler) LogError(operation string, err error, id ...interface{}) {
	// This can be extended to use a proper logging framework
	fmt.Printf("Repository Error [%s] Operation: %s, Table: %s, ID: %v, Error: %v\n",
		operation, h.tableName, getID(id), err)
}
