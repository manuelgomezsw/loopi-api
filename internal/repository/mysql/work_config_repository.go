package mysql

import (
	"loopi-api/internal/domain"
	"loopi-api/internal/repository"

	"gorm.io/gorm"
)

// workConfigRepository implements repository.WorkConfigRepository with improved maintainability
type workConfigRepository struct {
	*BaseRepository[domain.WorkConfig]
	errorHandler *ErrorHandler
}

// NewWorkConfigRepository creates a new work config repository with enhanced features
func NewWorkConfigRepository(db *gorm.DB) repository.WorkConfigRepository {
	return &workConfigRepository{
		BaseRepository: NewBaseRepository[domain.WorkConfig](db, "work_configs"),
		errorHandler:   NewErrorHandler("work_configs"),
	}
}

// GetActiveConfig retrieves the active work configuration with fallback to defaults
func (r *workConfigRepository) GetActiveConfig() domain.WorkConfig {
	var config domain.WorkConfig

	err := NewQueryBuilder(r.GetDB()).
		WhereActive().
		OrderBy("id", "DESC").
		Limit(1).
		GetDB().
		First(&config).Error

	if err != nil {
		// Log error but don't fail - return fallback configuration
		r.errorHandler.LogError("GetActiveConfig", err)

		// Return fallback configuration
		return r.getDefaultConfig()
	}

	return config
}

// getDefaultConfig returns the default work configuration
func (r *workConfigRepository) getDefaultConfig() domain.WorkConfig {
	return domain.WorkConfig{
		DiurnalStart: "06:00",
		DiurnalEnd:   "21:00",
		IsActive:     true,
	}
}

// Create creates a new work configuration with validation
func (r *workConfigRepository) Create(config *domain.WorkConfig) error {
	// Business validation before creation
	if err := r.validateWorkConfig(config); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}

	// If this config is active, deactivate all others first
	if config.IsActive {
		if err := r.deactivateAllConfigs(); err != nil {
			return r.errorHandler.HandleError("Create", err)
		}
	}

	if err := r.BaseRepository.Create(config); err != nil {
		return r.errorHandler.HandleError("Create", err)
	}
	return nil
}

// Update updates an existing work configuration
func (r *workConfigRepository) Update(config *domain.WorkConfig) error {
	// Business validation before update
	if err := r.validateWorkConfig(config); err != nil {
		return r.errorHandler.HandleError("Update", err, config.ID)
	}

	// If this config is being activated, deactivate all others first
	if config.IsActive {
		if err := r.deactivateOtherConfigs(int(config.ID)); err != nil {
			return r.errorHandler.HandleError("Update", err, config.ID)
		}
	}

	if err := r.BaseRepository.Update(config); err != nil {
		return r.errorHandler.HandleError("Update", err, config.ID)
	}
	return nil
}

// validateWorkConfig performs business validation
func (r *workConfigRepository) validateWorkConfig(config *domain.WorkConfig) error {
	if config.DiurnalStart == "" {
		return ErrInvalidInput
	}
	if config.DiurnalEnd == "" {
		return ErrInvalidInput
	}
	// Could add time format validation here
	return nil
}

// deactivateAllConfigs sets all configurations to inactive
func (r *workConfigRepository) deactivateAllConfigs() error {
	return r.GetDB().
		Model(&domain.WorkConfig{}).
		Where("is_active = ?", true).
		Update("is_active", false).Error
}

// deactivateOtherConfigs sets all other configurations to inactive except the specified ID
func (r *workConfigRepository) deactivateOtherConfigs(exceptID int) error {
	return r.GetDB().
		Model(&domain.WorkConfig{}).
		Where("is_active = ? AND id != ?", true, exceptID).
		Update("is_active", false).Error
}

// GetAllConfigs retrieves all work configurations
func (r *workConfigRepository) GetAllConfigs() ([]domain.WorkConfig, error) {
	configs, err := r.BaseRepository.GetAll()
	if err != nil {
		return nil, r.errorHandler.HandleError("GetAllConfigs", err)
	}
	return configs, nil
}

// GetConfigByID retrieves a work configuration by ID
func (r *workConfigRepository) GetConfigByID(id int) (*domain.WorkConfig, error) {
	config, err := r.BaseRepository.GetByID(id)
	if err != nil {
		if err == ErrNotFound {
			return nil, r.errorHandler.HandleNotFound("GetConfigByID", id)
		}
		return nil, r.errorHandler.HandleError("GetConfigByID", err, id)
	}
	return config, nil
}

// ActivateConfig activates a specific configuration and deactivates all others
func (r *workConfigRepository) ActivateConfig(id int) error {
	// Check if config exists
	exists, err := r.BaseRepository.Exists(id)
	if err != nil {
		return r.errorHandler.HandleError("ActivateConfig", err, id)
	}
	if !exists {
		return r.errorHandler.HandleNotFound("ActivateConfig", id)
	}

	// Use transaction to ensure atomicity
	return r.BaseRepository.Transaction(func(tx *gorm.DB) error {
		// Deactivate all configs
		if err := tx.Model(&domain.WorkConfig{}).
			Where("is_active = ?", true).
			Update("is_active", false).Error; err != nil {
			return err
		}

		// Activate the specified config
		return tx.Model(&domain.WorkConfig{}).
			Where("id = ?", id).
			Update("is_active", true).Error
	})
}
