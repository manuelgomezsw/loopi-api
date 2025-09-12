package mysql

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

// assignedShiftRepository implements repository.AssignedShiftRepository with improved maintainability
type assignedShiftRepository struct {
	*BaseRepository[domain.AssignedShift]
	errorHandler *ErrorHandler
}

// NewAssignedShiftRepository creates a new assigned shift repository with enhanced features
func NewAssignedShiftRepository(db *gorm.DB) repository.AssignedShiftRepository {
	return &assignedShiftRepository{
		BaseRepository: NewBaseRepository[domain.AssignedShift](db, "assigned_shifts"),
		errorHandler:   NewErrorHandler("assigned_shifts"),
	}
}

// GetByEmployeeAndMonth retrieves assigned shifts by employee and month using query helper
func (r *assignedShiftRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.AssignedShift, error) {
	shifts, err := FindByEmployeeAndMonth[domain.AssignedShift](r.GetDB(), employeeID, year, month)
	if err != nil {
		return nil, r.errorHandler.HandleError("GetByEmployeeAndMonth", err, employeeID)
	}
	return shifts, nil
}

// Create creates a new assigned shift with validation and error handling
func (r *assignedShiftRepository) Create(assignedShift *domain.AssignedShift) error {
	// Business validation before creation
	if err := r.validateAssignedShift(assignedShift); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	if err := r.BaseRepository.Create(assignedShift); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// validateAssignedShift performs business validation
func (r *assignedShiftRepository) validateAssignedShift(assignedShift *domain.AssignedShift) error {
	if assignedShift.Date == "" {
		return ErrInvalidInput
	}
	if assignedShift.StartTime == "" {
		return ErrInvalidInput
	}
	if assignedShift.EndTime == "" {
		return ErrInvalidInput
	}
	return nil
}

// GetByDateRange retrieves assigned shifts by custom date range
func (r *assignedShiftRepository) GetByDateRange(from, to time.Time) ([]domain.AssignedShift, error) {
	var shifts []domain.AssignedShift

	fromStr := from.Format("2006-01-02")
	toStr := to.Format("2006-01-02")

	err := NewQueryBuilder(r.GetDB()).
		WhereLike("date", fromStr+"' AND '"+toStr).
		OrderBy("date").
		GetDB().
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetByDateRange", err)
	}
	return shifts, nil
}

// GetByDateAndTime retrieves assigned shifts by specific date and time range
func (r *assignedShiftRepository) GetByDateAndTime(date, startTime, endTime string) ([]domain.AssignedShift, error) {
	var shifts []domain.AssignedShift

	err := NewQueryBuilder(r.GetDB()).
		WhereEquals("date", date).
		WhereLike("start_time", startTime).
		WhereLike("end_time", endTime).
		GetDB().
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetByDateAndTime", err)
	}
	return shifts, nil
}

// GetByPeriod retrieves assigned shifts by specific time period
func (r *assignedShiftRepository) GetByPeriod(startDate, endDate string) ([]domain.AssignedShift, error) {
	var shifts []domain.AssignedShift

	err := r.GetDB().
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Order("date").
		Find(&shifts).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetByPeriod", err)
	}

	return shifts, nil
}

// CountShiftsByDateRange returns count of assigned shifts in date range
func (r *assignedShiftRepository) CountShiftsByDateRange(startDate, endDate string) (int64, error) {
	var count int64

	err := r.GetDB().
		Model(&domain.AssignedShift{}).
		Where("date BETWEEN ? AND ?", startDate, endDate).
		Count(&count).Error

	if err != nil {
		return 0, r.errorHandler.HandleError("CountShiftsByDateRange", err)
	}

	return count, nil
}
