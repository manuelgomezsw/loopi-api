package mysql

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"

	"gorm.io/gorm"
)

// franchiseRepository implements repository.FranchiseRepository with improved maintainability
type franchiseRepository struct {
	*BaseRepository[domain.Franchise]
	errorHandler *ErrorHandler
}

// NewFranchiseRepository creates a new franchise repository with enhanced features
func NewFranchiseRepository(db *gorm.DB) repository.FranchiseRepository {
	return &franchiseRepository{
		BaseRepository: NewBaseRepository[domain.Franchise](db, "franchises"),
		errorHandler:   NewErrorHandler("franchises"),
	}
}

// GetAll retrieves all franchises (uses base repository)
func (r *franchiseRepository) GetAll() ([]domain.Franchise, error) {
	franchises, err := r.BaseRepository.GetAll()
	if err != nil {
		return nil, r.errorHandler.HandleError("GetAll", err)
	}
	return franchises, nil
}

// GetById retrieves a franchise by ID with proper error handling
func (r *franchiseRepository) GetById(id int) (domain.Franchise, error) {
	franchise, err := r.BaseRepository.GetByID(id)
	if err != nil {
		if err == ErrNotFound {
			return domain.Franchise{}, r.errorHandler.HandleNotFound("GetById", id)
		}
		return domain.Franchise{}, r.errorHandler.HandleError("GetById", err, id)
	}
	return *franchise, nil
}

// Create creates a new franchise with validation and error handling
func (r *franchiseRepository) Create(franchise *domain.Franchise) error {
	// Business validation before creation
	if err := r.validateFranchise(franchise); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	if err := r.BaseRepository.Create(franchise); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// validateFranchise performs business validation
func (r *franchiseRepository) validateFranchise(franchise *domain.Franchise) error {
	if franchise.Name == "" {
		return ErrInvalidInput
	}
	return nil
}

// GetActiveFranchises retrieves only active franchises
func (r *franchiseRepository) GetActiveFranchises() ([]domain.Franchise, error) {
	var franchises []domain.Franchise
	err := NewQueryBuilder(r.GetDB()).
		WhereActive().
		WhereNotDeleted().
		OrderBy("name").
		GetDB().
		Find(&franchises).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetActiveFranchises", err)
	}
	return franchises, nil
}
