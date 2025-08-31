package repository

import (
	"gorm.io/gorm"
	"loopi-api/internal/domain"
)

type WorkConfigRepository interface {
	GetActiveConfig() domain.WorkConfig
}

type workConfigRepo struct {
	db *gorm.DB
}

func NewWorkConfigRepository(db *gorm.DB) WorkConfigRepository {
	return &workConfigRepo{db}
}

func (r *workConfigRepo) GetActiveConfig() domain.WorkConfig {
	var config domain.WorkConfig
	err := r.db.Where("is_active = ?", true).Order("id DESC").First(&config).Error
	if err != nil {
		// Fallback a configuración por defecto
		return domain.WorkConfig{
			DiurnalStart: "06:00",
			DiurnalEnd:   "21:00",
			IsActive:     true,
		}
	}
	return config
}
