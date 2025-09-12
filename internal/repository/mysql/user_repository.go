package mysql

import (
	"errors"
	"fmt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"

	"gorm.io/gorm"
)

// userRepository implements repository.UserRepository with improved maintainability
type userRepository struct {
	*BaseRepository[domain.User]
	errorHandler *ErrorHandler
}

// NewUserRepository creates a new user repository with enhanced features
func NewUserRepository(db *gorm.DB) repository.UserRepository {
	return &userRepository{
		BaseRepository: NewBaseRepository[domain.User](db, "users"),
		errorHandler:   NewErrorHandler("users"),
	}
}

// FindByEmail retrieves a user by email with roles preloaded
func (r *userRepository) FindByEmail(email string) (*domain.User, error) {
	var user domain.User

	err := r.GetDB().
		Preload("UserRoles.Role").
		Where("email = ? AND is_active = ?", email, true).
		First(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Return nil for not found (business logic)
	}

	if err != nil {
		return nil, r.errorHandler.HandleError("FindByEmail", err, email)
	}

	return &user, nil
}

// FindByID retrieves a user by ID with all related data preloaded
func (r *userRepository) FindByID(userID int) (*domain.User, error) {
	var user domain.User
	err := r.GetDB().
		Preload("UserRoles.Role").
		Preload("UserRoles.Franchise").
		First(&user, userID).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // Return nil for not found (business logic)
	}

	if err != nil {
		return nil, r.errorHandler.HandleError("FindByID", err, userID)
	}

	return &user, nil
}

// GetAll retrieves all users with proper error handling
func (r *userRepository) GetAll() ([]domain.User, error) {
	users, err := r.BaseRepository.GetAll()
	if err != nil {
		return nil, r.errorHandler.HandleError("GetAll", err)
	}

	// Return empty slice instead of nil for consistency
	if len(users) == 0 {
		return []domain.User{}, nil
	}

	return users, nil
}

// GetNameByID retrieves formatted full name for a user
func (r *userRepository) GetNameByID(userID int) (string, error) {
	var user struct {
		FirstName string
		LastName  string
	}

	err := r.GetDB().
		Table("users").
		Select("first_name, last_name").
		Where("id = ?", userID).
		Scan(&user).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", nil // Return empty string for not found
	}

	if err != nil {
		return "", r.errorHandler.HandleError("GetNameByID", err, userID)
	}

	return fmt.Sprintf("%s %s", user.FirstName, user.LastName), nil
}

// GetByStore retrieves users associated with a specific store
func (r *userRepository) GetByStore(storeID int) ([]domain.User, error) {
	users, err := FindWithJoin[domain.User](
		r.GetDB(),
		"store_users",
		"store_users.user_id = users.id",
		map[string]interface{}{
			"store_users.store_id": storeID,
		},
	)

	if err != nil {
		return nil, r.errorHandler.HandleError("GetByStore", err, storeID)
	}

	// Return empty slice instead of nil for consistency
	if len(users) == 0 {
		return []domain.User{}, nil
	}

	return users, nil
}

// Create creates a new user with role assignment in a transaction
func (r *userRepository) Create(user domain.User, roleID, franchiseID int) error {
	// Business validation before creation
	if err := r.validateUser(&user); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	// Use transaction to ensure atomicity
	return r.BaseRepository.Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		// Create user role relationship
		userRole := domain.UserRole{
			UserID:      int(user.ID),
			RoleID:      roleID,
			FranchiseID: franchiseID,
		}

		return tx.Create(&userRole).Error
	})
}

// Update updates user fields with validation
func (r *userRepository) Update(id int, fields map[string]interface{}) error {
	// Check if user exists
	exists, err := r.BaseRepository.Exists(id)
	if err != nil {
		return r.errorHandler.HandleError("Update", err, id)
	}
	if !exists {
		return r.errorHandler.HandleNotFound("Update", id)
	}

	// Validate update fields
	if err := r.validateUpdateFields(fields); err != nil {
		return r.errorHandler.HandleError("Update", err, id)
	}

	err = r.GetDB().Model(&domain.User{}).Where("id = ?", id).Updates(fields).Error
	if err != nil {
		return r.errorHandler.HandleError("Update", err, id)
	}

	return nil
}

// Delete removes a user by ID with proper validation
func (r *userRepository) Delete(id int) error {
	// Check if user exists
	exists, err := r.BaseRepository.Exists(id)
	if err != nil {
		return r.errorHandler.HandleError("Delete", err, id)
	}
	if !exists {
		return r.errorHandler.HandleNotFound("Delete", id)
	}

	if err := r.BaseRepository.Delete(id); err != nil {
		return r.errorHandler.HandleError("Delete", err, id)
	}
	return nil
}

// validateUser performs business validation
func (r *userRepository) validateUser(user *domain.User) error {
	if user.Email == "" {
		return ErrInvalidInput
	}
	if user.FirstName == "" {
		return ErrInvalidInput
	}
	if user.LastName == "" {
		return ErrInvalidInput
	}
	return nil
}

// validateUpdateFields validates fields for update operations
func (r *userRepository) validateUpdateFields(fields map[string]interface{}) error {
	// Check for required fields if being updated
	if email, exists := fields["email"]; exists {
		if email == "" {
			return ErrInvalidInput
		}
	}
	if firstName, exists := fields["first_name"]; exists {
		if firstName == "" {
			return ErrInvalidInput
		}
	}
	if lastName, exists := fields["last_name"]; exists {
		if lastName == "" {
			return ErrInvalidInput
		}
	}
	return nil
}

// GetActiveUsers retrieves only active users
func (r *userRepository) GetActiveUsers() ([]domain.User, error) {
	var users []domain.User
	err := NewQueryBuilder(r.GetDB()).
		WhereActive().
		WhereNotDeleted().
		OrderBy("first_name").
		OrderBy("last_name").
		GetDB().
		Find(&users).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetActiveUsers", err)
	}
	return users, nil
}

// GetUsersByRole retrieves users with a specific role
func (r *userRepository) GetUsersByRole(roleID int) ([]domain.User, error) {
	var users []domain.User
	err := r.GetDB().
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.role_id = ? AND users.is_active = ?", roleID, true).
		Find(&users).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetUsersByRole", err, roleID)
	}
	return users, nil
}

// GetUsersByFranchise retrieves users associated with a specific franchise
func (r *userRepository) GetUsersByFranchise(franchiseID int) ([]domain.User, error) {
	var users []domain.User
	err := r.GetDB().
		Joins("JOIN user_roles ON user_roles.user_id = users.id").
		Where("user_roles.franchise_id = ? AND users.is_active = ?", franchiseID, true).
		Preload("UserRoles.Role").
		Find(&users).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetUsersByFranchise", err, franchiseID)
	}
	return users, nil
}

// GetUsersWithRoles retrieves users with their roles preloaded
func (r *userRepository) GetUsersWithRoles() ([]domain.User, error) {
	var users []domain.User
	err := r.GetDB().
		Preload("UserRoles.Role").
		Preload("UserRoles.Franchise").
		Where("is_active = ?", true).
		Order("first_name, last_name").
		Find(&users).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetUsersWithRoles", err)
	}
	return users, nil
}

// AssignUserToStore assigns a user to a store
func (r *userRepository) AssignUserToStore(userID, storeID int) error {
	// Check if user exists
	exists, err := r.BaseRepository.Exists(userID)
	if err != nil {
		return r.errorHandler.HandleError("AssignUserToStore", err, userID)
	}
	if !exists {
		return r.errorHandler.HandleNotFound("AssignUserToStore", userID)
	}

	// Create store-user relationship
	storeUser := struct {
		StoreID int `gorm:"column:store_id"`
		UserID  int `gorm:"column:user_id"`
	}{
		StoreID: storeID,
		UserID:  userID,
	}

	err = r.GetDB().Table("store_users").Create(&storeUser).Error
	if err != nil {
		return r.errorHandler.HandleError("AssignUserToStore", err, userID)
	}

	return nil
}

// RemoveUserFromStore removes a user from a store
func (r *userRepository) RemoveUserFromStore(userID, storeID int) error {
	err := r.GetDB().
		Table("store_users").
		Where("user_id = ? AND store_id = ?", userID, storeID).
		Delete(nil).Error

	if err != nil {
		return r.errorHandler.HandleError("RemoveUserFromStore", err, userID)
	}

	return nil
}
