package mysql

import (
	"errors"
	"fmt"
	"gorm.io/gorm"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User

	err := r.db.Preload("UserRoles.Role").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

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

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *userRepository) GetAll() ([]domain.User, error) {
	var users []domain.User

	err := r.db.Find(&users).Error
	if err != nil {
		return nil, err
	}
	if len(users) == 0 {
		return []domain.User{}, nil
	}

	return users, nil
}

func (r *userRepository) GetNameByID(userID int) (string, error) {
	var user struct {
		FirstName string
		LastName  string
	}
	err := r.db.Table("users").Select("first_name, last_name").Where("id = ?", userID).Scan(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil
	}

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%s %s", user.FirstName, user.LastName), nil
}

func (r *userRepository) GetByStore(storeID int) ([]domain.User, error) {
	var users []domain.User

	err := r.db.
		Joins("JOIN store_users ON store_users.user_id = users.id").
		Where("store_users.store_id = ?", storeID).
		Find(&users).Error

	if err != nil {
		return nil, err
	}

	if len(users) == 0 {
		return []domain.User{}, nil
	}

	return users, err
}

func (r *userRepository) Create(user domain.User, roleID, franchiseID int) error {
	return r.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(&user).Error; err != nil {
			return err
		}
		userRole := domain.UserRole{
			UserID:      user.UserID,
			RoleID:      roleID,
			FranchiseID: franchiseID,
		}
		return tx.Create(&userRole).Error
	})
}

func (r *userRepository) Update(id int, fields map[string]interface{}) error {
	return r.db.Model(&domain.User{}).Where("id = ?", id).Updates(fields).Error
}

func (r *userRepository) Delete(id int) error {
	return r.db.Delete(&domain.User{}, id).Error
}
