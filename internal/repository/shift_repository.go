package repository

import (
	"gorm.io/gorm"
	"loopi-api/internal/domain"
)

type ShiftRepository interface {
	Create(shift domain.Shift) error
}

type shiftRepository struct {
	db *gorm.DB
}

func NewShiftRepository(db *gorm.DB) ShiftRepository {
	return &shiftRepository{db: db}
}

func (r *shiftRepository) Create(shift domain.Shift) error {
	return r.db.Create(&shift).Error
}
