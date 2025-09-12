package mysql

import (
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
	if shift.Period == "" {
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

// GetShiftsByPeriod retrieves shifts by period (morning, afternoon, night)
func (r *shiftRepository) GetShiftsByPeriod(period string) ([]domain.Shift, error) {
	var shifts []domain.Shift
	err := NewQueryBuilder(r.GetDB()).
		WhereEquals("period", period).
		WhereActive().
		OrderBy("start_time").
		GetDB().
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetShiftsByPeriod", err)
	}
	return shifts, nil
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

		// Get shifts by period
		if err := tx.Model(&domain.Shift{}).
			Select("period, COUNT(*) as count").
			Where("store_id = ? AND is_active = ?", storeID, true).
			Group("period").
			Scan(&stats.ShiftsByPeriod).Error; err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, r.errorHandler.HandleError("GetShiftStatistics", err, storeID)
	}

	return &stats, nil
}
