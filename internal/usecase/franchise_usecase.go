package usecase

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type FranchiseUseCase interface {
	Create(franchise domain.Franchise) error
	GetAll() ([]domain.Franchise, error)
	GetById(id int) (domain.Franchise, error)
}

type franchiseUseCase struct {
	repo repository.FranchiseRepository
}

func NewFranchiseUseCase(repo repository.FranchiseRepository) FranchiseUseCase {
	return &franchiseUseCase{repo: repo}
}

func (f *franchiseUseCase) Create(franchise domain.Franchise) error {
	return f.repo.Create(&franchise)
}

func (f *franchiseUseCase) GetAll() ([]domain.Franchise, error) {
	return f.repo.GetAll()
}

func (f *franchiseUseCase) GetById(id int) (domain.Franchise, error) {
	return f.repo.GetById(id)
}
