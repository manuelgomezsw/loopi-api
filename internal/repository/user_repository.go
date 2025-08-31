package repository

import (
	"loopi-api/internal/domain"
)

type UserRepository interface {
	FindByEmail(email string) (*domain.User, error)
	FindByID(userID int) (*domain.User, error)
	Create(user domain.User, roleID, franchiseID int) error
	GetNameByID(userID int) (string, error)
}
