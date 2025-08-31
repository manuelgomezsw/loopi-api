package mysql

import (
	"fmt"
	"gorm.io/gorm"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type noveltyRepository struct {
	db *gorm.DB
}

func NewNoveltyRepository(db *gorm.DB) repository.NoveltyRepository {
	return &noveltyRepository{db: db}
}

func (r *noveltyRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error) {
	var novelties []domain.Novelty
	start := fmt.Sprintf("%04d-%02d-01", year, month)
	end := fmt.Sprintf("%04d-%02d-31", year, month)
	err := r.db.Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, start, end).Find(&novelties).Error
	return novelties, err
}

func (r *noveltyRepository) Create(n *domain.Novelty) error {
	return r.db.Create(n).Error
}
