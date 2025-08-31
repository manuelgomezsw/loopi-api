package usecase

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type ShiftUseCase interface {
	Create(cfg domain.Shift) error
	GetAll() ([]domain.Shift, error)
	GetByStore(storeID int) ([]domain.Shift, error)
	GetByID(id int) (*domain.Shift, error)
}

type shiftUseCase struct {
	repo repository.ShiftRepository
}

func NewShiftUseCase(repo repository.ShiftRepository) ShiftUseCase {
	return &shiftUseCase{repo: repo}
}

func (u *shiftUseCase) Create(cfg domain.Shift) error {
	if cfg.Name == "" || cfg.Period == "" || cfg.StartTime == "" || cfg.EndTime == "" {
		return errors.New("missing required fields")
	}
	return u.repo.Create(cfg)
}

func (u *shiftUseCase) GetAll() ([]domain.Shift, error) {
	return u.repo.ListAll()
}

func (u *shiftUseCase) GetByStore(storeID int) ([]domain.Shift, error) {
	return u.repo.ListByStore(storeID)
}

func (u *shiftUseCase) GetByID(id int) (*domain.Shift, error) {
	return u.repo.GetByID(id)
}
