package repository

import (
	"loopi-api/internal/domain"
)

type UserRepository interface {
	GetNameByID(userID int) (string, error)
	GetAll() ([]domain.User, error)
	GetByStore(storeID int) ([]domain.User, error)
	FindByEmail(email string) (*domain.User, error)
	FindByID(userID int) (*domain.User, error)
	Create(user domain.User, roleID, franchiseID int) error
	Update(id int, fields map[string]interface{}) error
	Delete(id int) error
}
