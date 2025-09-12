package repository

import (
	"loopi-api/internal/domain"
	"time"
)

// NoveltyTypeSummary represents a summary of novelties by type
type NoveltyTypeSummary struct {
	Type       string  `json:"type"`
	Count      int     `json:"count"`
	TotalHours float64 `json:"total_hours"`
}

type NoveltyRepository interface {
	// Basic operations
	GetByEmployeeAndMonth(employeeID, year, month int) ([]domain.Novelty, error)
	Create(novelty *domain.Novelty) error

	// Enhanced business operations
	GetByEmployeeAndDateRange(employeeID int, from, to time.Time) ([]domain.Novelty, error)
	GetTotalHoursByEmployeeAndType(employeeID, year, month int, noveltyType string) (float64, error)
	GetNoveltyTypesSummary(employeeID, year, month int) ([]NoveltyTypeSummary, error)
}
