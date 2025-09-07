package usecase

import (
	"golang.org/x/crypto/bcrypt"
	"log"
	appErr "loopi-api/internal/common/errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type EmployeeUseCase interface {
	GetAll() ([]domain.User, error)
	FindByID(id int) (*domain.User, error)
	Create(user domain.User, roleID, franchiseID int) error
	Update(emp *domain.User) error
	Delete(id int) error
}

type employeeUseCase struct {
	userRepo repository.UserRepository
}

func NewEmployeeUseCase(userRepo repository.UserRepository) EmployeeUseCase {
	return &employeeUseCase{userRepo: userRepo}
}

func (u *employeeUseCase) GetAll() ([]domain.User, error) {
	employees, err := u.userRepo.GetAll()
	if err != nil {
		log.Printf("error: %v\n", err)
		return nil, appErr.NewDomainError(500, err.Error())
	}

	if len(employees) == 0 {
		return nil, appErr.NewDomainError(404, "employee not found")
	}

	return employees, nil
}

func (u *employeeUseCase) FindByID(id int) (*domain.User, error) {
	user, err := u.userRepo.FindByID(id)

	if err != nil {
		return nil, appErr.NewDomainError(500, err.Error())
	}
	if user == nil {
		return nil, appErr.NewDomainError(404, "employee not found")
	}

	return user, nil
}

func (u *employeeUseCase) Create(user domain.User, roleID, franchiseID int) error {
	user.IsActive = true

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashed)

	return u.userRepo.Create(user, roleID, franchiseID)
}

func (u *employeeUseCase) Update(emp *domain.User) error {
	return u.userRepo.Update(emp)
}

func (u *employeeUseCase) Delete(id int) error {
	if err := u.userRepo.Delete(id); err != nil {
		return appErr.NewDomainError(500, "could not delete employee")
	}

	return nil
}
