package employee

import (
	"fmt"

	"github.com/manuelgomezsw/loopi-api/internal/domain/entity"
)

const defaultPasswordFallback = "password123"

// DefaultPasswordForReset returns the default password when resetting an employee's password.
// Rule: document_number + birth_year; if no birth date, document_number; if no document, fallback constant.
// Used by AdminService.ResetEmployeePassword; the service hashes the result before persisting.
func DefaultPasswordForReset(e *entity.Employee) string {
	if e == nil {
		return defaultPasswordFallback
	}
	if e.DocumentNumber != nil && e.BirthDate != nil {
		return *e.DocumentNumber + fmt.Sprintf("%d", e.BirthDate.Year())
	}
	if e.DocumentNumber != nil {
		return *e.DocumentNumber
	}
	return defaultPasswordFallback
}
