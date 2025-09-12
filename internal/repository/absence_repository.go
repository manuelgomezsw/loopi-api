package repository

import (
	"loopi-api/internal/domain"
	"time"
)

type AbsenceRepository interface {
	// Basic operations
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error)
	Create(absence *domain.Absence) error

	// Enhanced business operations
	GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Absence, error)
	GetTotalHoursByEmployee(employeeID, year, month int) (float64, error)
}
