package repository

import "loopi-api/internal/domain"

type AbsenceRepository interface {
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error)
	Create(absence *domain.Absence) error
}
