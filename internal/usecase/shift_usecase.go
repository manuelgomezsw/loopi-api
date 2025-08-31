package usecase

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type ShiftUseCase interface {
	CreateShift(shift domain.Shift) error
}

type shiftUseCase struct {
	repo repository.ShiftRepository
}

func NewShiftUseCase(repo repository.ShiftRepository) ShiftUseCase {
	return &shiftUseCase{repo: repo}
}

func (u *shiftUseCase) CreateShift(shift domain.Shift) error {
	if shift.Name == "" || shift.StartTime == "" || shift.EndTime == "" || shift.StoreID == 0 {
		return errors.New("missing required shift fields")
	}
	return u.repo.Create(shift)
}
