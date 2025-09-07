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
	GetByStore(storeID int) ([]domain.User, error)
	Create(user domain.User, roleID, franchiseID int) error
	Update(id int, fields map[string]interface{}) error
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
		return nil, appErr.NewDomainError(404, "employees not found")
	}

	return employees, nil
}

func (u *employeeUseCase) GetByStore(storeID int) ([]domain.User, error) {
	employees, err := u.userRepo.GetByStore(storeID)
	if err != nil {
		log.Printf("error: %v\n", err)
		return nil, appErr.NewDomainError(500, err.Error())
	}

	if len(employees) == 0 {
		return nil, appErr.NewDomainError(404, "employees not found")
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

func (u *employeeUseCase) Update(id int, fields map[string]interface{}) error {
	if len(fields) == 0 {
		return appErr.NewDomainError(400, "no fields to update")
	}

	// Validar campos permitidos
	allowed := map[string]bool{
		"first_name":      true,
		"last_name":       true,
		"phone":           true,
		"email":           true,
		"position":        true,
		"salary":          true,
		"document_type":   true,
		"document_number": true,
		"password_hash":   true,
	}

	clean := make(map[string]interface{})
	for k, v := range fields {
		if !allowed[k] {
			continue
		}

		if k == "password_hash" {
			raw, ok := v.(string)
			if !ok || raw == "" {
				return appErr.NewDomainError(400, "invalid password")
			}
			hashed, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.DefaultCost)
			if err != nil {
				return appErr.NewDomainError(500, "could not hash password")
			}
			clean[k] = string(hashed)
			continue
		}

		clean[k] = v
	}

	if len(clean) == 0 {
		return appErr.NewDomainError(400, "no valid fields to update")
	}

	return u.userRepo.Update(id, clean)
}

func (u *employeeUseCase) Delete(id int) error {
	if err := u.userRepo.Delete(id); err != nil {
		return appErr.NewDomainError(500, "could not delete employee")
	}

	return nil
}
