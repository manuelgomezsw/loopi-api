package repository

import "loopi-api/internal/domain"

type AssignedShiftRepository interface {
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.AssignedShift, error)
}
