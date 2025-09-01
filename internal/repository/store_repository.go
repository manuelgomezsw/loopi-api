package repository

import (
	"loopi-api/internal/domain"
)

type StoreRepository interface {
	GetAll() ([]domain.Store, error)
	GetByID(id int) (domain.Store, error)
	GetByFranchiseID(franchiseID int) ([]domain.Store, error)
	Create(s *domain.Store) error
	Update(s *domain.Store) error
	Delete(id int) error
}
