package repository

import (
	"loopi-api/internal/domain"
)

type ShiftRepository interface {
	Create(cfg domain.Shift) error
	ListAll() ([]domain.Shift, error)
	ListByStore(storeID int) ([]domain.Shift, error)
	GetByID(id int) (*domain.Shift, error)
}
