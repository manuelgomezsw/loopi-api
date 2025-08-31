package usecase

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"time"
)

type NoveltyUseCase interface {
	Create(n domain.Novelty) error
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error)
}

type noveltyUseCase struct {
	repo repository.NoveltyRepository
}

func NewNoveltyUseCase(repo repository.NoveltyRepository) NoveltyUseCase {
	return &noveltyUseCase{repo}
}

func (u *noveltyUseCase) Create(n domain.Novelty) error {
	if n.EmployeeID == 0 || n.Date.IsZero() || n.Hours <= 0 || (n.Type != "positive" && n.Type != "negative") {
		return errors.New("missing or invalid fields")
	}
	n.CreatedAt = time.Now()
	n.UpdatedAt = time.Now()
	return u.repo.Create(&n)
}

func (u *noveltyUseCase) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error) {
	return u.repo.GetByEmployeeAndMonth(employeeID, year, month)
}
