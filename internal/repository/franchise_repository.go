package repository

import "loopi-api/internal/domain"

type FranchiseRepository interface {
  GetAll() ([]domain.Franchise, error)
  GetById(id int) (domain.Franchise, error)
  Create(franchise *domain.Franchise) error
}
