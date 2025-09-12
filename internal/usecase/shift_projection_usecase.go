package usecase

import (
	"fmt"
	"loopi-api/internal/calendar"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"loopi-api/internal/usecase/dto"
	"loopi-api/internal/usecase/utils"
)

type ShiftProjectionUseCase interface {
	// Standard operations
	PreviewHours(req dto.ShiftProjectionRequest) (domain.ExtraHourSummary, error)

	// Business-specific operations
	ValidateProjectionRequest(req *dto.ShiftProjectionRequest) error
	GetShiftProjectionSummary(shiftID, year, month int) (domain.ExtraHourSummary, error)
	CalculateProjectedDays(shiftID, year, month int) (int, error)
	ValidateShiftExists(shiftID int) (*domain.Shift, error)
	GetWorkConfig() (domain.WorkConfig, error)
}

type shiftProjectionUseCase struct {
	shiftRepo      repository.ShiftRepository
	workConfigRepo repository.WorkConfigRepository
	errorHandler   *base.ErrorHandler
	validator      *base.Validator
	logger         *base.Logger
}

func NewShiftProjectionUseCase(
	shiftRepo repository.ShiftRepository,
	workConfigRepo repository.WorkConfigRepository,
) ShiftProjectionUseCase {
	return &shiftProjectionUseCase{
		shiftRepo:      shiftRepo,
		workConfigRepo: workConfigRepo,
		errorHandler:   base.NewErrorHandler("ShiftProjection"),
		validator:      base.NewValidator(),
		logger:         base.NewLogger("ShiftProjection"),
	}
}

// ✅ Enhanced operations with logging, validation, and error handling

// PreviewHours generates projected hours summary based on shift and period
func (uc *shiftProjectionUseCase) PreviewHours(req dto.ShiftProjectionRequest) (domain.ExtraHourSummary, error) {
	uc.logger.LogOperation("PreviewHours", "start", map[string]interface{}{
		"shift_id": req.ShiftID,
		"year":     req.Year,
		"month":    req.Month,
	})

	// Validate request
	if err := uc.ValidateProjectionRequest(&req); err != nil {
		return domain.ExtraHourSummary{}, err
	}

	// Validate shift exists
	shift, err := uc.ValidateShiftExists(req.ShiftID)
	if err != nil {
		return domain.ExtraHourSummary{}, err
	}

	// Get work configuration
	workConfig, err := uc.GetWorkConfig()
	if err != nil {
		return domain.ExtraHourSummary{}, err
	}

	// Build holiday map for the period
	holidayMap := utils.HolidaysToMap(
		calendar.GetColombianHolidaysByMonthCached(req.Year, req.Month),
	)

	uc.logger.LogOperation("PreviewHours", "data_prepared", map[string]interface{}{
		"shift_id":   req.ShiftID,
		"shift_name": shift.Name,
		"holidays":   len(holidayMap),
	})

	// Build calendar days
	calendarDays := utils.BuildCalendarDays(req.Year, req.Month, holidayMap)

	// Apply shift to calendar and calculate projection
	projected := utils.ApplyShiftToCalendar(calendarDays, *shift, workConfig)
	summary := utils.SummarizeProjection(projected)

	// Set period information
	summary.Period = domain.Period{
		Year:  req.Year,
		Month: req.Month,
	}

	uc.logger.LogOperation("PreviewHours", "success", map[string]interface{}{
		"shift_id":       req.ShiftID,
		"shift_name":     shift.Name,
		"year":           req.Year,
		"month":          req.Month,
		"projected_days": len(projected),
		"summary":        summary,
	})

	return summary, nil
}

// ✅ Business-specific operations with enhanced validation and logging

// ValidateProjectionRequest validates a shift projection request
func (uc *shiftProjectionUseCase) ValidateProjectionRequest(req *dto.ShiftProjectionRequest) error {
	uc.logger.LogOperation("ValidateProjectionRequest", "start", map[string]interface{}{
		"shift_id": req.ShiftID,
		"year":     req.Year,
		"month":    req.Month,
	})

	// Validate shift ID
	if err := uc.validator.ValidateID(req.ShiftID); err != nil {
		uc.logger.LogValidation("ValidateProjectionRequest", "shift_id", "failed", map[string]interface{}{
			"error":    err.Error(),
			"shift_id": req.ShiftID,
		})
		return uc.errorHandler.HandleValidationError("ValidateProjectionRequest", err)
	}

	// Validate year
	if req.Year < 2000 || req.Year > 2030 {
		err := fmt.Errorf("year must be between 2000 and 2030, got: %d", req.Year)
		uc.logger.LogValidation("ValidateProjectionRequest", "year_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"year":  req.Year,
		})
		return uc.errorHandler.HandleValidationError("ValidateProjectionRequest", err)
	}

	// Validate month
	if req.Month < 1 || req.Month > 12 {
		err := fmt.Errorf("month must be between 1 and 12, got: %d", req.Month)
		uc.logger.LogValidation("ValidateProjectionRequest", "month_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"month": req.Month,
		})
		return uc.errorHandler.HandleValidationError("ValidateProjectionRequest", err)
	}

	uc.logger.LogValidation("ValidateProjectionRequest", "all_fields", "passed", map[string]interface{}{
		"shift_id": req.ShiftID,
		"year":     req.Year,
		"month":    req.Month,
	})

	return nil
}

// ValidateShiftExists validates that a shift exists and returns it
func (uc *shiftProjectionUseCase) ValidateShiftExists(shiftID int) (*domain.Shift, error) {
	uc.logger.LogOperation("ValidateShiftExists", "start", map[string]interface{}{
		"shift_id": shiftID,
	})

	// Get shift from repository
	shift, err := uc.shiftRepo.GetByID(shiftID)
	if err != nil {
		uc.logger.LogError("ValidateShiftExists", err, map[string]interface{}{
			"shift_id": shiftID,
		})
		return nil, uc.errorHandler.HandleRepositoryError("ValidateShiftExists", err)
	}

	if shift == nil {
		err := fmt.Errorf("shift not found with ID: %d", shiftID)
		uc.logger.LogValidation("ValidateShiftExists", "shift_exists", "failed", map[string]interface{}{
			"error":    err.Error(),
			"shift_id": shiftID,
		})
		return nil, uc.errorHandler.HandleNotFound("ValidateShiftExists", fmt.Sprintf("shift not found with ID: %d", shiftID))
	}

	uc.logger.LogValidation("ValidateShiftExists", "shift_exists", "passed", map[string]interface{}{
		"shift_id":   shiftID,
		"shift_name": shift.Name,
	})

	return shift, nil
}

// GetWorkConfig retrieves the active work configuration
func (uc *shiftProjectionUseCase) GetWorkConfig() (domain.WorkConfig, error) {
	uc.logger.LogOperation("GetWorkConfig", "start", nil)

	// Get active work configuration
	workConfig := uc.workConfigRepo.GetActiveConfig()

	// Business rule: Work config should have valid diurnal periods
	if workConfig.DiurnalStart == "" || workConfig.DiurnalEnd == "" {
		err := fmt.Errorf("invalid diurnal periods in work config: start=%s, end=%s", workConfig.DiurnalStart, workConfig.DiurnalEnd)
		uc.logger.LogValidation("GetWorkConfig", "diurnal_periods", "failed", map[string]interface{}{
			"error":         err.Error(),
			"diurnal_start": workConfig.DiurnalStart,
			"diurnal_end":   workConfig.DiurnalEnd,
		})
		return workConfig, uc.errorHandler.HandleBusinessRuleViolation("GetWorkConfig", "diurnal_periods_validation", err.Error())
	}

	// Business rule: Work config should be active
	if !workConfig.IsActive {
		err := fmt.Errorf("work configuration is not active")
		uc.logger.LogValidation("GetWorkConfig", "is_active", "failed", map[string]interface{}{
			"error":     err.Error(),
			"is_active": workConfig.IsActive,
		})
		return workConfig, uc.errorHandler.HandleBusinessRuleViolation("GetWorkConfig", "active_config_validation", err.Error())
	}

	uc.logger.LogOperation("GetWorkConfig", "success", map[string]interface{}{
		"diurnal_start": workConfig.DiurnalStart,
		"diurnal_end":   workConfig.DiurnalEnd,
		"is_active":     workConfig.IsActive,
	})

	return workConfig, nil
}

// GetShiftProjectionSummary provides a comprehensive projection summary
func (uc *shiftProjectionUseCase) GetShiftProjectionSummary(shiftID, year, month int) (domain.ExtraHourSummary, error) {
	uc.logger.LogOperation("GetShiftProjectionSummary", "start", map[string]interface{}{
		"shift_id": shiftID,
		"year":     year,
		"month":    month,
	})

	// Create request object
	req := dto.ShiftProjectionRequest{
		ShiftID: shiftID,
		Year:    year,
		Month:   month,
	}

	// Use existing PreviewHours method
	summary, err := uc.PreviewHours(req)
	if err != nil {
		return domain.ExtraHourSummary{}, err // Error already logged by PreviewHours
	}

	uc.logger.LogOperation("GetShiftProjectionSummary", "success", map[string]interface{}{
		"shift_id": shiftID,
		"year":     year,
		"month":    month,
		"summary":  summary,
	})

	return summary, nil
}

// CalculateProjectedDays calculates the number of projected working days
func (uc *shiftProjectionUseCase) CalculateProjectedDays(shiftID, year, month int) (int, error) {
	uc.logger.LogOperation("CalculateProjectedDays", "start", map[string]interface{}{
		"shift_id": shiftID,
		"year":     year,
		"month":    month,
	})

	// Validate inputs through projection request
	req := dto.ShiftProjectionRequest{
		ShiftID: shiftID,
		Year:    year,
		Month:   month,
	}

	if err := uc.ValidateProjectionRequest(&req); err != nil {
		return 0, err
	}

	// Validate shift exists
	shift, err := uc.ValidateShiftExists(shiftID)
	if err != nil {
		return 0, err
	}

	// Get work configuration
	workConfig, err := uc.GetWorkConfig()
	if err != nil {
		return 0, err
	}

	// Build calendar for the period
	holidayMap := utils.HolidaysToMap(
		calendar.GetColombianHolidaysByMonthCached(year, month),
	)
	calendarDays := utils.BuildCalendarDays(year, month, holidayMap)

	// Apply shift to calendar to get projected days
	projected := utils.ApplyShiftToCalendar(calendarDays, *shift, workConfig)
	projectedDays := len(projected)

	uc.logger.LogOperation("CalculateProjectedDays", "success", map[string]interface{}{
		"shift_id":       shiftID,
		"shift_name":     shift.Name,
		"year":           year,
		"month":          month,
		"projected_days": projectedDays,
		"calendar_days":  len(calendarDays),
		"holidays":       len(holidayMap),
	})

	return projectedDays, nil
}
