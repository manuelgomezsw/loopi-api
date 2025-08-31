package usecase

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"time"
)

type AbsenceUseCase interface {
	Create(absence domain.Absence) error
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error)
}

type absenceUseCase struct {
	repo repository.AbsenceRepository
}

func NewAbsenceUseCase(repo repository.AbsenceRepository) AbsenceUseCase {
	return &absenceUseCase{repo}
}

func (u *absenceUseCase) Create(abs domain.Absence) error {
	if abs.EmployeeID == 0 || abs.Date.IsZero() || abs.Hours <= 0 {
		return errors.New("missing required fields")
	}
	abs.CreatedAt = time.Now()
	abs.UpdatedAt = time.Now()
	return u.repo.Create(&abs)
}

func (u *absenceUseCase) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error) {
	return u.repo.GetByEmployeeAndMonth(employeeID, year, month)
}
