package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type absenceRepository struct {
	db *gorm.DB
}

func NewAbsenceRepository(db *gorm.DB) repository.AbsenceRepository {
	return &absenceRepository{db: db}
}

func (r *absenceRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error) {
	var absences []domain.Absence
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)
	err := r.db.Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, start, end).Find(&absences).Error
	return absences, err
}

func (r *absenceRepository) Create(absence *domain.Absence) error {
	return r.db.Create(absence).Error
}
