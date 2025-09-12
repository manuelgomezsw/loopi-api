package usecase

import (
	"fmt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"strings"
	"time"
)

type NoveltyUseCase interface {
	// Standard CRUD operations
	Create(novelty domain.Novelty) error
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error)

	// Business-specific operations
	GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Novelty, error)
	GetTotalHoursByEmployeeAndType(employeeID, year, month int, noveltyType string) (float64, error)
	GetNoveltyTypesSummary(employeeID, year, month int) ([]repository.NoveltyTypeSummary, error)
	ValidateNoveltyData(novelty *domain.Novelty) error
	ValidateNoveltyType(noveltyType string) error
	ValidateNoveltyDate(date time.Time) error
}

type noveltyUseCase struct {
	repo         repository.NoveltyRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewNoveltyUseCase(repo repository.NoveltyRepository) NoveltyUseCase {
	return &noveltyUseCase{
		repo:         repo,
		errorHandler: base.NewErrorHandler("Novelty"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Novelty"),
	}
}

// ✅ Enhanced CRUD operations with logging, validation, and error handling

// Create creates a new novelty with validation and business rules
func (uc *noveltyUseCase) Create(novelty domain.Novelty) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"employee_id": novelty.EmployeeID,
		"date":        novelty.Date.Format("2006-01-02"),
		"hours":       novelty.Hours,
		"type":        novelty.Type,
	})

	// Validate business rules
	if err := uc.ValidateNoveltyData(&novelty); err != nil {
		return err
	}

	// Validate novelty type
	if err := uc.ValidateNoveltyType(novelty.Type); err != nil {
		return err
	}

	// Validate novelty date
	if err := uc.ValidateNoveltyDate(novelty.Date); err != nil {
		return err
	}

	// Set timestamps
	novelty.CreatedAt = time.Now()
	novelty.UpdatedAt = time.Now()

	// Execute creation
	if err := uc.repo.Create(&novelty); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"employee_id": novelty.EmployeeID,
			"date":        novelty.Date.Format("2006-01-02"),
			"type":        novelty.Type,
		})
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"novelty_id":  novelty.ID,
		"employee_id": novelty.EmployeeID,
		"hours":       novelty.Hours,
		"type":        novelty.Type,
	})

	return nil
}

// GetByEmployeeAndMonth retrieves novelties by employee and month with validation
func (uc *noveltyUseCase) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error) {
	uc.logger.LogOperation("GetByEmployeeAndMonth", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
	})

	// Validate employee ID
	if err := uc.validator.ValidateID(employeeID); err != nil {
		uc.logger.LogError("GetByEmployeeAndMonth", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return nil, uc.errorHandler.HandleValidationError("GetByEmployeeAndMonth", err)
	}

	// Validate year and month (reuse validation logic from absence)
	if year < 2000 || year > 2100 {
		err := fmt.Errorf("invalid year: %d. Must be between 2000-2100", year)
		uc.logger.LogError("GetByEmployeeAndMonth", err, map[string]interface{}{
			"year": year,
		})
		return nil, uc.errorHandler.HandleValidationError("GetByEmployeeAndMonth", err)
	}

	if month < 1 || month > 12 {
		err := fmt.Errorf("invalid month: %d. Must be between 1-12", month)
		uc.logger.LogError("GetByEmployeeAndMonth", err, map[string]interface{}{
			"month": month,
		})
		return nil, uc.errorHandler.HandleValidationError("GetByEmployeeAndMonth", err)
	}

	novelties, err := uc.repo.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("GetByEmployeeAndMonth", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return nil, uc.errorHandler.HandleRepositoryError("GetByEmployeeAndMonth", err)
	}

	uc.logger.LogOperation("GetByEmployeeAndMonth", "success", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
		"count":       len(novelties),
	})

	return novelties, nil
}

// ✅ Business-specific operations with enhanced features

// GetByEmployeeAndDateRange retrieves novelties by employee within a custom date range
func (uc *noveltyUseCase) GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Novelty, error) {
	// Start performance timer
	timer := uc.logger.StartTimer("GetByEmployeeAndDateRange", map[string]interface{}{
		"employee_id": employeeID,
		"from":        from.Format("2006-01-02"),
		"to":          to.Format("2006-01-02"),
	})
	defer timer.Stop()

	// Validate employee ID
	if err := uc.validator.ValidateID(employeeID); err != nil {
		uc.logger.LogError("GetByEmployeeAndDateRange", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return nil, uc.errorHandler.HandleValidationError("GetByEmployeeAndDateRange", err)
	}

	// Validate date range
	if from.After(to) {
		err := fmt.Errorf("from date (%s) cannot be after to date (%s)", from.Format("2006-01-02"), to.Format("2006-01-02"))
		uc.logger.LogError("GetByEmployeeAndDateRange", err, map[string]interface{}{
			"from": from.Format("2006-01-02"),
			"to":   to.Format("2006-01-02"),
		})
		return nil, uc.errorHandler.HandleValidationError("GetByEmployeeAndDateRange", err)
	}

	// Execute query
	novelties, err := uc.repo.GetByEmployeeAndDateRange(employeeID, from, to)
	if err != nil {
		uc.logger.LogError("GetByEmployeeAndDateRange", err, map[string]interface{}{
			"employee_id": employeeID,
			"from":        from.Format("2006-01-02"),
			"to":          to.Format("2006-01-02"),
		})
		return nil, uc.errorHandler.HandleRepositoryError("GetByEmployeeAndDateRange", err)
	}

	uc.logger.LogOperation("GetByEmployeeAndDateRange", "success", map[string]interface{}{
		"employee_id": employeeID,
		"from":        from.Format("2006-01-02"),
		"to":          to.Format("2006-01-02"),
		"count":       len(novelties),
	})

	return novelties, nil
}

// GetTotalHoursByEmployeeAndType retrieves total novelty hours by type for an employee in a specific month
func (uc *noveltyUseCase) GetTotalHoursByEmployeeAndType(employeeID, year, month int, noveltyType string) (float64, error) {
	uc.logger.LogOperation("GetTotalHoursByEmployeeAndType", "start", map[string]interface{}{
		"employee_id":  employeeID,
		"year":         year,
		"month":        month,
		"novelty_type": noveltyType,
	})

	// Validate employee ID
	if err := uc.validator.ValidateID(employeeID); err != nil {
		uc.logger.LogError("GetTotalHoursByEmployeeAndType", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return 0, uc.errorHandler.HandleValidationError("GetTotalHoursByEmployeeAndType", err)
	}

	// Validate novelty type
	if err := uc.ValidateNoveltyType(noveltyType); err != nil {
		return 0, err // Error already logged by ValidateNoveltyType
	}

	// Validate year and month
	if year < 2000 || year > 2100 {
		err := fmt.Errorf("invalid year: %d. Must be between 2000-2100", year)
		uc.logger.LogError("GetTotalHoursByEmployeeAndType", err, map[string]interface{}{
			"year": year,
		})
		return 0, uc.errorHandler.HandleValidationError("GetTotalHoursByEmployeeAndType", err)
	}

	if month < 1 || month > 12 {
		err := fmt.Errorf("invalid month: %d. Must be between 1-12", month)
		uc.logger.LogError("GetTotalHoursByEmployeeAndType", err, map[string]interface{}{
			"month": month,
		})
		return 0, uc.errorHandler.HandleValidationError("GetTotalHoursByEmployeeAndType", err)
	}

	// Execute query
	totalHours, err := uc.repo.GetTotalHoursByEmployeeAndType(employeeID, year, month, noveltyType)
	if err != nil {
		uc.logger.LogError("GetTotalHoursByEmployeeAndType", err, map[string]interface{}{
			"employee_id":  employeeID,
			"year":         year,
			"month":        month,
			"novelty_type": noveltyType,
		})
		return 0, uc.errorHandler.HandleRepositoryError("GetTotalHoursByEmployeeAndType", err)
	}

	uc.logger.LogOperation("GetTotalHoursByEmployeeAndType", "success", map[string]interface{}{
		"employee_id":  employeeID,
		"year":         year,
		"month":        month,
		"novelty_type": noveltyType,
		"total_hours":  totalHours,
	})

	return totalHours, nil
}

// GetNoveltyTypesSummary retrieves a comprehensive summary of novelty types for an employee in a month
func (uc *noveltyUseCase) GetNoveltyTypesSummary(employeeID, year, month int) ([]repository.NoveltyTypeSummary, error) {
	uc.logger.LogOperation("GetNoveltyTypesSummary", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
	})

	// Validate employee ID
	if err := uc.validator.ValidateID(employeeID); err != nil {
		uc.logger.LogError("GetNoveltyTypesSummary", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return nil, uc.errorHandler.HandleValidationError("GetNoveltyTypesSummary", err)
	}

	// Validate year and month
	if year < 2000 || year > 2100 {
		err := fmt.Errorf("invalid year: %d. Must be between 2000-2100", year)
		uc.logger.LogError("GetNoveltyTypesSummary", err, map[string]interface{}{
			"year": year,
		})
		return nil, uc.errorHandler.HandleValidationError("GetNoveltyTypesSummary", err)
	}

	if month < 1 || month > 12 {
		err := fmt.Errorf("invalid month: %d. Must be between 1-12", month)
		uc.logger.LogError("GetNoveltyTypesSummary", err, map[string]interface{}{
			"month": month,
		})
		return nil, uc.errorHandler.HandleValidationError("GetNoveltyTypesSummary", err)
	}

	// Execute query
	summary, err := uc.repo.GetNoveltyTypesSummary(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("GetNoveltyTypesSummary", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return nil, uc.errorHandler.HandleRepositoryError("GetNoveltyTypesSummary", err)
	}

	uc.logger.LogOperation("GetNoveltyTypesSummary", "success", map[string]interface{}{
		"employee_id":   employeeID,
		"year":          year,
		"month":         month,
		"summary_count": len(summary),
	})

	return summary, nil
}

// ValidateNoveltyData validates novelty data according to business rules
func (uc *noveltyUseCase) ValidateNoveltyData(novelty *domain.Novelty) error {
	// Basic entity validation
	if err := uc.validator.ValidateEntity(novelty); err != nil {
		uc.logger.LogValidation("ValidateNoveltyData", "entity", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate employee ID
	if err := uc.validator.ValidateNumber(novelty.EmployeeID, "employee_id", "positive"); err != nil {
		uc.logger.LogValidation("ValidateNoveltyData", "employee_id", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate hours
	if err := uc.validator.ValidateNumber(novelty.Hours, "hours", "positive"); err != nil {
		uc.logger.LogValidation("ValidateNoveltyData", "hours", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule: Maximum 24 hours per novelty entry
	if novelty.Hours > 24.0 {
		err := fmt.Errorf("novelty hours cannot exceed 24 hours per entry, got: %.2f", novelty.Hours)
		uc.logger.LogValidation("ValidateNoveltyData", "max_hours", "failed", map[string]interface{}{
			"error": err.Error(),
			"hours": novelty.Hours,
		})
		return err
	}

	// Business rule: Minimum 0.25 hours (15 minutes)
	if novelty.Hours < 0.25 {
		err := fmt.Errorf("novelty hours must be at least 0.25 hours (15 minutes), got: %.2f", novelty.Hours)
		uc.logger.LogValidation("ValidateNoveltyData", "min_hours", "failed", map[string]interface{}{
			"error": err.Error(),
			"hours": novelty.Hours,
		})
		return err
	}

	// Validate comment (optional but if provided, must be meaningful)
	if novelty.Comment != "" {
		if err := uc.validator.ValidateString(novelty.Comment, "comment", "min:3", "max:255"); err != nil {
			uc.logger.LogValidation("ValidateNoveltyData", "comment", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	uc.logger.LogValidation("ValidateNoveltyData", "all_fields", "passed", nil)
	return nil
}

// ValidateNoveltyType validates novelty type according to business rules
func (uc *noveltyUseCase) ValidateNoveltyType(noveltyType string) error {
	uc.logger.LogOperation("ValidateNoveltyType", "start", map[string]interface{}{
		"type": noveltyType,
	})

	// Normalize type
	normalizedType := strings.TrimSpace(strings.ToLower(noveltyType))

	// Business rule: Only positive and negative types allowed
	validTypes := []string{"positive", "negative"}
	isValidType := false
	for _, validType := range validTypes {
		if normalizedType == validType {
			isValidType = true
			break
		}
	}

	if !isValidType {
		err := fmt.Errorf("invalid novelty type: %s. Valid types are: %s", noveltyType, strings.Join(validTypes, ", "))
		uc.logger.LogValidation("ValidateNoveltyType", "type_validation", "failed", map[string]interface{}{
			"error":       err.Error(),
			"provided":    noveltyType,
			"normalized":  normalizedType,
			"valid_types": validTypes,
		})
		return err
	}

	uc.logger.LogValidation("ValidateNoveltyType", "type_validation", "passed", map[string]interface{}{
		"type": normalizedType,
	})

	return nil
}

// ValidateNoveltyDate validates novelty date according to business rules
func (uc *noveltyUseCase) ValidateNoveltyDate(date time.Time) error {
	uc.logger.LogOperation("ValidateNoveltyDate", "start", map[string]interface{}{
		"date": date.Format("2006-01-02"),
	})

	// Business rule: Cannot register novelty for future dates (except today)
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrowStart := todayStart.AddDate(0, 0, 1)

	if date.After(tomorrowStart) || date.Equal(tomorrowStart) {
		err := fmt.Errorf("cannot register novelty for future date: %s", date.Format("2006-01-02"))
		uc.logger.LogValidation("ValidateNoveltyDate", "future_date", "failed", map[string]interface{}{
			"error": err.Error(),
			"date":  date.Format("2006-01-02"),
		})
		return err
	}

	// Business rule: Cannot register novelty for dates older than 30 days (more flexible than absences)
	thirtyDaysAgo := now.AddDate(0, 0, -30)
	if date.Before(thirtyDaysAgo) {
		err := fmt.Errorf("cannot register novelty for dates older than 30 days: %s", date.Format("2006-01-02"))
		uc.logger.LogValidation("ValidateNoveltyDate", "too_old_date", "failed", map[string]interface{}{
			"error": err.Error(),
			"date":  date.Format("2006-01-02"),
		})
		return err
	}

	uc.logger.LogValidation("ValidateNoveltyDate", "date_rules", "passed", map[string]interface{}{
		"date": date.Format("2006-01-02"),
	})

	return nil
}
