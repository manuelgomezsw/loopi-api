package usecase

import (
	"fmt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"time"
)

type AbsenceUseCase interface {
	// Standard CRUD operations
	Create(absence domain.Absence) error
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error)

	// Business-specific operations
	GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Absence, error)
	GetTotalHoursByEmployee(employeeID, year, month int) (float64, error)
	ValidateAbsenceData(absence *domain.Absence) error
	ValidateAbsenceDate(date time.Time) error
	ValidateEmployeeAbsenceLimit(employeeID int, year, month int, additionalHours float64) error
}

type absenceUseCase struct {
	repo         repository.AbsenceRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewAbsenceUseCase(repo repository.AbsenceRepository) AbsenceUseCase {
	return &absenceUseCase{
		repo:         repo,
		errorHandler: base.NewErrorHandler("Absence"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Absence"),
	}
}

// ✅ Enhanced CRUD operations with logging, validation, and error handling

// Create creates a new absence with validation and business rules
func (uc *absenceUseCase) Create(absence domain.Absence) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"employee_id": absence.EmployeeID,
		"date":        absence.Date.Format("2006-01-02"),
		"hours":       absence.Hours,
	})

	// Validate business rules
	if err := uc.ValidateAbsenceData(&absence); err != nil {
		return err
	}

	// Validate absence date
	if err := uc.ValidateAbsenceDate(absence.Date); err != nil {
		return err
	}

	// Validate employee absence limit
	if err := uc.ValidateEmployeeAbsenceLimit(absence.EmployeeID, absence.Date.Year(), int(absence.Date.Month()), absence.Hours); err != nil {
		return err
	}

	// Set timestamps
	absence.CreatedAt = time.Now()
	absence.UpdatedAt = time.Now()

	// Execute creation
	if err := uc.repo.Create(&absence); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"employee_id": absence.EmployeeID,
			"date":        absence.Date.Format("2006-01-02"),
		})
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"absence_id":  absence.ID,
		"employee_id": absence.EmployeeID,
		"hours":       absence.Hours,
	})

	return nil
}

// GetByEmployeeAndMonth retrieves absences by employee and month with validation
func (uc *absenceUseCase) GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Absence, error) {
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

	// Validate year and month
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

	absences, err := uc.repo.GetByEmployeeAndMonth(employeeID, year, month)
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
		"count":       len(absences),
	})

	return absences, nil
}

// ✅ Business-specific operations with enhanced features

// GetByEmployeeAndDateRange retrieves absences by employee within a custom date range
func (uc *absenceUseCase) GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Absence, error) {
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
	absences, err := uc.repo.GetByEmployeeAndDateRange(employeeID, from, to)
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
		"count":       len(absences),
	})

	return absences, nil
}

// GetTotalHoursByEmployee retrieves total absence hours for an employee in a specific month
func (uc *absenceUseCase) GetTotalHoursByEmployee(employeeID, year, month int) (float64, error) {
	uc.logger.LogOperation("GetTotalHoursByEmployee", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
	})

	// Validate employee ID
	if err := uc.validator.ValidateID(employeeID); err != nil {
		uc.logger.LogError("GetTotalHoursByEmployee", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return 0, uc.errorHandler.HandleValidationError("GetTotalHoursByEmployee", err)
	}

	// Validate year and month (reuse validation logic)
	if year < 2000 || year > 2100 {
		err := fmt.Errorf("invalid year: %d. Must be between 2000-2100", year)
		uc.logger.LogError("GetTotalHoursByEmployee", err, map[string]interface{}{
			"year": year,
		})
		return 0, uc.errorHandler.HandleValidationError("GetTotalHoursByEmployee", err)
	}

	if month < 1 || month > 12 {
		err := fmt.Errorf("invalid month: %d. Must be between 1-12", month)
		uc.logger.LogError("GetTotalHoursByEmployee", err, map[string]interface{}{
			"month": month,
		})
		return 0, uc.errorHandler.HandleValidationError("GetTotalHoursByEmployee", err)
	}

	// Execute query
	totalHours, err := uc.repo.GetTotalHoursByEmployee(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("GetTotalHoursByEmployee", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return 0, uc.errorHandler.HandleRepositoryError("GetTotalHoursByEmployee", err)
	}

	uc.logger.LogOperation("GetTotalHoursByEmployee", "success", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
		"total_hours": totalHours,
	})

	return totalHours, nil
}

// ValidateAbsenceData validates absence data according to business rules
func (uc *absenceUseCase) ValidateAbsenceData(absence *domain.Absence) error {
	// Basic entity validation
	if err := uc.validator.ValidateEntity(absence); err != nil {
		uc.logger.LogValidation("ValidateAbsenceData", "entity", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate employee ID
	if err := uc.validator.ValidateNumber(absence.EmployeeID, "employee_id", "positive"); err != nil {
		uc.logger.LogValidation("ValidateAbsenceData", "employee_id", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate hours
	if err := uc.validator.ValidateNumber(absence.Hours, "hours", "positive"); err != nil {
		uc.logger.LogValidation("ValidateAbsenceData", "hours", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule: Maximum 24 hours per absence entry
	if absence.Hours > 24.0 {
		err := fmt.Errorf("absence hours cannot exceed 24 hours per entry, got: %.2f", absence.Hours)
		uc.logger.LogValidation("ValidateAbsenceData", "max_hours", "failed", map[string]interface{}{
			"error": err.Error(),
			"hours": absence.Hours,
		})
		return err
	}

	// Business rule: Minimum 0.25 hours (15 minutes)
	if absence.Hours < 0.25 {
		err := fmt.Errorf("absence hours must be at least 0.25 hours (15 minutes), got: %.2f", absence.Hours)
		uc.logger.LogValidation("ValidateAbsenceData", "min_hours", "failed", map[string]interface{}{
			"error": err.Error(),
			"hours": absence.Hours,
		})
		return err
	}

	// Validate reason (optional but if provided, must be meaningful)
	if absence.Reason != "" {
		if err := uc.validator.ValidateString(absence.Reason, "reason", "min:3", "max:255"); err != nil {
			uc.logger.LogValidation("ValidateAbsenceData", "reason", "failed", map[string]interface{}{
				"error": err.Error(),
			})
			return err
		}
	}

	uc.logger.LogValidation("ValidateAbsenceData", "all_fields", "passed", nil)
	return nil
}

// ValidateAbsenceDate validates absence date according to business rules
func (uc *absenceUseCase) ValidateAbsenceDate(date time.Time) error {
	uc.logger.LogOperation("ValidateAbsenceDate", "start", map[string]interface{}{
		"date": date.Format("2006-01-02"),
	})

	// Business rule: Cannot register absence for future dates
	now := time.Now()
	if date.After(now) {
		err := fmt.Errorf("cannot register absence for future date: %s", date.Format("2006-01-02"))
		uc.logger.LogValidation("ValidateAbsenceDate", "future_date", "failed", map[string]interface{}{
			"error": err.Error(),
			"date":  date.Format("2006-01-02"),
		})
		return err
	}

	// Business rule: Cannot register absence for dates older than 90 days
	ninetyDaysAgo := now.AddDate(0, 0, -90)
	if date.Before(ninetyDaysAgo) {
		err := fmt.Errorf("cannot register absence for dates older than 90 days: %s", date.Format("2006-01-02"))
		uc.logger.LogValidation("ValidateAbsenceDate", "too_old_date", "failed", map[string]interface{}{
			"error": err.Error(),
			"date":  date.Format("2006-01-02"),
		})
		return err
	}

	uc.logger.LogValidation("ValidateAbsenceDate", "date_rules", "passed", map[string]interface{}{
		"date": date.Format("2006-01-02"),
	})

	return nil
}

// ValidateEmployeeAbsenceLimit validates that an employee doesn't exceed absence limits
func (uc *absenceUseCase) ValidateEmployeeAbsenceLimit(employeeID int, year, month int, additionalHours float64) error {
	uc.logger.LogOperation("ValidateEmployeeAbsenceLimit", "start", map[string]interface{}{
		"employee_id":      employeeID,
		"year":             year,
		"month":            month,
		"additional_hours": additionalHours,
	})

	// Get current total hours for the employee in the month
	currentTotalHours, err := uc.GetTotalHoursByEmployee(employeeID, year, month)
	if err != nil {
		return err // Error already logged by GetTotalHoursByEmployee
	}

	// Calculate total hours after adding the new absence
	newTotalHours := currentTotalHours + additionalHours

	// Business rule: Maximum 40 absence hours per month per employee
	maxMonthlyHours := 40.0
	if newTotalHours > maxMonthlyHours {
		err := fmt.Errorf("employee %d would exceed monthly absence limit of %.1f hours. Current: %.2f, Adding: %.2f, Total would be: %.2f",
			employeeID, maxMonthlyHours, currentTotalHours, additionalHours, newTotalHours)
		uc.logger.LogValidation("ValidateEmployeeAbsenceLimit", "monthly_limit", "failed", map[string]interface{}{
			"error":             err.Error(),
			"employee_id":       employeeID,
			"current_hours":     currentTotalHours,
			"additional_hours":  additionalHours,
			"would_be_total":    newTotalHours,
			"max_monthly_hours": maxMonthlyHours,
		})
		return err
	}

	uc.logger.LogValidation("ValidateEmployeeAbsenceLimit", "monthly_limit", "passed", map[string]interface{}{
		"employee_id":       employeeID,
		"current_hours":     currentTotalHours,
		"additional_hours":  additionalHours,
		"total_after":       newTotalHours,
		"max_monthly_hours": maxMonthlyHours,
	})

	return nil
}
