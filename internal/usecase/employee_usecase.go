package usecase

import (
	"golang.org/x/crypto/bcrypt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type EmployeeUseCase interface {
	CreateEmployee(user domain.User, roleID, franchiseID int) error
}

type employeeUseCase struct {
	userRepo repository.UserRepository
}

func NewEmployeeUseCase(userRepo repository.UserRepository) EmployeeUseCase {
	return &employeeUseCase{userRepo: userRepo}
}

func (u *employeeUseCase) CreateEmployee(user domain.User, roleID, franchiseID int) error {
	user.IsActive = true

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashed)

	return u.userRepo.Create(user, roleID, franchiseID)
}
