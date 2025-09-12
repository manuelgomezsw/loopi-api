package usecase

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
)

type FranchiseUseCase interface {
	// Standard CRUD operations
	GetAll() ([]domain.Franchise, error)
	GetById(id int) (domain.Franchise, error)
	Create(franchise domain.Franchise) error

	// Business-specific operations
	GetActiveFranchises() ([]domain.Franchise, error)
	ValidateFranchiseData(franchise *domain.Franchise) error
}

type franchiseUseCase struct {
	repo         repository.FranchiseRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewFranchiseUseCase(repo repository.FranchiseRepository) FranchiseUseCase {
	return &franchiseUseCase{
		repo:         repo,
		errorHandler: base.NewErrorHandler("Franchise"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Franchise"),
	}
}

// ✅ Enhanced CRUD operations with logging, validation, and error handling

// GetAll retrieves all franchises with proper error handling and logging
func (uc *franchiseUseCase) GetAll() ([]domain.Franchise, error) {
	uc.logger.LogOperation("GetAll", "start", nil)

	franchises, err := uc.repo.GetAll()
	if err != nil {
		uc.logger.LogError("GetAll", err, nil)
		return nil, uc.errorHandler.HandleRepositoryError("GetAll", err)
	}

	if len(franchises) == 0 {
		uc.logger.LogOperation("GetAll", "no_franchises_found", nil)
		return nil, uc.errorHandler.HandleNotFound("GetAll", "No franchises found")
	}

	uc.logger.LogOperation("GetAll", "success", map[string]interface{}{
		"count": len(franchises),
	})

	return franchises, nil
}

// GetById retrieves a franchise by ID with validation and error handling
func (uc *franchiseUseCase) GetById(id int) (domain.Franchise, error) {
	uc.logger.LogOperation("GetById", "start", map[string]interface{}{"id": id})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("GetById", err, map[string]interface{}{"id": id})
		return domain.Franchise{}, uc.errorHandler.HandleValidationError("GetById", err)
	}

	franchise, err := uc.repo.GetById(id)
	if err != nil {
		uc.logger.LogError("GetById", err, map[string]interface{}{"id": id})
		return domain.Franchise{}, uc.errorHandler.HandleRepositoryError("GetById", err)
	}

	uc.logger.LogOperation("GetById", "success", map[string]interface{}{"id": id})
	return franchise, nil
}

// Create creates a new franchise with validation and business rules
func (uc *franchiseUseCase) Create(franchise domain.Franchise) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"franchise_name": franchise.Name,
	})

	// Validate business rules
	if err := uc.ValidateFranchiseData(&franchise); err != nil {
		return err
	}

	// Set default values (business rule)
	franchise.IsActive = true

	// Execute creation
	if err := uc.repo.Create(&franchise); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"franchise_name": franchise.Name,
		})
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"franchise_id":   franchise.ID,
		"franchise_name": franchise.Name,
	})

	return nil
}

// GetActiveFranchises retrieves only active franchises with business rule filtering
func (uc *franchiseUseCase) GetActiveFranchises() ([]domain.Franchise, error) {
	// Start performance timer
	timer := uc.logger.StartTimer("GetActiveFranchises", nil)
	defer timer.Stop()

	// Get all franchises
	franchises, err := uc.repo.GetAll()
	if err != nil {
		uc.logger.LogError("GetActiveFranchises", err, nil)
		return nil, uc.errorHandler.HandleRepositoryError("GetActiveFranchises", err)
	}

	// Apply business rule: filter only active franchises
	activeFranchises := make([]domain.Franchise, 0)
	for _, franchise := range franchises {
		if franchise.IsActive {
			activeFranchises = append(activeFranchises, franchise)
		}
	}

	uc.logger.LogBusinessRule("GetActiveFranchises", "filter_active_only", "applied", map[string]interface{}{
		"total_franchises":  len(franchises),
		"active_franchises": len(activeFranchises),
	})

	uc.logger.LogOperation("GetActiveFranchises", "success", map[string]interface{}{
		"active_count": len(activeFranchises),
	})

	return activeFranchises, nil
}

// ValidateFranchiseData validates franchise data according to business rules
func (uc *franchiseUseCase) ValidateFranchiseData(franchise *domain.Franchise) error {
	// Basic entity validation
	if err := uc.validator.ValidateEntity(franchise); err != nil {
		uc.logger.LogValidation("ValidateFranchiseData", "entity", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule validations
	if err := uc.validator.ValidateString(franchise.Name, "name", "required", "min:3", "max:100"); err != nil {
		uc.logger.LogValidation("ValidateFranchiseData", "name", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	uc.logger.LogValidation("ValidateFranchiseData", "all_fields", "passed", nil)
	return nil
}
