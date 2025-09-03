package usecase

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type StoreUseCase interface {
	GetAll() ([]domain.Store, error)
	GetByID(id int) (domain.Store, error)
	GetByFranchiseID(franchiseID int) ([]domain.Store, error)
	Create(s *domain.Store) error
	Update(s *domain.Store) error
	Delete(id int) error
}

type storeUseCase struct {
	repo repository.StoreRepository
}

func NewStoreUseCase(repo repository.StoreRepository) StoreUseCase {
	return &storeUseCase{repo}
}

func (u *storeUseCase) GetAll() ([]domain.Store, error) {
	return u.repo.GetAll()
}

func (u *storeUseCase) GetByID(id int) (domain.Store, error) {
	return u.repo.GetByID(id)
}

func (u *storeUseCase) GetByFranchiseID(franchiseID int) ([]domain.Store, error) {
	return u.repo.GetByFranchiseID(franchiseID)
}

func (u *storeUseCase) Create(s *domain.Store) error {
	return u.repo.Create(s)
}

func (u *storeUseCase) Update(s *domain.Store) error {
	return u.repo.Update(s)
}

func (u *storeUseCase) Delete(id int) error {
	return u.repo.Delete(id)
}
