package repository

import (
	"gorm.io/gorm"
	"loopi-api/internal/domain"
)

type UserRepository interface {
	FindByEmail(email string) (*domain.User, error)
	FindByID(userID int) (*domain.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User

	err := r.db.Preload("UserRoles.Role").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) FindByID(userID int) (*domain.User, error) {
	var user domain.User
	err := r.db.
		Preload("UserRoles.Role").
		Preload("UserRoles.Franchise").
		First(&user, userID).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}
