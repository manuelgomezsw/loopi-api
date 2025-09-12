package usecase

import (
	"fmt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"regexp"
	"strings"
)

type StoreUseCase interface {
	// Standard CRUD operations
	GetAll() ([]domain.Store, error)
	GetByID(id int) (domain.Store, error)
	GetByFranchiseID(franchiseID int) ([]domain.Store, error)
	Create(store *domain.Store) error
	Update(store *domain.Store) error
	Delete(id int) error

	// Business-specific operations
	GetActiveStoresByFranchise(franchiseID int) ([]domain.Store, error)
	GetStoresWithEmployeeCount(franchiseID int) ([]repository.StoreWithEmployeeCount, error)
	GetStoreStatistics(storeID int) (*repository.StoreStatistics, error)
	ValidateStoreData(store *domain.Store) error
	ValidateStoreCode(code string) error
	ValidateUpdateFields(fields map[string]interface{}) (map[string]interface{}, error)
}

type storeUseCase struct {
	repo         repository.StoreRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewStoreUseCase(repo repository.StoreRepository) StoreUseCase {
	return &storeUseCase{
		repo:         repo,
		errorHandler: base.NewErrorHandler("Store"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Store"),
	}
}

// ✅ Enhanced CRUD operations with logging, validation, and error handling

// GetAll retrieves all stores with proper error handling and logging
func (uc *storeUseCase) GetAll() ([]domain.Store, error) {
	uc.logger.LogOperation("GetAll", "start", nil)

	stores, err := uc.repo.GetAll()
	if err != nil {
		uc.logger.LogError("GetAll", err, nil)
		return nil, uc.errorHandler.HandleRepositoryError("GetAll", err)
	}

	if len(stores) == 0 {
		uc.logger.LogOperation("GetAll", "no_stores_found", nil)
		return nil, uc.errorHandler.HandleNotFound("GetAll", "No stores found")
	}

	uc.logger.LogOperation("GetAll", "success", map[string]interface{}{
		"count": len(stores),
	})

	return stores, nil
}

// GetByID retrieves a store by ID with validation and error handling
func (uc *storeUseCase) GetByID(id int) (domain.Store, error) {
	uc.logger.LogOperation("GetByID", "start", map[string]interface{}{"id": id})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("GetByID", err, map[string]interface{}{"id": id})
		return domain.Store{}, uc.errorHandler.HandleValidationError("GetByID", err)
	}

	store, err := uc.repo.GetByID(id)
	if err != nil {
		uc.logger.LogError("GetByID", err, map[string]interface{}{"id": id})
		return domain.Store{}, uc.errorHandler.HandleRepositoryError("GetByID", err)
	}

	uc.logger.LogOperation("GetByID", "success", map[string]interface{}{"id": id})
	return store, nil
}

// GetByFranchiseID retrieves stores by franchise ID with validation and error handling
func (uc *storeUseCase) GetByFranchiseID(franchiseID int) ([]domain.Store, error) {
	uc.logger.LogOperation("GetByFranchiseID", "start", map[string]interface{}{"franchise_id": franchiseID})

	// Validate franchise ID
	if err := uc.validator.ValidateID(franchiseID); err != nil {
		uc.logger.LogError("GetByFranchiseID", err, map[string]interface{}{"franchise_id": franchiseID})
		return nil, uc.errorHandler.HandleValidationError("GetByFranchiseID", err)
	}

	stores, err := uc.repo.GetByFranchiseID(franchiseID)
	if err != nil {
		uc.logger.LogError("GetByFranchiseID", err, map[string]interface{}{"franchise_id": franchiseID})
		return nil, uc.errorHandler.HandleRepositoryError("GetByFranchiseID", err)
	}

	uc.logger.LogOperation("GetByFranchiseID", "success", map[string]interface{}{
		"franchise_id": franchiseID,
		"count":        len(stores),
	})

	return stores, nil
}

// Create creates a new store with validation and business rules
func (uc *storeUseCase) Create(store *domain.Store) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"store_name": store.Name,
		"store_code": store.Code,
	})

	// Validate business rules
	if err := uc.ValidateStoreData(store); err != nil {
		return err
	}

	// Set default values (business rule)
	store.IsActive = true

	// Execute creation
	if err := uc.repo.Create(store); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"store_name": store.Name,
			"store_code": store.Code,
		})
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"store_id":   store.ID,
		"store_name": store.Name,
		"store_code": store.Code,
	})

	return nil
}

// Update updates an existing store with validation and business rules
func (uc *storeUseCase) Update(store *domain.Store) error {
	uc.logger.LogOperation("Update", "start", map[string]interface{}{
		"store_id":   store.ID,
		"store_name": store.Name,
	})

	// Validate ID
	if err := uc.validator.ValidateID(int(store.ID)); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{"store_id": store.ID})
		return uc.errorHandler.HandleValidationError("Update", err)
	}

	// Validate business rules
	if err := uc.ValidateStoreData(store); err != nil {
		return err
	}

	// Execute update
	if err := uc.repo.Update(store); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{
			"store_id":   store.ID,
			"store_name": store.Name,
		})
		return uc.errorHandler.HandleRepositoryError("Update", err)
	}

	uc.logger.LogOperation("Update", "success", map[string]interface{}{
		"store_id":   store.ID,
		"store_name": store.Name,
	})

	return nil
}

// Delete removes a store by ID with validation and business rules
func (uc *storeUseCase) Delete(id int) error {
	uc.logger.LogOperation("Delete", "start", map[string]interface{}{"id": id})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{"id": id})
		return uc.errorHandler.HandleValidationError("Delete", err)
	}

	// Execute deletion
	if err := uc.repo.Delete(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{"id": id})
		return uc.errorHandler.HandleRepositoryError("Delete", err)
	}

	uc.logger.LogOperation("Delete", "success", map[string]interface{}{"id": id})
	return nil
}

// ✅ Business-specific operations with enhanced features

// GetActiveStoresByFranchise retrieves only active stores for a franchise with business rule filtering
func (uc *storeUseCase) GetActiveStoresByFranchise(franchiseID int) ([]domain.Store, error) {
	// Start performance timer
	timer := uc.logger.StartTimer("GetActiveStoresByFranchise", map[string]interface{}{"franchise_id": franchiseID})
	defer timer.Stop()

	// Validate franchise ID
	if err := uc.validator.ValidateID(franchiseID); err != nil {
		uc.logger.LogError("GetActiveStoresByFranchise", err, map[string]interface{}{"franchise_id": franchiseID})
		return nil, uc.errorHandler.HandleValidationError("GetActiveStoresByFranchise", err)
	}

	// Use repository method for active stores
	stores, err := uc.repo.GetActiveStoresByFranchise(franchiseID)
	if err != nil {
		uc.logger.LogError("GetActiveStoresByFranchise", err, map[string]interface{}{"franchise_id": franchiseID})
		return nil, uc.errorHandler.HandleRepositoryError("GetActiveStoresByFranchise", err)
	}

	uc.logger.LogBusinessRule("GetActiveStoresByFranchise", "filter_active_only", "applied", map[string]interface{}{
		"franchise_id":  franchiseID,
		"active_stores": len(stores),
	})

	uc.logger.LogOperation("GetActiveStoresByFranchise", "success", map[string]interface{}{
		"franchise_id": franchiseID,
		"active_count": len(stores),
	})

	return stores, nil
}

// GetStoresWithEmployeeCount retrieves stores with their employee counts
func (uc *storeUseCase) GetStoresWithEmployeeCount(franchiseID int) ([]repository.StoreWithEmployeeCount, error) {
	uc.logger.LogOperation("GetStoresWithEmployeeCount", "start", map[string]interface{}{"franchise_id": franchiseID})

	// Validate franchise ID
	if err := uc.validator.ValidateID(franchiseID); err != nil {
		uc.logger.LogError("GetStoresWithEmployeeCount", err, map[string]interface{}{"franchise_id": franchiseID})
		return nil, uc.errorHandler.HandleValidationError("GetStoresWithEmployeeCount", err)
	}

	// Use repository method for stores with employee count
	stores, err := uc.repo.GetStoresWithEmployeeCount(franchiseID)
	if err != nil {
		uc.logger.LogError("GetStoresWithEmployeeCount", err, map[string]interface{}{"franchise_id": franchiseID})
		return nil, uc.errorHandler.HandleRepositoryError("GetStoresWithEmployeeCount", err)
	}

	uc.logger.LogOperation("GetStoresWithEmployeeCount", "success", map[string]interface{}{
		"franchise_id": franchiseID,
		"count":        len(stores),
	})

	return stores, nil
}

// GetStoreStatistics retrieves comprehensive store statistics
func (uc *storeUseCase) GetStoreStatistics(storeID int) (*repository.StoreStatistics, error) {
	uc.logger.LogOperation("GetStoreStatistics", "start", map[string]interface{}{"store_id": storeID})

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("GetStoreStatistics", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleValidationError("GetStoreStatistics", err)
	}

	// Use repository method for statistics
	stats, err := uc.repo.GetStoreStatistics(storeID)
	if err != nil {
		uc.logger.LogError("GetStoreStatistics", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleRepositoryError("GetStoreStatistics", err)
	}

	uc.logger.LogOperation("GetStoreStatistics", "success", map[string]interface{}{
		"store_id":       storeID,
		"employee_count": stats.EmployeeCount,
		"shift_count":    stats.ShiftCount,
	})

	return stats, nil
}

// ValidateStoreData validates store data according to business rules
func (uc *storeUseCase) ValidateStoreData(store *domain.Store) error {
	// Basic entity validation
	if err := uc.validator.ValidateEntity(store); err != nil {
		uc.logger.LogValidation("ValidateStoreData", "entity", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule validations
	if err := uc.validator.ValidateString(store.Name, "name", "required", "min:3", "max:100"); err != nil {
		uc.logger.LogValidation("ValidateStoreData", "name", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate store code
	if err := uc.ValidateStoreCode(store.Code); err != nil {
		return err
	}

	// Validate franchise ID
	if err := uc.validator.ValidateNumber(store.FranchiseID, "franchise_id", "positive"); err != nil {
		uc.logger.LogValidation("ValidateStoreData", "franchise_id", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate location (optional but if provided, must be valid)
	if store.Location != "" {
		if err := uc.validator.ValidateString(store.Location, "location", "max:255"); err != nil {
			uc.logger.LogValidation("ValidateStoreData", "location", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	// Validate address (optional but if provided, must be valid)
	if store.Address != "" {
		if err := uc.validator.ValidateString(store.Address, "address", "max:255"); err != nil {
			uc.logger.LogValidation("ValidateStoreData", "address", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	uc.logger.LogValidation("ValidateStoreData", "all_fields", "passed", nil)
	return nil
}

// ValidateStoreCode validates store code according to business rules
func (uc *storeUseCase) ValidateStoreCode(code string) error {
	uc.logger.LogOperation("ValidateStoreCode", "start", map[string]interface{}{
		"code": code,
	})

	// Business rule: Store code is required
	if strings.TrimSpace(code) == "" {
		err := fmt.Errorf("store code is required")
		uc.logger.LogValidation("ValidateStoreCode", "required", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule: Store code must be exactly 3 characters
	if len(code) != 3 {
		err := fmt.Errorf("store code must be exactly 3 characters, got: %d", len(code))
		uc.logger.LogValidation("ValidateStoreCode", "length", "failed", map[string]interface{}{
			"error":       err.Error(),
			"code_length": len(code),
		})
		return err
	}

	// Business rule: Store code must contain only uppercase letters and numbers
	alphanumericRegex := regexp.MustCompile(`^[A-Z0-9]{3}$`)
	if !alphanumericRegex.MatchString(code) {
		err := fmt.Errorf("store code must contain only uppercase letters and numbers: %s", code)
		uc.logger.LogValidation("ValidateStoreCode", "format", "failed", map[string]interface{}{
			"error": err.Error(),
			"code":  code,
		})
		return err
	}

	uc.logger.LogValidation("ValidateStoreCode", "all_code_rules", "passed", map[string]interface{}{
		"code": code,
	})

	return nil
}

// ValidateUpdateFields validates fields for update operations
func (uc *storeUseCase) ValidateUpdateFields(fields map[string]interface{}) (map[string]interface{}, error) {
	uc.logger.LogOperation("ValidateUpdateFields", "start", map[string]interface{}{
		"field_count": len(fields),
	})

	// Define allowed fields for update
	allowedFields := []string{
		"name", "code", "location", "address", "is_active",
	}

	// Use validator to clean and validate fields
	cleanedFields, err := uc.validator.ValidateUpdateFields(fields, allowedFields)
	if err != nil {
		uc.logger.LogValidation("ValidateUpdateFields", "validation", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return nil, uc.errorHandler.HandleValidationError("ValidateUpdateFields", err)
	}

	// Additional business rule validations for specific fields
	if code, exists := cleanedFields["code"]; exists {
		if codeStr, ok := code.(string); ok {
			if err := uc.ValidateStoreCode(codeStr); err != nil {
				return nil, err
			}
		}
	}

	if name, exists := cleanedFields["name"]; exists {
		if nameStr, ok := name.(string); ok {
			if err := uc.validator.ValidateString(nameStr, "name", "required", "min:3", "max:100"); err != nil {
				uc.logger.LogValidation("ValidateUpdateFields", "name", "failed", map[string]interface{}{
					"error": err.Error(),
				})
				return nil, err
			}
		}
	}

	uc.logger.LogOperation("ValidateUpdateFields", "success", map[string]interface{}{
		"cleaned_field_count": len(cleanedFields),
	})

	return cleanedFields, nil
}
