package mysql

import (
	"gorm.io/gorm"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type storeRepo struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) repository.StoreRepository {
	return &storeRepo{db}
}

func (r *storeRepo) GetAll() ([]domain.Store, error) {
	var stores []domain.Store
	err := r.db.Find(&stores).Error
	return stores, err
}

func (r *storeRepo) GetByID(id int) (domain.Store, error) {
	var store domain.Store
	err := r.db.First(&store, id).Error
	return store, err
}

func (r *storeRepo) GetByFranchiseID(franchiseID int) ([]domain.Store, error) {
	var stores []domain.Store
	err := r.db.Where("franchise_id = ?", franchiseID).Find(&stores).Error
	return stores, err
}

func (r *storeRepo) Create(s *domain.Store) error {
	return r.db.Create(s).Error
}

func (r *storeRepo) Update(s *domain.Store) error {
	return r.db.Save(s).Error
}

func (r *storeRepo) Delete(id int) error {
	return r.db.Delete(&domain.Store{}, id).Error
}
