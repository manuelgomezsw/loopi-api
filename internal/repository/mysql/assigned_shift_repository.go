package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type assignedShiftRepository struct {
	db *gorm.DB
}

func NewAssignedShiftRepository(db *gorm.DB) repository.AssignedShiftRepository {
	return &assignedShiftRepository{db: db}
}

func (r *assignedShiftRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.AssignedShift, error) {
	var shifts []domain.AssignedShift
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)
	err := r.db.Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, start, end).Find(&shifts).Error
	return shifts, err
}
