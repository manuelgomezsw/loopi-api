package usecase

import (
	"fmt"
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"
	"loopi-api/internal/usecase/base"
	"strconv"
	"strings"
	"time"
)

type ShiftUseCase interface {
	// Standard CRUD operations
	GetAll() ([]domain.Shift, error)
	GetByID(id int) (*domain.Shift, error)
	GetByStore(storeID int) ([]domain.Shift, error)
	Create(shift domain.Shift) error
	Update(shift domain.Shift) error
	Delete(id int) error

	// Business-specific operations
	GetActiveShiftsByStore(storeID int) ([]domain.Shift, error)
	ValidateShiftData(shift *domain.Shift) error
	ValidateShiftTiming(shift *domain.Shift) error
	GetShiftStatistics(storeID int) (*repository.ShiftStatistics, error)
}

type shiftUseCase struct {
	repo         repository.ShiftRepository
	errorHandler *base.ErrorHandler
	validator    *base.Validator
	logger       *base.Logger
}

func NewShiftUseCase(repo repository.ShiftRepository) ShiftUseCase {
	return &shiftUseCase{
		repo:         repo,
		errorHandler: base.NewErrorHandler("Shift"),
		validator:    base.NewValidator(),
		logger:       base.NewLogger("Shift"),
	}
}

// ✅ Enhanced CRUD operations with logging, validation, and error handling

// GetAll retrieves all shifts with proper error handling and logging
func (uc *shiftUseCase) GetAll() ([]domain.Shift, error) {
	uc.logger.LogOperation("GetAll", "start", nil)

	shifts, err := uc.repo.ListAll()
	if err != nil {
		uc.logger.LogError("GetAll", err, nil)
		return nil, uc.errorHandler.HandleRepositoryError("GetAll", err)
	}

	if len(shifts) == 0 {
		uc.logger.LogOperation("GetAll", "no_shifts_found", nil)
		return nil, uc.errorHandler.HandleNotFound("GetAll", "No shifts found")
	}

	uc.logger.LogOperation("GetAll", "success", map[string]interface{}{
		"count": len(shifts),
	})

	return shifts, nil
}

// GetByID retrieves a shift by ID with validation and error handling
func (uc *shiftUseCase) GetByID(id int) (*domain.Shift, error) {
	uc.logger.LogOperation("GetByID", "start", map[string]interface{}{"id": id})

	// Validate ID
	if err := uc.validator.ValidateID(id); err != nil {
		uc.logger.LogError("GetByID", err, map[string]interface{}{"id": id})
		return nil, uc.errorHandler.HandleValidationError("GetByID", err)
	}

	shift, err := uc.repo.GetByID(id)
	if err != nil {
		uc.logger.LogError("GetByID", err, map[string]interface{}{"id": id})
		return nil, uc.errorHandler.HandleRepositoryError("GetByID", err)
	}

	uc.logger.LogOperation("GetByID", "success", map[string]interface{}{"id": id})
	return shift, nil
}

// GetByStore retrieves shifts by store ID with validation and error handling
func (uc *shiftUseCase) GetByStore(storeID int) ([]domain.Shift, error) {
	uc.logger.LogOperation("GetByStore", "start", map[string]interface{}{"store_id": storeID})

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("GetByStore", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleValidationError("GetByStore", err)
	}

	shifts, err := uc.repo.ListByStore(storeID)
	if err != nil {
		uc.logger.LogError("GetByStore", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleRepositoryError("GetByStore", err)
	}

	uc.logger.LogOperation("GetByStore", "success", map[string]interface{}{
		"store_id": storeID,
		"count":    len(shifts),
	})

	return shifts, nil
}

// Create creates a new shift with validation and business rules
func (uc *shiftUseCase) Create(shift domain.Shift) error {
	uc.logger.LogOperation("Create", "start", map[string]interface{}{
		"shift_name": shift.Name,
		"store_id":   shift.StoreID,
	})

	// Validate business rules
	if err := uc.ValidateShiftData(&shift); err != nil {
		return err
	}

	// Validate shift timing
	if err := uc.ValidateShiftTiming(&shift); err != nil {
		return err
	}

	// Set default values (business rule)
	shift.IsActive = true

	// Execute creation
	if err := uc.repo.Create(shift); err != nil {
		uc.logger.LogError("Create", err, map[string]interface{}{
			"shift_name": shift.Name,
			"store_id":   shift.StoreID,
		})
		return uc.errorHandler.HandleRepositoryError("Create", err)
	}

	uc.logger.LogOperation("Create", "success", map[string]interface{}{
		"shift_id":   shift.ID,
		"shift_name": shift.Name,
		"store_id":   shift.StoreID,
	})

	return nil
}

// Update modifies an existing shift with business validation
func (uc *shiftUseCase) Update(shift domain.Shift) error {
	uc.logger.LogOperation("Update", "start", map[string]interface{}{
		"id":       shift.ID,
		"name":     shift.Name,
		"store_id": shift.StoreID,
	})

	// Validate shift ID
	if shift.ID == 0 {
		err := fmt.Errorf("shift ID is required for update")
		uc.logger.LogValidation("Update", "id", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return uc.errorHandler.HandleValidationError("Update", err)
	}

	// Validate shift data
	if err := uc.ValidateShiftData(&shift); err != nil {
		uc.logger.LogValidation("Update", "data", "failed", map[string]interface{}{
			"error": err.Error(),
			"id":    shift.ID,
		})
		return err
	}

	// Validate shift timing
	if err := uc.ValidateShiftTiming(&shift); err != nil {
		uc.logger.LogValidation("Update", "timing", "failed", map[string]interface{}{
			"error": err.Error(),
			"id":    shift.ID,
		})
		return err
	}

	// Check if shift exists and is active
	existingShift, err := uc.repo.GetByID(int(shift.ID))
	if err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{"id": shift.ID})
		return uc.errorHandler.HandleRepositoryError("Update", err)
	}

	// Business rule: Cannot update inactive shifts
	if !existingShift.IsActive {
		err := fmt.Errorf("cannot update inactive shift with ID %d", shift.ID)
		uc.logger.LogValidation("Update", "inactive_shift", "failed", map[string]interface{}{
			"error": err.Error(),
			"id":    shift.ID,
		})
		return uc.errorHandler.HandleBusinessRuleViolation("Update", "inactive_shift", err.Error())
	}

	// Preserve timestamps and ensure IsActive is maintained
	shift.CreatedAt = existingShift.CreatedAt
	shift.UpdatedAt = time.Now()
	if shift.IsActive == false && existingShift.IsActive == true {
		// If trying to deactivate, use Delete method instead
		uc.logger.LogValidation("Update", "deactivation_attempt", "failed", map[string]interface{}{
			"message": "Use Delete method to deactivate shifts",
			"id":      shift.ID,
		})
		return uc.errorHandler.HandleBusinessRuleViolation("Update", "deactivation_attempt", "Use Delete method to deactivate shifts")
	}
	shift.IsActive = existingShift.IsActive

	// Perform update
	if err := uc.repo.Update(shift); err != nil {
		uc.logger.LogError("Update", err, map[string]interface{}{"id": shift.ID})
		return uc.errorHandler.HandleRepositoryError("Update", err)
	}

	uc.logger.LogOperation("Update", "success", map[string]interface{}{
		"id":         shift.ID,
		"shift_name": shift.Name,
		"store_id":   shift.StoreID,
	})

	return nil
}

// Delete removes a shift by ID with business validation
func (uc *shiftUseCase) Delete(id int) error {
	uc.logger.LogOperation("Delete", "start", map[string]interface{}{"id": id})

	// Validate ID
	if id <= 0 {
		err := fmt.Errorf("invalid shift ID: %d", id)
		uc.logger.LogValidation("Delete", "id", "failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return uc.errorHandler.HandleValidationError("Delete", err)
	}

	// Check if shift exists before deletion
	shift, err := uc.repo.GetByID(id)
	if err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{"id": id})
		return uc.errorHandler.HandleRepositoryError("Delete", err)
	}

	// Business rule: Cannot delete shift if it's not active (already deleted)
	if !shift.IsActive {
		err := fmt.Errorf("shift with ID %d is already inactive/deleted", id)
		uc.logger.LogValidation("Delete", "already_inactive", "failed", map[string]interface{}{
			"error": err.Error(),
			"id":    id,
		})
		return uc.errorHandler.HandleBusinessRuleViolation("Delete", "already_inactive", err.Error())
	}

	// Perform deletion
	if err := uc.repo.Delete(id); err != nil {
		uc.logger.LogError("Delete", err, map[string]interface{}{"id": id})
		return uc.errorHandler.HandleRepositoryError("Delete", err)
	}

	uc.logger.LogOperation("Delete", "success", map[string]interface{}{
		"id":         id,
		"shift_name": shift.Name,
		"store_id":   shift.StoreID,
	})

	return nil
}

// ✅ Business-specific operations with enhanced features

// GetActiveShiftsByStore retrieves only active shifts for a store with business rule filtering
func (uc *shiftUseCase) GetActiveShiftsByStore(storeID int) ([]domain.Shift, error) {
	// Start performance timer
	timer := uc.logger.StartTimer("GetActiveShiftsByStore", map[string]interface{}{"store_id": storeID})
	defer timer.Stop()

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("GetActiveShiftsByStore", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleValidationError("GetActiveShiftsByStore", err)
	}

	// Get all shifts for the store
	shifts, err := uc.repo.ListByStore(storeID)
	if err != nil {
		uc.logger.LogError("GetActiveShiftsByStore", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleRepositoryError("GetActiveShiftsByStore", err)
	}

	// Apply business rule: filter only active shifts
	activeShifts := make([]domain.Shift, 0)
	for _, shift := range shifts {
		if shift.IsActive {
			activeShifts = append(activeShifts, shift)
		}
	}

	uc.logger.LogBusinessRule("GetActiveShiftsByStore", "filter_active_only", "applied", map[string]interface{}{
		"store_id":      storeID,
		"total_shifts":  len(shifts),
		"active_shifts": len(activeShifts),
	})

	uc.logger.LogOperation("GetActiveShiftsByStore", "success", map[string]interface{}{
		"store_id":     storeID,
		"active_count": len(activeShifts),
	})

	return activeShifts, nil
}

// ValidateShiftData validates shift data according to business rules
func (uc *shiftUseCase) ValidateShiftData(shift *domain.Shift) error {
	// Basic entity validation
	if err := uc.validator.ValidateEntity(shift); err != nil {
		uc.logger.LogValidation("ValidateShiftData", "entity", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule validations
	if err := uc.validator.ValidateString(shift.Name, "name", "required", "min:3", "max:50"); err != nil {
		uc.logger.LogValidation("ValidateShiftData", "name", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate store ID
	if err := uc.validator.ValidateNumber(shift.StoreID, "store_id", "positive"); err != nil {
		uc.logger.LogValidation("ValidateShiftData", "store_id", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Validate lunch minutes
	if err := uc.validator.ValidateNumber(shift.LunchMinutes, "lunch_minutes", "non_negative"); err != nil {
		uc.logger.LogValidation("ValidateShiftData", "lunch_minutes", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	uc.logger.LogValidation("ValidateShiftData", "all_fields", "passed", nil)
	return nil
}

// ValidateShiftTiming validates shift timing according to business rules
func (uc *shiftUseCase) ValidateShiftTiming(shift *domain.Shift) error {
	uc.logger.LogOperation("ValidateShiftTiming", "start", map[string]interface{}{
		"start_time": shift.StartTime,
		"end_time":   shift.EndTime,
	})

	// Validate time format (HH:MM)
	if !isValidTimeFormat(shift.StartTime) {
		err := fmt.Errorf("invalid start_time format: %s. Expected format: HH:MM", shift.StartTime)
		uc.logger.LogValidation("ValidateShiftTiming", "start_time_format", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	if !isValidTimeFormat(shift.EndTime) {
		err := fmt.Errorf("invalid end_time format: %s. Expected format: HH:MM", shift.EndTime)
		uc.logger.LogValidation("ValidateShiftTiming", "end_time_format", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Parse times for validation
	startTime, err := time.Parse("15:04", shift.StartTime)
	if err != nil {
		uc.logger.LogValidation("ValidateShiftTiming", "start_time_parse", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("invalid start_time: %s", shift.StartTime)
	}

	endTime, err := time.Parse("15:04", shift.EndTime)
	if err != nil {
		uc.logger.LogValidation("ValidateShiftTiming", "end_time_parse", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return fmt.Errorf("invalid end_time: %s", shift.EndTime)
	}

	// Business rule: end time must be after start time
	if !endTime.After(startTime) {
		err := fmt.Errorf("end_time (%s) must be after start_time (%s)", shift.EndTime, shift.StartTime)
		uc.logger.LogValidation("ValidateShiftTiming", "time_sequence", "failed", map[string]interface{}{
			"error": err.Error(),
		})
		return err
	}

	// Business rule: validate minimum shift duration (at least 1 hour)
	duration := endTime.Sub(startTime)
	if duration < time.Hour {
		err := fmt.Errorf("shift duration must be at least 1 hour, got: %v", duration)
		uc.logger.LogValidation("ValidateShiftTiming", "minimum_duration", "failed", map[string]interface{}{
			"error":    err.Error(),
			"duration": duration.String(),
		})
		return err
	}

	// Business rule: validate maximum shift duration (no more than 12 hours)
	if duration > 12*time.Hour {
		err := fmt.Errorf("shift duration must not exceed 12 hours, got: %v", duration)
		uc.logger.LogValidation("ValidateShiftTiming", "maximum_duration", "failed", map[string]interface{}{
			"error":    err.Error(),
			"duration": duration.String(),
		})
		return err
	}

	uc.logger.LogValidation("ValidateShiftTiming", "all_timing_rules", "passed", map[string]interface{}{
		"duration": duration.String(),
	})

	return nil
}

// GetShiftStatistics retrieves comprehensive shift statistics for a store
func (uc *shiftUseCase) GetShiftStatistics(storeID int) (*repository.ShiftStatistics, error) {
	uc.logger.LogOperation("GetShiftStatistics", "start", map[string]interface{}{"store_id": storeID})

	// Validate store ID
	if err := uc.validator.ValidateID(storeID); err != nil {
		uc.logger.LogError("GetShiftStatistics", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleValidationError("GetShiftStatistics", err)
	}

	// Use repository method for statistics (delegating to repository for better performance)
	stats, err := uc.repo.GetShiftStatistics(storeID)
	if err != nil {
		uc.logger.LogError("GetShiftStatistics", err, map[string]interface{}{"store_id": storeID})
		return nil, uc.errorHandler.HandleRepositoryError("GetShiftStatistics", err)
	}

	uc.logger.LogOperation("GetShiftStatistics", "success", map[string]interface{}{
		"store_id":      storeID,
		"total_shifts":  stats.TotalShifts,
		"active_shifts": stats.ActiveShifts,
	})

	return stats, nil
}

// Helper functions for validation

// isValidTimeFormat validates time format (HH:MM)
func isValidTimeFormat(timeStr string) bool {
	parts := strings.Split(timeStr, ":")
	if len(parts) != 2 {
		return false
	}

	hour, err := strconv.Atoi(parts[0])
	if err != nil || hour < 0 || hour > 23 {
		return false
	}

	minute, err := strconv.Atoi(parts[1])
	if err != nil || minute < 0 || minute > 59 {
		return false
	}

	return true
}
