package usecase

import (
	"fmt"
	"loopi-api/config"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type AuthUseCase interface {
	// Standard authentication operations
	Login(email, password string) (string, error)
	SelectContext(userID int, franchiseID, storeID int) (string, error)

	// Business-specific operations
	ValidateLoginCredentials(email, password string) error
	ValidateUserAccess(userID, franchiseID int) error
	GetUserRoles(user *domain.User) []string
	GenerateToken(userID int, email string, roles []string, franchiseID, storeID int) (string, error)
}

type authUseCase struct {
	userRepo     repository.UserRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewAuthUseCase(userRepo repository.UserRepository) AuthUseCase {
	return &authUseCase{
		userRepo:     userRepo,
		errorHandler: base.NewErrorHandler("Auth"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Auth"),
	}
}

// ✅ Enhanced authentication operations with logging, validation, and error handling

// Login authenticates a user and returns a JWT token
func (uc *authUseCase) Login(email, password string) (string, error) {
	uc.logger.LogOperation("Login", "start", map[string]interface{}{
		"email": email,
	})

	// Validate credentials format
	if err := uc.ValidateLoginCredentials(email, password); err != nil {
		return "", err
	}

	// Find user by email
	user, err := uc.userRepo.FindByEmail(email)
	if err != nil {
		uc.logger.LogError("Login", err, map[string]interface{}{
			"email": email,
		})
		return "", uc.errorHandler.HandleRepositoryError("Login", err)
	}

	// Check if user exists
	if user == nil {
		err := fmt.Errorf("user not found with email: %s", email)
		uc.logger.LogError("Login", err, map[string]interface{}{
			"email": email,
		})
		return "", uc.errorHandler.HandleNotFound("Login", fmt.Sprintf("user not found with email: %s", email))
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		uc.logger.LogError("Login", err, map[string]interface{}{
			"email":   email,
			"user_id": user.ID,
		})
		return "", uc.errorHandler.HandleUnauthorized("Login", fmt.Sprintf("invalid credentials for user: %s", email))
	}

	// Generate JWT token without context (initial login)
	token, err := uc.GenerateToken(
		int(user.ID),
		user.Email,
		uc.GetUserRoles(user),
		0, // No franchise context yet
		0, // No store context yet
	)
	if err != nil {
		return "", err // Error already handled by GenerateToken
	}

	uc.logger.LogOperation("Login", "success", map[string]interface{}{
		"user_id": user.ID,
		"email":   user.Email,
		"roles":   uc.GetUserRoles(user),
	})

	return token, nil
}

// SelectContext validates user access to a franchise/store and generates contextual token
func (uc *authUseCase) SelectContext(userID int, franchiseID, storeID int) (string, error) {
	uc.logger.LogOperation("SelectContext", "start", map[string]interface{}{
		"user_id":      userID,
		"franchise_id": franchiseID,
		"store_id":     storeID,
	})

	// Validate IDs
	if err := uc.validator.ValidateID(userID); err != nil {
		uc.logger.LogError("SelectContext", err, map[string]interface{}{
			"user_id": userID,
		})
		return "", uc.errorHandler.HandleValidationError("SelectContext", err)
	}

	if err := uc.validator.ValidateID(franchiseID); err != nil {
		uc.logger.LogError("SelectContext", err, map[string]interface{}{
			"franchise_id": franchiseID,
		})
		return "", uc.errorHandler.HandleValidationError("SelectContext", err)
	}

	// Find user
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		uc.logger.LogError("SelectContext", err, map[string]interface{}{
			"user_id": userID,
		})
		return "", uc.errorHandler.HandleRepositoryError("SelectContext", err)
	}

	if user == nil {
		err := fmt.Errorf("user not found with ID: %d", userID)
		uc.logger.LogError("SelectContext", err, map[string]interface{}{
			"user_id": userID,
		})
		return "", uc.errorHandler.HandleNotFound("SelectContext", fmt.Sprintf("user not found with ID: %d", userID))
	}

	// Validate user access to franchise
	if err := uc.ValidateUserAccess(userID, franchiseID); err != nil {
		return "", err // Error already handled by ValidateUserAccess
	}

	// Generate contextual JWT token
	token, err := uc.GenerateToken(
		userID,
		user.Email,
		uc.GetUserRoles(user),
		franchiseID,
		storeID,
	)
	if err != nil {
		return "", err // Error already handled by GenerateToken
	}

	uc.logger.LogOperation("SelectContext", "success", map[string]interface{}{
		"user_id":      userID,
		"franchise_id": franchiseID,
		"store_id":     storeID,
		"roles":        uc.GetUserRoles(user),
	})

	return token, nil
}

// ✅ Business-specific operations with enhanced validation and logging

// ValidateLoginCredentials validates the format and requirements of login credentials
func (uc *authUseCase) ValidateLoginCredentials(email, password string) error {
	uc.logger.LogOperation("ValidateLoginCredentials", "start", map[string]interface{}{
		"email": email,
	})

	// Validate email format
	if err := uc.validator.ValidateString(email, "email", "required", "email"); err != nil {
		uc.logger.LogValidation("ValidateLoginCredentials", "email", "failed", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		})
		return uc.errorHandler.HandleValidationError("ValidateLoginCredentials", err)
	}

	// Validate password requirements
	if err := uc.validator.ValidateString(password, "password", "required", "min:6"); err != nil {
		uc.logger.LogValidation("ValidateLoginCredentials", "password", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateLoginCredentials", err)
	}

	// Business rule: Email cannot be empty and must be trimmed
	email = strings.TrimSpace(email)
	if email == "" {
		err := fmt.Errorf("email cannot be empty")
		uc.logger.LogValidation("ValidateLoginCredentials", "empty_email", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateLoginCredentials", err)
	}

	// Business rule: Password minimum length
	if len(strings.TrimSpace(password)) < 6 {
		err := fmt.Errorf("password must be at least 6 characters long")
		uc.logger.LogValidation("ValidateLoginCredentials", "password_length", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateLoginCredentials", err)
	}

	uc.logger.LogValidation("ValidateLoginCredentials", "all_fields", "passed", map[string]interface{}{
		"email": email,
	})

	return nil
}

// ValidateUserAccess validates that a user has access to a specific franchise
func (uc *authUseCase) ValidateUserAccess(userID, franchiseID int) error {
	uc.logger.LogOperation("ValidateUserAccess", "start", map[string]interface{}{
		"user_id":      userID,
		"franchise_id": franchiseID,
	})

	// Get user to check franchise access
	user, err := uc.userRepo.FindByID(userID)
	if err != nil {
		uc.logger.LogError("ValidateUserAccess", err, map[string]interface{}{
			"user_id": userID,
		})
		return uc.errorHandler.HandleRepositoryError("ValidateUserAccess", err)
	}

	if user == nil {
		err := fmt.Errorf("user not found with ID: %d", userID)
		uc.logger.LogError("ValidateUserAccess", err, map[string]interface{}{
			"user_id": userID,
		})
		return uc.errorHandler.HandleNotFound("ValidateUserAccess", fmt.Sprintf("user not found with ID: %d", userID))
	}

	// Business rule: User must belong to the requested franchise
	hasAccess := false
	userFranchises := make([]int, 0)

	for _, userRole := range user.UserRoles {
		userFranchises = append(userFranchises, userRole.FranchiseID)
		if userRole.FranchiseID == franchiseID {
			hasAccess = true
			break
		}
	}

	if !hasAccess {
		err := fmt.Errorf("user %d does not have access to franchise %d. Available franchises: %v",
			userID, franchiseID, userFranchises)
		uc.logger.LogValidation("ValidateUserAccess", "franchise_access", "failed", map[string]interface{}{
			"error":                err.Error(),
			"user_id":              userID,
			"requested_franchise":  franchiseID,
			"available_franchises": userFranchises,
		})
		return uc.errorHandler.HandleForbidden("ValidateUserAccess",
			fmt.Sprintf("user %d does not have access to franchise %d", userID, franchiseID))
	}

	uc.logger.LogValidation("ValidateUserAccess", "franchise_access", "passed", map[string]interface{}{
		"user_id":      userID,
		"franchise_id": franchiseID,
	})

	return nil
}

// GetUserRoles extracts roles from user domain object
func (uc *authUseCase) GetUserRoles(user *domain.User) []string {
	uc.logger.LogOperation("GetUserRoles", "start", map[string]interface{}{
		"user_id": user.ID,
	})

	roles := make([]string, 0, len(user.UserRoles))
	roleSet := make(map[string]bool) // To avoid duplicates

	for _, userRole := range user.UserRoles {
		roleName := userRole.Role.Name
		if !roleSet[roleName] {
			roles = append(roles, roleName)
			roleSet[roleName] = true
		}
	}

	uc.logger.LogOperation("GetUserRoles", "success", map[string]interface{}{
		"user_id":    user.ID,
		"roles":      roles,
		"role_count": len(roles),
	})

	return roles
}

// GenerateToken creates a JWT token with the provided claims
func (uc *authUseCase) GenerateToken(userID int, email string, roles []string, franchiseID, storeID int) (string, error) {
	uc.logger.LogOperation("GenerateToken", "start", map[string]interface{}{
		"user_id":      userID,
		"email":        email,
		"roles":        roles,
		"franchise_id": franchiseID,
		"store_id":     storeID,
	})

	// Business rule: UserID must be positive
	if userID <= 0 {
		err := fmt.Errorf("invalid user ID: %d. Must be positive", userID)
		uc.logger.LogValidation("GenerateToken", "user_id", "failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		})
		return "", uc.errorHandler.HandleValidationError("GenerateToken", err)
	}

	// Business rule: Email must be valid
	if err := uc.validator.ValidateString(email, "email", "required", "email"); err != nil {
		uc.logger.LogValidation("GenerateToken", "email", "failed", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		})
		return "", uc.errorHandler.HandleValidationError("GenerateToken", err)
	}

	// Generate JWT token using config
	token, err := config.GenerateJWT(userID, email, roles, franchiseID, storeID)
	if err != nil {
		uc.logger.LogError("GenerateToken", err, map[string]interface{}{
			"user_id":      userID,
			"email":        email,
			"franchise_id": franchiseID,
			"store_id":     storeID,
		})
		tokenErr := fmt.Errorf("failed to generate JWT token for user %d: %v", userID, err)
		return "", uc.errorHandler.HandleInternalError("GenerateToken", tokenErr)
	}

	uc.logger.LogOperation("GenerateToken", "success", map[string]interface{}{
		"user_id":      userID,
		"email":        email,
		"roles":        roles,
		"franchise_id": franchiseID,
		"store_id":     storeID,
	})

	return token, nil
}
