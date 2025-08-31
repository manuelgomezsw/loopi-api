package repository

import (
	"gorm.io/gorm"
	"loopi-api/internal/domain"
)

type ShiftRepository interface {
	Create(cfg domain.Shift) error
	ListAll() ([]domain.Shift, error)
	ListByStore(storeID int) ([]domain.Shift, error)
	GetByID(id int) (*domain.Shift, error)
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) Create(cfg domain.Shift) error {
	return r.db.Create(&cfg).Error
}

func (r *shiftRepository) ListAll() ([]domain.Shift, error) {
	var shifts []domain.Shift
	err := r.db.Order("name").Find(&shifts).Error
	return shifts, err
}

func (r *shiftRepository) ListByStore(storeID int) ([]domain.Shift, error) {
	var shifts []domain.Shift
	err := r.db.
		Where("store_id = ?", storeID).
		Order("name").
		Find(&shifts).Error
	return shifts, err
}

func (r *shiftRepository) GetByID(id int) (*domain.Shift, error) {
	var shift domain.Shift
	err := r.db.First(&shift, id).Error
	if err != nil {
		return nil, err
	}
	return &shift, nil
}
