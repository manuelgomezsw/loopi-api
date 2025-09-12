package mysql

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"time"

	"gorm.io/gorm"
)

// noveltyRepository implements repository.NoveltyRepository with improved maintainability
type noveltyRepository struct {
	*BaseRepository[domain.Novelty]
	errorHandler *ErrorHandler
}

// NewNoveltyRepository creates a new novelty repository with enhanced features
func NewNoveltyRepository(db *gorm.DB) repository.NoveltyRepository {
	return &noveltyRepository{
		BaseRepository: NewBaseRepository[domain.Novelty](db, "novelties"),
		errorHandler:   NewErrorHandler("novelties"),
	}
}

// GetByEmployeeAndMonth retrieves novelties by employee and month using query helper
func (r *noveltyRepository) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error) {
	novelties, err := FindByEmployeeAndMonth[domain.Novelty](r.GetDB(), employeeID, year, month)
	if err != nil {
		return nil, r.errorHandler.HandleError("GetByEmployeeAndMonth", err, employeeID)
	}
	return novelties, nil
}

// Create creates a new novelty with validation and error handling
func (r *noveltyRepository) Create(novelty *domain.Novelty) error {
	// Business validation before creation
	if err := r.validateNovelty(novelty); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	// Set timestamps if not already set
	if novelty.CreatedAt.IsZero() {
		novelty.CreatedAt = time.Now()
	}
	novelty.UpdatedAt = time.Now()

	if err := r.BaseRepository.Create(novelty); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// validateNovelty performs business validation
func (r *noveltyRepository) validateNovelty(novelty *domain.Novelty) error {
	if novelty.EmployeeID <= 0 {
		return ErrInvalidInput
	}
	if novelty.Date.IsZero() {
		return ErrInvalidInput
	}
	if novelty.Hours <= 0 {
		return ErrInvalidInput
	}
	if novelty.Type == "" {
		return ErrInvalidInput
	}
	return nil
}

// GetByEmployeeAndDateRange retrieves novelties by employee and custom date range
func (r *noveltyRepository) GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Novelty, error) {
	novelties, err := FindByStoreAndDateRange[domain.Novelty](r.GetDB(), employeeID, from, to)
	if err != nil {
		return nil, r.errorHandler.HandleError("GetByEmployeeAndDateRange", err, employeeID)
	}
	return novelties, nil
}

// GetTotalHoursByEmployeeAndType retrieves total novelty hours by type for an employee
func (r *noveltyRepository) GetTotalHoursByEmployeeAndType(employeeID, year, month int, noveltyType string) (float64, error) {
	var totalHours float64

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.GetDB().
		Model(&domain.Novelty{}).
		Select("COALESCE(SUM(hours), 0)").
		Where("employee_id = ? AND type = ? AND date BETWEEN ? AND ?", employeeID, noveltyType, startDate, endDate).
		Scan(&totalHours).Error

	if err != nil {
		return 0, r.errorHandler.HandleError("GetTotalHoursByEmployeeAndType", err, employeeID)
	}

	return totalHours, nil
}

// GetNoveltyTypesSummary retrieves a summary of novelty types for an employee in a month
func (r *noveltyRepository) GetNoveltyTypesSummary(employeeID, year, month int) ([]repository.NoveltyTypeSummary, error) {
	var summary []repository.NoveltyTypeSummary

	startDate := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
	endDate := startDate.AddDate(0, 1, -1).Add(23*time.Hour + 59*time.Minute + 59*time.Second)

	err := r.GetDB().
		Model(&domain.Novelty{}).
		Select("type, COUNT(*) as count, SUM(hours) as total_hours").
		Where("employee_id = ? AND date BETWEEN ? AND ?", employeeID, startDate, endDate).
		Group("type").
		Order("type").
		Scan(&summary).Error

	if err != nil {
		return nil, r.errorHandler.HandleError("GetNoveltyTypesSummary", err, employeeID)
	}

	return summary, nil
}
