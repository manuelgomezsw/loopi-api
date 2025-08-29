package repository

import (
	"gorm.io/gorm"
	"loopi-api/internal/models"
)

type UserRepository interface {
	FindByEmail(email string) (*models.User, error)
}

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*models.User, error) {
	var user models.User

	err := r.db.Preload("UserRoles.Role").
		Preload("UserRoles.Franchise").
		Preload("StoreUsers.Store").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if err != nil {
		return nil, err
	}

	return &user, nil
}
