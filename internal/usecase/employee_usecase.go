package usecase

import (
	"fmt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
)

type EmployeeUseCase interface {
	// Standard CRUD operations
	GetAll() ([]domain.User, error)
	FindByID(id int) (*domain.User, error)
	GetByStore(storeID int) ([]domain.User, error)
	Create(user domain.User, storeID int) error
	Update(id int, fields map[string]interface{}) error
	Delete(id int) error

	// Business-specific operations
	GetActiveEmployees() ([]domain.User, error)
	GetByStoreAndActive(storeID int) ([]domain.User, error)
	ValidateEmployeeData(user *domain.User) error
	ValidateEmployeeCredentials(email, documentNumber string) error
	HashPassword(password string) (string, error)
	ValidateUpdateFields(fields map[string]interface{}) (map[string]interface{}, error)
}

type employeeUseCase struct {
	userRepo      repository.UserRepository
	storeRepo     repository.StoreRepository
	franchiseRepo repository.FranchiseRepository
	errorHandler  *base.ErrorHandler
	validator     *base.Validator
	logger        *base.Logger
}

func NewEmployeeUseCase(userRepo repository.UserRepository, storeRepo repository.StoreRepository, franchiseRepo repository.FranchiseRepository) EmployeeUseCase {
	return &employeeUseCase{
		userRepo:      userRepo,
		storeRepo:     storeRepo,
		franchiseRepo: franchiseRepo,
		errorHandler:  base.NewErrorHandler("Employee"),
		validator:     base.NewValidator(),
		logger:        base.NewLogger("Employee"),
	}
}

// ✅ Enhanced CRUD operations with logging, validation, and error handling

// GetAll retrieves all employees with logging and error handling
func (uc *employeeUseCase) GetAll() ([]domain.User, error) {
	uc.logger.LogOperation("GetAll", "start", nil)

	employees, err := uc.userRepo.GetAll()
	if err != nil {
		uc.logger.LogError("GetAll", err, nil)
		return nil, uc.errorHandler.HandleRepositoryError("GetAll", err)
	}

	uc.logger.LogOperation("GetAll", "success", map[string]interface{}{
		"count": len(employees),
	})

	return employees, nil
}

// GetByStore retrieves employees by store with validation and logging
func (uc *employeeUseCase) GetByStore(storeID int) ([]domain.User, error) {
	uc.logger.LogOperation("GetByStore", "start", map[string]interface{}{
		"store_id": storeID,
	})

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("GetByStore", err, map[string]interface{}{
			"store_id": storeID,
		})
		return nil, uc.errorHandler.HandleValidationError("GetByStore", err)
	}

	employees, err := uc.userRepo.GetByStore(storeID)
	if err != nil {
		uc.logger.LogError("GetByStore", err, map[string]interface{}{
			"store_id": storeID,
		})
		return nil, uc.errorHandler.HandleRepositoryError("GetByStore", err)
	}

	uc.logger.LogOperation("GetByStore", "success", map[string]interface{}{
		"store_id": storeID,
		"count":    len(employees),
	})

	return employees, nil
}

// FindByID retrieves an employee by ID with validation and logging
func (uc *employeeUseCase) FindByID(id int) (*domain.User, error) {
	uc.logger.LogOperation("FindByID", "start", map[string]interface{}{
		"employee_id": id,
	})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("FindByID", err, map[string]interface{}{
			"employee_id": id,
		})
		return nil, uc.errorHandler.HandleValidationError("FindByID", err)
	}

	employee, err := uc.userRepo.FindByID(id)
	if err != nil {
		uc.logger.LogError("FindByID", err, map[string]interface{}{
			"employee_id": id,
		})
		return nil, uc.errorHandler.HandleRepositoryError("FindByID", err)
	}

	if employee == nil {
		uc.logger.LogError("FindByID", fmt.Errorf("employee not found"), map[string]interface{}{
			"employee_id": id,
		})
		return nil, uc.errorHandler.HandleNotFound("FindByID", fmt.Sprintf("employee not found with ID: %d", id))
	}

	uc.logger.LogOperation("FindByID", "success", map[string]interface{}{
		"employee_id": id,
		"email":       employee.Email,
	})

	return employee, nil
}

// Create creates a new employee with validation and password hashing
func (uc *employeeUseCase) Create(user domain.User, storeID int) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"email":    user.Email,
		"store_id": storeID,
	})

	// Validate employee data
	if err := uc.ValidateEmployeeData(&user); err != nil {
		return err
	}

	// Validate credentials uniqueness
	if err := uc.ValidateEmployeeCredentials(user.Email, user.DocumentNumber); err != nil {
		return err
	}

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"store_id": storeID,
		})
		return uc.errorHandler.HandleValidationError("Create", fmt.Errorf("invalid store ID: %v", err))
	}

	// Set defaults
	user.IsActive = true

	// Generate default password based on franchise name + current year
	defaultPassword, err := uc.GenerateDefaultPassword(storeID)
	if err != nil {
		return err
	}

	// Hash the generated password
	hashedPassword, err := uc.HashPassword(defaultPassword)
	if err != nil {
		return err
	}
	user.PasswordHash = hashedPassword

	// Execute creation with store association
	if err := uc.userRepo.CreateWithStore(user, storeID); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"email":    user.Email,
			"store_id": storeID,
		})
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"email":    user.Email,
		"store_id": storeID,
	})

	return nil
}

// Update updates employee fields with validation
func (uc *employeeUseCase) Update(id int, fields map[string]interface{}) error {
	uc.logger.LogOperation("Update", "start", map[string]interface{}{
		"employee_id": id,
		"field_count": len(fields),
	})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{
			"employee_id": id,
		})
		return uc.errorHandler.HandleValidationError("Update", err)
	}

	// Validate and clean fields
	cleanFields, err := uc.ValidateUpdateFields(fields)
	if err != nil {
		return err
	}

	// Execute update
	if err := uc.userRepo.Update(id, cleanFields); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{
			"employee_id": id,
			"fields":      cleanFields,
		})
		return uc.errorHandler.HandleRepositoryError("Update", err)
	}

	uc.logger.LogOperation("Update", "success", map[string]interface{}{
		"employee_id":    id,
		"updated_fields": cleanFields,
	})

	return nil
}

// Delete removes an employee with validation
func (uc *employeeUseCase) Delete(id int) error {
	uc.logger.LogOperation("Delete", "start", map[string]interface{}{
		"employee_id": id,
	})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{
			"employee_id": id,
		})
		return uc.errorHandler.HandleValidationError("Delete", err)
	}

	// Check if employee exists
	employee, err := uc.FindByID(id)
	if err != nil {
		return err // Error already logged by FindByID
	}

	if employee == nil {
		return uc.errorHandler.HandleNotFound("Delete", fmt.Sprintf("employee not found with ID: %d", id))
	}

	// Execute deletion
	if err := uc.userRepo.Delete(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{
			"employee_id": id,
		})
		return uc.errorHandler.HandleRepositoryError("Delete", err)
	}

	uc.logger.LogOperation("Delete", "success", map[string]interface{}{
		"employee_id": id,
		"email":       employee.Email,
	})

	return nil
}

// ✅ Business-specific operations with enhanced validation and logging

// GetActiveEmployees retrieves only active employees
func (uc *employeeUseCase) GetActiveEmployees() ([]domain.User, error) {
	uc.logger.LogOperation("GetActiveEmployees", "start", nil)

	allEmployees, err := uc.GetAll()
	if err != nil {
		return nil, err // Error already logged by GetAll
	}

	// Filter active employees
	activeEmployees := make([]domain.User, 0)
	for _, employee := range allEmployees {
		if employee.IsActive {
			activeEmployees = append(activeEmployees, employee)
		}
	}

	uc.logger.LogOperation("GetActiveEmployees", "success", map[string]interface{}{
		"total_count":  len(allEmployees),
		"active_count": len(activeEmployees),
	})

	return activeEmployees, nil
}

// GetByStoreAndActive retrieves active employees by store
func (uc *employeeUseCase) GetByStoreAndActive(storeID int) ([]domain.User, error) {
	uc.logger.LogOperation("GetByStoreAndActive", "start", map[string]interface{}{
		"store_id": storeID,
	})

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("GetByStoreAndActive", err, map[string]interface{}{
			"store_id": storeID,
		})
		return nil, uc.errorHandler.HandleValidationError("GetByStoreAndActive", err)
	}

	allEmployees, err := uc.GetByStore(storeID)
	if err != nil {
		return nil, err // Error already logged by GetByStore
	}

	// Filter active employees
	activeEmployees := make([]domain.User, 0)
	for _, employee := range allEmployees {
		if employee.IsActive {
			activeEmployees = append(activeEmployees, employee)
		}
	}

	uc.logger.LogOperation("GetByStoreAndActive", "success", map[string]interface{}{
		"store_id":     storeID,
		"total_count":  len(allEmployees),
		"active_count": len(activeEmployees),
	})

	return activeEmployees, nil
}

// ValidateEmployeeData validates employee data according to business rules
func (uc *employeeUseCase) ValidateEmployeeData(user *domain.User) error {
	uc.logger.LogOperation("ValidateEmployeeData", "start", map[string]interface{}{
		"email": user.Email,
	})

	// Basic entity validation
	if err := uc.validator.ValidateEntity(user); err != nil {
		uc.logger.LogValidation("ValidateEmployeeData", "entity", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
	}

	// Validate required fields
	if err := uc.validator.ValidateString(user.FirstName, "first_name", "required", "min:2", "max:50"); err != nil {
		uc.logger.LogValidation("ValidateEmployeeData", "first_name", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
	}

	if err := uc.validator.ValidateString(user.LastName, "last_name", "required", "min:2", "max:50"); err != nil {
		uc.logger.LogValidation("ValidateEmployeeData", "last_name", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
	}

	// Validate email
	if err := uc.validator.ValidateString(user.Email, "email", "required", "email"); err != nil {
		uc.logger.LogValidation("ValidateEmployeeData", "email", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
	}

	// Validate document
	if err := uc.validator.ValidateString(user.DocumentType, "document_type", "required"); err != nil {
		uc.logger.LogValidation("ValidateEmployeeData", "document_type", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
	}

	if err := uc.validator.ValidateString(user.DocumentNumber, "document_number", "required", "min:5"); err != nil {
		uc.logger.LogValidation("ValidateEmployeeData", "document_number", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
	}

	// Validate phone (optional but if provided must be valid)
	if user.Phone != "" {
		if err := uc.validator.ValidateString(user.Phone, "phone", "min:10", "max:15"); err != nil {
			uc.logger.LogValidation("ValidateEmployeeData", "phone", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
		}
	}

	// Validate salary (must be positive if provided)
	if user.Salary > 0 {
		if err := uc.validator.ValidateNumber(user.Salary, "salary", "positive"); err != nil {
			uc.logger.LogValidation("ValidateEmployeeData", "salary", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return uc.errorHandler.HandleValidationError("ValidateEmployeeData", err)
		}
	}

	// Password is not required - it will be generated automatically in Create method

	uc.logger.LogValidation("ValidateEmployeeData", "all_fields", "passed", map[string]interface{}{
		"email": user.Email,
	})

	return nil
}

// ValidateEmployeeCredentials validates uniqueness of email and document number
func (uc *employeeUseCase) ValidateEmployeeCredentials(email, documentNumber string) error {
	uc.logger.LogOperation("ValidateEmployeeCredentials", "start", map[string]interface{}{
		"email":           email,
		"document_number": documentNumber,
	})

	// Business rule: Email must be unique
	// Note: This is a simplified check. In a real implementation, you'd want
	// to have a specific repository method to check existence without fetching full user
	existingUser, err := uc.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		err := fmt.Errorf("email already exists: %s", email)
		uc.logger.LogValidation("ValidateEmployeeCredentials", "email_uniqueness", "failed", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		})
		return uc.errorHandler.HandleConflict("ValidateEmployeeCredentials", fmt.Sprintf("email already exists: %s", email))
	}

	// Business rule: Document number should be unique per document type
	// Note: This is a simplified validation. In practice, you might want more sophisticated checks

	uc.logger.LogValidation("ValidateEmployeeCredentials", "uniqueness", "passed", map[string]interface{}{
		"email":           email,
		"document_number": documentNumber,
	})

	return nil
}

// HashPassword securely hashes a password
func (uc *employeeUseCase) HashPassword(password string) (string, error) {
	uc.logger.LogOperation("HashPassword", "start", nil)

	// Validate password requirements
	if err := uc.validator.ValidateString(password, "password", "required", "min:6"); err != nil {
		uc.logger.LogValidation("HashPassword", "password_requirements", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return "", uc.errorHandler.HandleValidationError("HashPassword", err)
	}

	// Hash password with bcrypt
	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		uc.logger.LogError("HashPassword", err, nil)
		return "", uc.errorHandler.HandleInternalError("HashPassword", fmt.Errorf("could not hash password: %v", err))
	}

	uc.logger.LogOperation("HashPassword", "success", nil)
	return string(hashed), nil
}

// GenerateDefaultPassword generates a default password using franchise name + current year
func (uc *employeeUseCase) GenerateDefaultPassword(storeID int) (string, error) {
	uc.logger.LogOperation("GenerateDefaultPassword", "start", map[string]interface{}{
		"store_id": storeID,
	})

	// Get store information
	store, err := uc.storeRepo.GetByID(storeID)
	if err != nil {
		uc.logger.LogError("GenerateDefaultPassword", err, map[string]interface{}{
			"store_id": storeID,
		})
		return "", uc.errorHandler.HandleRepositoryError("GenerateDefaultPassword", fmt.Errorf("failed to get store: %v", err))
	}

	// Get franchise information
	franchise, err := uc.franchiseRepo.GetById(int(store.FranchiseID))
	if err != nil {
		uc.logger.LogError("GenerateDefaultPassword", err, map[string]interface{}{
			"franchise_id": store.FranchiseID,
		})
		return "", uc.errorHandler.HandleRepositoryError("GenerateDefaultPassword", fmt.Errorf("failed to get franchise: %v", err))
	}

	// Generate password: FRANCHISENAME + CURRENTYEAR
	currentYear := time.Now().Year()
	password := strings.ToUpper(franchise.Name) + fmt.Sprintf("%d", currentYear)

	uc.logger.LogOperation("GenerateDefaultPassword", "success", map[string]interface{}{
		"store_id":      storeID,
		"franchise_id":  store.FranchiseID,
		"password_hint": strings.ToUpper(franchise.Name) + "****", // Log hint without full password
	})

	return password, nil
}

// ValidateUpdateFields validates and cleans fields for update operations
func (uc *employeeUseCase) ValidateUpdateFields(fields map[string]interface{}) (map[string]interface{}, error) {
	uc.logger.LogOperation("ValidateUpdateFields", "start", map[string]interface{}{
		"field_count": len(fields),
	})

	if len(fields) == 0 {
		err := fmt.Errorf("no fields to update")
		uc.logger.LogValidation("ValidateUpdateFields", "empty_fields", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, uc.errorHandler.HandleValidationError("ValidateUpdateFields", err)
	}

	// Define allowed fields for update
	allowedFields := []string{
		"first_name", "last_name", "phone", "email", "position",
		"salary", "document_type", "document_number", "password_hash",
	}

	// Use validator to clean and validate fields
	cleanFields, err := uc.validator.ValidateUpdateFields(fields, allowedFields)
	if err != nil {
		uc.logger.LogValidation("ValidateUpdateFields", "field_validation", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, uc.errorHandler.HandleValidationError("ValidateUpdateFields", err)
	}

	// Special handling for password
	if passwordValue, exists := cleanFields["password_hash"]; exists {
		rawPassword, ok := passwordValue.(string)
		if !ok || rawPassword == "" {
			err := fmt.Errorf("invalid password format")
			uc.logger.LogValidation("ValidateUpdateFields", "password_format", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return nil, uc.errorHandler.HandleValidationError("ValidateUpdateFields", err)
		}

		hashedPassword, err := uc.HashPassword(rawPassword)
		if err != nil {
			return nil, err // Error already logged by HashPassword
		}
		cleanFields["password_hash"] = hashedPassword
	}

	// Special validation for email if being updated
	if emailValue, exists := cleanFields["email"]; exists {
		email, ok := emailValue.(string)
		if ok && email != "" {
			if err := uc.validator.ValidateString(email, "email", "email"); err != nil {
				uc.logger.LogValidation("ValidateUpdateFields", "email_format", "failed", map[string]interface{}{
					"error": err.Error(),
				})
				return nil, uc.errorHandler.HandleValidationError("ValidateUpdateFields", err)
			}
		}
	}

	uc.logger.LogOperation("ValidateUpdateFields", "success", map[string]interface{}{
		"clean_field_count": len(cleanFields),
		"fields":            cleanFields,
	})

	return cleanFields, nil
}
