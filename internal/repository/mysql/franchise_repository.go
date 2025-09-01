package mysql

import (
	"gorm.io/gorm"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type franchiseRepository struct {
	db *gorm.DB
}

func NewFranchiseRepository(db *gorm.DB) repository.FranchiseRepository {
	return &franchiseRepository{db: db}
}

func (r *franchiseRepository) GetAll() ([]domain.Franchise, error) {
	var franchises []domain.Franchise
	err := r.db.Find(&franchises).Error
	return franchises, err
}

func (r *franchiseRepository) GetById(id int) (domain.Franchise, error) {
	var franchise domain.Franchise
	err := r.db.First(id).Find(&franchise).Error
	return franchise, err
}

func (r *franchiseRepository) Create(franchise *domain.Franchise) error {
	return r.db.Create(franchise).Error
}
