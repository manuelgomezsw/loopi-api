package mysql

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

// absenceRepository implements repository.AbsenceRepository with improved maintainability
type absenceRepository struct {
	*BaseRepository[domain.Absence]
	errorHandler *ErrorHandler
}

// NewAbsenceRepository creates a new absence repository with enhanced features
func NewAbsenceRepository(db *gorm.DB) repository.AbsenceRepository {
	return &absenceRepository{
		BaseRepository: NewBaseRepository[domain.Absence](db, "absences"),
		errorHandler:   NewErrorHandler("absences"),
	}
}

// GetByEmployeeAndMonth retrieves absences by employee and month using query helper
func (r *absenceRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error) {
	absences, err := FindByEmployeeAndMonth[domain.Absence](r.GetDB(), employeeID, year, month)
	if err != nil {
		return nil, r.errorHandler.HandleError("GetByEmployeeAndMonth", err, employeeID)
	}
	return absences, nil
}

// Create creates a new absence with validation and error handling
func (r *absenceRepository) Create(absence *domain.Absence) error {
	// Business validation before creation
	if err := r.validateAbsence(absence); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	// Set timestamps if not already set
	if absence.CreatedAt.IsZero() {
		absence.CreatedAt = time.Now()
	}
	absence.UpdatedAt = time.Now()

	if err := r.BaseRepository.Create(absence); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// validateAbsence performs business validation
func (r *absenceRepository) validateAbsence(absence *domain.Absence) error {
	if absence.EmployeeID <= 0 {
		return ErrInvalidInput
	}
	if absence.Date.IsZero() {
		return ErrInvalidInput
	}
	if absence.Hours <= 0 {
		return ErrInvalidInput
	}
	return nil
}

// GetByEmployeeAndDateRange retrieves absences by employee and custom date range
func (r *absenceRepository) GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Absence, error) {
	absences, err := FindByStoreAndDateRange[domain.Absence](r.GetDB(), employeeID, from, to)
	if err != nil {
		return nil, r.errorHandler.HandleError("GetByEmployeeAndDateRange", err, employeeID)
	}
	return absences, nil
}

// GetTotalHoursByEmployee retrieves total absence hours for an employee in a period
func (r *absenceRepository) GetTotalHoursByEmployee(employeeID, year, month int) (float64, error) {
	var totalHours float64

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.GetDB().
		Model(&domain.Absence{}).
		Select("COALESCE(SUM(hours), 0)").
		Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, startDate, endDate).
		Scan(&totalHours).Error

	if err != nil {
		return 0, r.errorHandler.HandleError("GetTotalHoursByEmployee", err, employeeID)
	}

	return totalHours, nil
}
