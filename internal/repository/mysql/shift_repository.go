package mysql

import (
	"errors"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"

	"gorm.io/gorm"
)

// shiftRepository implements repository.ShiftRepository with improved maintainability
type shiftRepository struct {
	*BaseRepository[domain.Shift]
	errorHandler *ErrorHandler
}

// NewShiftRepository creates a new shift repository with enhanced features
func NewShiftRepository(db *gorm.DB) repository.ShiftRepository {
	return &shiftRepository{
		BaseRepository: NewBaseRepository[domain.Shift](db, "shifts"),
		errorHandler:   NewErrorHandler("shifts"),
	}
}

// Create creates a new shift with validation and error handling
func (r *shiftRepository) Create(shift domain.Shift) error {
	// Business validation before creation
	if err := r.validateShift(&shift); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	if err := r.BaseRepository.Create(&shift); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// ListAll retrieves all shifts with proper ordering and error handling
func (r *shiftRepository) ListAll() ([]domain.Shift, error) {
	var shifts []domain.Shift
	err := NewQueryBuilder(r.GetDB()).
		OrderBy("name").
		GetDB().
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("ListAll", err)
	}
	return shifts, nil
}

// ListByStore retrieves shifts by store ID with proper ordering
func (r *shiftRepository) ListByStore(storeID int) ([]domain.Shift, error) {
	var shifts []domain.Shift
	err := NewQueryBuilder(r.GetDB()).
		WhereEquals("store_id", storeID).
		OrderBy("name").
		GetDB().
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("ListByStore", err, storeID)
	}
	return shifts, nil
}

// GetByID retrieves a shift by ID with proper error handling
func (r *shiftRepository) GetByID(id int) (*domain.Shift, error) {
	shift, err := r.BaseRepository.GetByID(id)
	if err != nil {
		if err == ErrNotFound {
			return nil, r.errorHandler.HandleNotFound("GetByID", id)
		}
		return nil, r.errorHandler.HandleError("GetByID", err, id)
	}
	return shift, nil
}

// validateShift performs business validation
func (r *shiftRepository) validateShift(shift *domain.Shift) error {
	if shift.Name == "" {
		return ErrInvalidInput
	}
	if shift.StartTime == "" {
		return ErrInvalidInput
	}
	if shift.EndTime == "" {
		return ErrInvalidInput
	}
	if shift.StoreID <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// GetActiveShiftsByStore retrieves only active shifts for a store
func (r *shiftRepository) GetActiveShiftsByStore(storeID int) ([]domain.Shift, error) {
	var shifts []domain.Shift
	err := NewQueryBuilder(r.GetDB()).
		WhereEquals("store_id", storeID).
		WhereActive().
		WhereNotDeleted().
		OrderBy("name").
		GetDB().
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetActiveShiftsByStore", err, storeID)
	}
	return shifts, nil
}

// Update modifies an existing shift
func (r *shiftRepository) Update(shift domain.Shift) error {
	if shift.ID == 0 {
		return r.errorHandler.HandleError("Update", ErrInvalidInput)
	}

	// Validate shift data
	if err := r.validateShift(&shift); err != nil {
		return r.errorHandler.HandleError("Update", err)
	}

	// Check if shift exists
	existingShift, err := r.GetByID(int(shift.ID))
	if err != nil {
		return r.errorHandler.HandleError("Update", err)
	}

	if existingShift == nil {
		return r.errorHandler.HandleError("Update", ErrNotFound)
	}

	// Update the shift
	if err := r.GetDB().Save(&shift).Error; err != nil {
		return r.errorHandler.HandleError("Update", err)
	}

	return nil
}

// Delete removes a shift by ID
func (r *shiftRepository) Delete(id int) error {
	if id <= 0 {
		return r.errorHandler.HandleError("Delete", ErrInvalidInput)
	}

	// Check if shift exists
	var shift domain.Shift
	if err := r.GetDB().First(&shift, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return r.errorHandler.HandleError("Delete", ErrNotFound)
		}
		return r.errorHandler.HandleError("Delete", err)
	}

	// Perform soft delete by setting is_active to false
	if err := r.GetDB().Model(&shift).Update("is_active", false).Error; err != nil {
		return r.errorHandler.HandleError("Delete", err)
	}

	return nil
}

// GetShiftStatistics retrieves comprehensive shift statistics for a store
func (r *shiftRepository) GetShiftStatistics(storeID int) (*repository.ShiftStatistics, error) {
	var stats repository.ShiftStatistics

	// Use transaction for consistency
	err := r.BaseRepository.Transaction(func(tx *gorm.DB) error {
		// Count total shifts
		if err := tx.Model(&domain.Shift{}).
			Where("store_id = ?", storeID).
			Count(&stats.TotalShifts).Error; err != nil {
			return err
		}

		// Count active shifts
		if err := tx.Model(&domain.Shift{}).
			Where("store_id = ? AND is_active = ?", storeID, true).
			Count(&stats.ActiveShifts).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, r.errorHandler.HandleError("GetShiftStatistics", err, storeID)
	}

	return &stats, nil
}
