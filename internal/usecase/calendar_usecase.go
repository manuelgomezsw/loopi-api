package usecase

import (
	"fmt"
	"loopi-api/internal/calendar"
	"loopi-api/internal/usecase/base"
	"time"
)

// CalendarSummary represents comprehensive calendar data for a month
type CalendarSummary struct {
	Year         int         `json:"year"`
	Month        int         `json:"month"`
	Holidays     []time.Time `json:"holidays"`
	OrdinaryDays int         `json:"ordinary_days"`
	Sundays      int         `json:"sundays"`
	WorkingDays  int         `json:"working_days"`
}

type CalendarUseCase interface {
	// Standard operations
	GetHolidays(year int) ([]time.Time, error)
	GetHolidaysByMonth(year int, month int) ([]time.Time, error)
	CountOrdinaryDays(year int, month int) (int, error)
	CountSundays(year int, month int) (int, error)

	// Business-specific operations
	GetMonthSummary(year int, month int) (*CalendarSummary, error)
	GetWorkingDays(year int, month int) (int, error)
	ValidateYear(year int) error
	ValidateMonth(month int) error
	ClearCache() error
}

type calendarUseCase struct {
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewCalendarUseCase() CalendarUseCase {
	return &calendarUseCase{
		errorHandler: base.NewErrorHandler("Calendar"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Calendar"),
	}
}

// ✅ Enhanced operations with logging, validation, and error handling

// GetHolidays retrieves all holidays for a specific year with validation
func (uc *calendarUseCase) GetHolidays(year int) ([]time.Time, error) {
	uc.logger.LogOperation("GetHolidays", "start", map[string]interface{}{
		"year": year,
	})

	// Validate year
	if err := uc.ValidateYear(year); err != nil {
		return nil, err
	}

	// Get holidays from calendar package
	holidays := calendar.GetColombianHolidaysCached(year)

	uc.logger.LogOperation("GetHolidays", "success", map[string]interface{}{
		"year":  year,
		"count": len(holidays),
	})

	return holidays, nil
}

// GetHolidaysByMonth retrieves holidays for a specific year and month with validation
func (uc *calendarUseCase) GetHolidaysByMonth(year int, month int) ([]time.Time, error) {
	uc.logger.LogOperation("GetHolidaysByMonth", "start", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	// Validate year and month
	if err := uc.ValidateYear(year); err != nil {
		return nil, err
	}

	if err := uc.ValidateMonth(month); err != nil {
		return nil, err
	}

	// Get holidays from calendar package
	holidays := calendar.GetColombianHolidaysByMonthCached(year, month)

	uc.logger.LogOperation("GetHolidaysByMonth", "success", map[string]interface{}{
		"year":  year,
		"month": month,
		"count": len(holidays),
	})

	return holidays, nil
}

// CountOrdinaryDays counts ordinary days in a month with validation
func (uc *calendarUseCase) CountOrdinaryDays(year int, month int) (int, error) {
	uc.logger.LogOperation("CountOrdinaryDays", "start", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	// Validate year and month
	if err := uc.ValidateYear(year); err != nil {
		return 0, err
	}

	if err := uc.ValidateMonth(month); err != nil {
		return 0, err
	}

	// Get ordinary days from calendar package
	ord, _ := calendar.GetMonthSummaryCached(year, month)

	uc.logger.LogOperation("CountOrdinaryDays", "success", map[string]interface{}{
		"year":          year,
		"month":         month,
		"ordinary_days": ord,
	})

	return ord, nil
}

// CountSundays counts Sundays in a month with validation
func (uc *calendarUseCase) CountSundays(year int, month int) (int, error) {
	uc.logger.LogOperation("CountSundays", "start", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	// Validate year and month
	if err := uc.ValidateYear(year); err != nil {
		return 0, err
	}

	if err := uc.ValidateMonth(month); err != nil {
		return 0, err
	}

	// Get Sundays from calendar package
	_, sun := calendar.GetMonthSummaryCached(year, month)

	uc.logger.LogOperation("CountSundays", "success", map[string]interface{}{
		"year":    year,
		"month":   month,
		"sundays": sun,
	})

	return sun, nil
}

// ✅ Business-specific operations with enhanced features

// GetMonthSummary retrieves comprehensive calendar summary for a month
func (uc *calendarUseCase) GetMonthSummary(year int, month int) (*CalendarSummary, error) {
	// Start performance timer
	timer := uc.logger.StartTimer("GetMonthSummary", map[string]interface{}{
		"year":  year,
		"month": month,
	})
	defer timer.Stop()

	// Validate year and month
	if err := uc.ValidateYear(year); err != nil {
		return nil, err
	}

	if err := uc.ValidateMonth(month); err != nil {
		return nil, err
	}

	// Get holidays
	holidays, err := uc.GetHolidaysByMonth(year, month)
	if err != nil {
		return nil, err
	}

	// Get ordinary days and sundays
	ordinaryDays, err := uc.CountOrdinaryDays(year, month)
	if err != nil {
		return nil, err
	}

	sundays, err := uc.CountSundays(year, month)
	if err != nil {
		return nil, err
	}

	// Calculate working days (business rule: ordinary days - holidays that don't fall on Sunday)
	workingDays, err := uc.GetWorkingDays(year, month)
	if err != nil {
		return nil, err
	}

	summary := &CalendarSummary{
		Year:         year,
		Month:        month,
		Holidays:     holidays,
		OrdinaryDays: ordinaryDays,
		Sundays:      sundays,
		WorkingDays:  workingDays,
	}

	uc.logger.LogOperation("GetMonthSummary", "success", map[string]interface{}{
		"year":          year,
		"month":         month,
		"holidays":      len(holidays),
		"ordinary_days": ordinaryDays,
		"sundays":       sundays,
		"working_days":  workingDays,
	})

	return summary, nil
}

// GetWorkingDays calculates actual working days in a month (excludes holidays that don't fall on Sunday)
func (uc *calendarUseCase) GetWorkingDays(year int, month int) (int, error) {
	uc.logger.LogOperation("GetWorkingDays", "start", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	// Validate year and month
	if err := uc.ValidateYear(year); err != nil {
		return 0, err
	}

	if err := uc.ValidateMonth(month); err != nil {
		return 0, err
	}

	// Get ordinary days
	ordinaryDays, err := uc.CountOrdinaryDays(year, month)
	if err != nil {
		return 0, err
	}

	// Get holidays
	holidays, err := uc.GetHolidaysByMonth(year, month)
	if err != nil {
		return 0, err
	}

	// Business rule: Count holidays that don't fall on Sunday (those reduce working days)
	holidaysNotOnSunday := 0
	for _, holiday := range holidays {
		if holiday.Weekday() != time.Sunday {
			holidaysNotOnSunday++
		}
	}

	// Calculate working days
	workingDays := ordinaryDays - holidaysNotOnSunday
	if workingDays < 0 {
		workingDays = 0
	}

	uc.logger.LogBusinessRule("GetWorkingDays", "exclude_non_sunday_holidays", "applied", map[string]interface{}{
		"year":                   year,
		"month":                  month,
		"ordinary_days":          ordinaryDays,
		"total_holidays":         len(holidays),
		"holidays_not_on_sunday": holidaysNotOnSunday,
		"final_working_days":     workingDays,
	})

	uc.logger.LogOperation("GetWorkingDays", "success", map[string]interface{}{
		"year":         year,
		"month":        month,
		"working_days": workingDays,
	})

	return workingDays, nil
}

// ClearCache clears the calendar cache
func (uc *calendarUseCase) ClearCache() error {
	uc.logger.LogOperation("ClearCache", "start", nil)

	// Clear cache using calendar package function
	calendar.ClearCalendarCache()

	uc.logger.LogOperation("ClearCache", "success", nil)
	return nil
}

// ValidateYear validates year according to business rules
func (uc *calendarUseCase) ValidateYear(year int) error {
	uc.logger.LogOperation("ValidateYear", "start", map[string]interface{}{
		"year": year,
	})

	// Business rule: Year must be reasonable (between 2000 and 2100)
	if year < 2000 || year > 2100 {
		err := fmt.Errorf("invalid year: %d. Must be between 2000-2100", year)
		uc.logger.LogValidation("ValidateYear", "year_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"year":  year,
		})
		return uc.errorHandler.HandleValidationError("ValidateYear", err)
	}

	uc.logger.LogValidation("ValidateYear", "year_range", "passed", map[string]interface{}{
		"year": year,
	})

	return nil
}

// ValidateMonth validates month according to business rules
func (uc *calendarUseCase) ValidateMonth(month int) error {
	uc.logger.LogOperation("ValidateMonth", "start", map[string]interface{}{
		"month": month,
	})

	// Business rule: Month must be between 1 and 12
	if month < 1 || month > 12 {
		err := fmt.Errorf("invalid month: %d. Must be between 1-12", month)
		uc.logger.LogValidation("ValidateMonth", "month_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"month": month,
		})
		return uc.errorHandler.HandleValidationError("ValidateMonth", err)
	}

	uc.logger.LogValidation("ValidateMonth", "month_range", "passed", map[string]interface{}{
		"month": month,
	})

	return nil
}
