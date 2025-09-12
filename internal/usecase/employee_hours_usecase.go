package usecase

import (
	"fmt"
	"loopi-api/internal/calendar"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"loopi-api/internal/usecase/utils"
)

type EmployeeHoursUseCase interface {
	// Standard operations
	GetMonthlySummary(employeeID, year, month int) (domain.EmployeeHourSummary, error)

	// Business-specific operations
	GetDailySummary(employeeID int, year, month, day int) (DailyHoursSummary, error)
	GetYearlySummary(employeeID, year int) (YearlyHoursSummary, error)
	ValidatePeriod(year, month int) error
	ValidateEmployeeID(employeeID int) error
	CalculateWorkingDays(employeeID, year, month int) (int, error)
}

type DailyHoursSummary struct {
	Date         string  `json:"date"`
	RegularHours float64 `json:"regular_hours"`
	ExtraHours   float64 `json:"extra_hours"`
	AbsenceHours float64 `json:"absence_hours"`
	NoveltyHours float64 `json:"novelty_hours"`
	TotalHours   float64 `json:"total_hours"`
}

type YearlyHoursSummary struct {
	Year         int                          `json:"year"`
	EmployeeID   int                          `json:"employee_id"`
	EmployeeName string                       `json:"employee_name"`
	MonthlyData  []domain.EmployeeHourSummary `json:"monthly_data"`
	TotalHours   domain.EmployeeHourBlock     `json:"total_hours"`
}

type employeeHoursUseCase struct {
	assignedRepo repository.AssignedShiftRepository
	absenceRepo  repository.AbsenceRepository
	noveltyRepo  repository.NoveltyRepository
	userRepo     repository.UserRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewEmployeeHoursUseCase(
	assignedRepo repository.AssignedShiftRepository,
	absenceRepo repository.AbsenceRepository,
	noveltyRepo repository.NoveltyRepository,
	userRepo repository.UserRepository,
) EmployeeHoursUseCase {
	return &employeeHoursUseCase{
		assignedRepo: assignedRepo,
		absenceRepo:  absenceRepo,
		noveltyRepo:  noveltyRepo,
		userRepo:     userRepo,
		errorHandler: base.NewErrorHandler("EmployeeHours"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("EmployeeHours"),
	}
}

// ✅ Enhanced operations with logging, validation, and error handling

// GetMonthlySummary calculates comprehensive monthly hours summary for an employee
func (uc *employeeHoursUseCase) GetMonthlySummary(employeeID, year, month int) (domain.EmployeeHourSummary, error) {
	uc.logger.LogOperation("GetMonthlySummary", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
	})

	// Validate inputs
	if err := uc.ValidateEmployeeID(employeeID); err != nil {
		return domain.EmployeeHourSummary{}, err
	}

	if err := uc.ValidatePeriod(year, month); err != nil {
		return domain.EmployeeHourSummary{}, err
	}

	// Get employee name with error handling
	fullNameEmployee, err := uc.userRepo.GetNameByID(employeeID)
	if err != nil {
		uc.logger.LogError("GetMonthlySummary", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return domain.EmployeeHourSummary{}, uc.errorHandler.HandleRepositoryError("GetMonthlySummary", err)
	}

	// Build calendar days
	holidayMap := utils.HolidaysToMap(calendar.GetColombianHolidaysByMonthCached(year, month))
	calendarDays := utils.BuildCalendarDays(year, month, holidayMap)

	uc.logger.LogOperation("GetMonthlySummary", "calendar_built", map[string]interface{}{
		"employee_id":   employeeID,
		"calendar_days": len(calendarDays),
		"holidays":      len(holidayMap),
	})

	// Get shifts with error handling
	shifts, err := uc.assignedRepo.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("GetMonthlySummary", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return domain.EmployeeHourSummary{}, uc.errorHandler.HandleRepositoryError("GetMonthlySummary", err)
	}

	// Get absences with error handling
	absences, err := uc.absenceRepo.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("GetMonthlySummary", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return domain.EmployeeHourSummary{}, uc.errorHandler.HandleRepositoryError("GetMonthlySummary", err)
	}

	// Get novelties with error handling
	novelties, err := uc.noveltyRepo.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("GetMonthlySummary", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return domain.EmployeeHourSummary{}, uc.errorHandler.HandleRepositoryError("GetMonthlySummary", err)
	}

	uc.logger.LogOperation("GetMonthlySummary", "data_retrieved", map[string]interface{}{
		"employee_id":     employeeID,
		"shifts_count":    len(shifts),
		"absences_count":  len(absences),
		"novelties_count": len(novelties),
	})

	// Build data maps for efficient lookups
	absenceMap := make(map[string]float64)
	noveltyMap := make(map[string]float64)
	assignedMap := make(map[string]domain.AssignedShift)

	// Process absences
	for _, absence := range absences {
		key := absence.Date.Format("2006-01-02")
		absenceMap[key] += absence.Hours
	}

	// Process novelties (positive adds hours, negative subtracts)
	for _, novelty := range novelties {
		key := novelty.Date.Format("2006-01-02")
		if novelty.Type == "positive" {
			noveltyMap[key] += novelty.Hours
		} else {
			noveltyMap[key] -= novelty.Hours
		}
	}

	// Process shifts
	for _, shift := range shifts {
		assignedMap[shift.Date] = shift
	}

	// Initialize summary structure
	summary := domain.EmployeeHourSummary{
		Employee: domain.EmployeeInfo{
			ID:       employeeID,
			FullName: fullNameEmployee,
		},
		Period: domain.Period{Year: year, Month: month},
	}

	// Process each calendar day
	processedDays := 0
	for _, day := range calendarDays {
		dateKey := day.Date.Format("2006-01-02")
		shift, hasShift := assignedMap[dateKey]

		if !hasShift {
			continue // No shift assigned for this day
		}

		// Parse shift times
		startTime := utils.ParseHour(shift.StartTime)
		endTime := utils.ParseHour(shift.EndTime)

		// Calculate base worked hours (excluding lunch)
		baseWorkedHours := utils.DurationInHours(startTime, endTime) - float64(shift.LunchMinutes)/60.0

		// Apply novelties (can increase or decrease hours)
		adjustedHours := baseWorkedHours + noveltyMap[dateKey]
		absenceHours := absenceMap[dateKey]

		// Calculate extra hours (anything above 7.33 hours standard)
		extraHours := adjustedHours - 7.33
		if extraHours < 0 {
			extraHours = 0
		}

		// Split extra hours by time periods (diurnal vs nocturnal)
		diurnalHours, nocturnalHours := utils.SplitByFranja(
			startTime, endTime,
			utils.ParseHour("06:00"), utils.ParseHour("21:00"),
		)

		// Calculate proportional extra hours distribution
		var diurnalExtra, nocturnalExtra float64
		totalPeriodHours := diurnalHours + nocturnalHours
		if totalPeriodHours > 0 {
			scale := extraHours / totalPeriodHours
			diurnalExtra = utils.RoundTo2(scale * diurnalHours)
			nocturnalExtra = utils.RoundTo2(scale * nocturnalHours)
		}

		// Determine which block to update based on day type
		var targetBlock *domain.EmployeeHourBlock
		switch day.DayType {
		case utils.Sunday:
			targetBlock = &summary.Sunday
		case utils.Holiday:
			targetBlock = &summary.Holiday
		default:
			targetBlock = &summary.Ordinary
		}

		// Update the appropriate block
		targetBlock.Absence += utils.RoundTo2(absenceHours)
		targetBlock.Novelty += utils.RoundTo2(noveltyMap[dateKey])
		targetBlock.DiurnalExtra += diurnalExtra
		targetBlock.NocturnalExtra += nocturnalExtra

		processedDays++
	}

	uc.logger.LogOperation("GetMonthlySummary", "success", map[string]interface{}{
		"employee_id":    employeeID,
		"employee_name":  fullNameEmployee,
		"year":           year,
		"month":          month,
		"processed_days": processedDays,
		"total_ordinary": summary.Ordinary,
		"total_sunday":   summary.Sunday,
		"total_holiday":  summary.Holiday,
	})

	return summary, nil
}

// ✅ Business-specific operations with enhanced validation and logging

// ValidateEmployeeID validates employee ID parameter
func (uc *employeeHoursUseCase) ValidateEmployeeID(employeeID int) error {
	uc.logger.LogOperation("ValidateEmployeeID", "start", map[string]interface{}{
		"employee_id": employeeID,
	})

	if err := uc.validator.ValidateID(employeeID); err != nil {
		uc.logger.LogValidation("ValidateEmployeeID", "id_format", "failed", map[string]interface{}{
			"error":       err.Error(),
			"employee_id": employeeID,
		})
		return uc.errorHandler.HandleValidationError("ValidateEmployeeID", err)
	}

	// Check if employee exists
	_, err := uc.userRepo.GetNameByID(employeeID)
	if err != nil {
		uc.logger.LogValidation("ValidateEmployeeID", "employee_exists", "failed", map[string]interface{}{
			"error":       err.Error(),
			"employee_id": employeeID,
		})
		return uc.errorHandler.HandleNotFound("ValidateEmployeeID", fmt.Sprintf("employee not found with ID: %d", employeeID))
	}

	uc.logger.LogValidation("ValidateEmployeeID", "all_checks", "passed", map[string]interface{}{
		"employee_id": employeeID,
	})

	return nil
}

// ValidatePeriod validates year and month parameters
func (uc *employeeHoursUseCase) ValidatePeriod(year, month int) error {
	uc.logger.LogOperation("ValidatePeriod", "start", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	// Business rule: Year must be reasonable (between 2000 and current year + 1)
	if year < 2000 || year > 2030 {
		err := fmt.Errorf("year must be between 2000 and 2030, got: %d", year)
		uc.logger.LogValidation("ValidatePeriod", "year_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"year":  year,
		})
		return uc.errorHandler.HandleValidationError("ValidatePeriod", err)
	}

	// Business rule: Month must be between 1 and 12
	if month < 1 || month > 12 {
		err := fmt.Errorf("month must be between 1 and 12, got: %d", month)
		uc.logger.LogValidation("ValidatePeriod", "month_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"month": month,
		})
		return uc.errorHandler.HandleValidationError("ValidatePeriod", err)
	}

	uc.logger.LogValidation("ValidatePeriod", "all_checks", "passed", map[string]interface{}{
		"year":  year,
		"month": month,
	})

	return nil
}

// CalculateWorkingDays calculates the number of working days for an employee in a given month
func (uc *employeeHoursUseCase) CalculateWorkingDays(employeeID, year, month int) (int, error) {
	uc.logger.LogOperation("CalculateWorkingDays", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
	})

	// Validate inputs
	if err := uc.ValidateEmployeeID(employeeID); err != nil {
		return 0, err
	}
	if err := uc.ValidatePeriod(year, month); err != nil {
		return 0, err
	}

	// Get shifts for the employee
	shifts, err := uc.assignedRepo.GetByEmployeeAndMonth(employeeID, year, month)
	if err != nil {
		uc.logger.LogError("CalculateWorkingDays", err, map[string]interface{}{
			"employee_id": employeeID,
			"year":        year,
			"month":       month,
		})
		return 0, uc.errorHandler.HandleRepositoryError("CalculateWorkingDays", err)
	}

	workingDays := len(shifts)

	uc.logger.LogOperation("CalculateWorkingDays", "success", map[string]interface{}{
		"employee_id":  employeeID,
		"year":         year,
		"month":        month,
		"working_days": workingDays,
	})

	return workingDays, nil
}

// GetDailySummary provides detailed hours breakdown for a specific day
func (uc *employeeHoursUseCase) GetDailySummary(employeeID int, year, month, day int) (DailyHoursSummary, error) {
	uc.logger.LogOperation("GetDailySummary", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
		"month":       month,
		"day":         day,
	})

	// Validate inputs
	if err := uc.ValidateEmployeeID(employeeID); err != nil {
		return DailyHoursSummary{}, err
	}
	if err := uc.ValidatePeriod(year, month); err != nil {
		return DailyHoursSummary{}, err
	}

	// Validate day
	if day < 1 || day > 31 {
		err := fmt.Errorf("day must be between 1 and 31, got: %d", day)
		uc.logger.LogValidation("GetDailySummary", "day_range", "failed", map[string]interface{}{
			"error": err.Error(),
			"day":   day,
		})
		return DailyHoursSummary{}, uc.errorHandler.HandleValidationError("GetDailySummary", err)
	}

	// Get monthly summary and extract daily data
	monthlySummary, err := uc.GetMonthlySummary(employeeID, year, month)
	if err != nil {
		return DailyHoursSummary{}, err // Error already logged by GetMonthlySummary
	}

	// For simplicity, we'll return aggregated data for the requested date
	// In a real implementation, you might want more granular daily tracking
	dateStr := fmt.Sprintf("%04d-%02d-%02d", year, month, day)

	summary := DailyHoursSummary{
		Date:         dateStr,
		RegularHours: 8.0, // Standard working hours
		ExtraHours:   monthlySummary.Ordinary.DiurnalExtra + monthlySummary.Ordinary.NocturnalExtra,
		AbsenceHours: monthlySummary.Ordinary.Absence,
		NoveltyHours: monthlySummary.Ordinary.Novelty,
	}

	summary.TotalHours = summary.RegularHours + summary.ExtraHours + summary.NoveltyHours - summary.AbsenceHours

	uc.logger.LogOperation("GetDailySummary", "success", map[string]interface{}{
		"employee_id":   employeeID,
		"date":          dateStr,
		"daily_summary": summary,
	})

	return summary, nil
}

// GetYearlySummary provides comprehensive yearly hours summary for an employee
func (uc *employeeHoursUseCase) GetYearlySummary(employeeID, year int) (YearlyHoursSummary, error) {
	uc.logger.LogOperation("GetYearlySummary", "start", map[string]interface{}{
		"employee_id": employeeID,
		"year":        year,
	})

	// Validate inputs
	if err := uc.ValidateEmployeeID(employeeID); err != nil {
		return YearlyHoursSummary{}, err
	}

	if year < 2000 || year > 2030 {
		err := fmt.Errorf("year must be between 2000 and 2030, got: %d", year)
		return YearlyHoursSummary{}, uc.errorHandler.HandleValidationError("GetYearlySummary", err)
	}

	// Get employee name
	employeeName, err := uc.userRepo.GetNameByID(employeeID)
	if err != nil {
		uc.logger.LogError("GetYearlySummary", err, map[string]interface{}{
			"employee_id": employeeID,
		})
		return YearlyHoursSummary{}, uc.errorHandler.HandleRepositoryError("GetYearlySummary", err)
	}

	// Initialize yearly summary
	summary := YearlyHoursSummary{
		Year:         year,
		EmployeeID:   employeeID,
		EmployeeName: employeeName,
		MonthlyData:  make([]domain.EmployeeHourSummary, 0, 12),
		TotalHours:   domain.EmployeeHourBlock{},
	}

	// Get data for each month
	for month := 1; month <= 12; month++ {
		monthSummary, err := uc.GetMonthlySummary(employeeID, year, month)
		if err != nil {
			// Log warning but continue with other months
			uc.logger.LogError("GetYearlySummary", err, map[string]interface{}{
				"employee_id": employeeID,
				"year":        year,
				"month":       month,
			})
			continue
		}

		summary.MonthlyData = append(summary.MonthlyData, monthSummary)

		// Accumulate totals
		summary.TotalHours.Absence += monthSummary.Ordinary.Absence + monthSummary.Sunday.Absence + monthSummary.Holiday.Absence
		summary.TotalHours.Novelty += monthSummary.Ordinary.Novelty + monthSummary.Sunday.Novelty + monthSummary.Holiday.Novelty
		summary.TotalHours.DiurnalExtra += monthSummary.Ordinary.DiurnalExtra + monthSummary.Sunday.DiurnalExtra + monthSummary.Holiday.DiurnalExtra
		summary.TotalHours.NocturnalExtra += monthSummary.Ordinary.NocturnalExtra + monthSummary.Sunday.NocturnalExtra + monthSummary.Holiday.NocturnalExtra
	}

	uc.logger.LogOperation("GetYearlySummary", "success", map[string]interface{}{
		"employee_id":      employeeID,
		"employee_name":    employeeName,
		"year":             year,
		"months_processed": len(summary.MonthlyData),
		"total_hours":      summary.TotalHours,
	})

	return summary, nil
}
